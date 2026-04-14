package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type CreateConferenceClient interface {
	CreateConference(ctx context.Context, input client.CreateConferenceRequest) (*client.ConferenceResponse, error)
}

func newCreateCmd(createClient CreateConferenceClient) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new conference",
		Long:  "Launches an interactive flow to create a conference as an organizer.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, err := promptCreateConferenceInput(cmd)
			if err != nil {
				return err
			}

			created, err := createClient.CreateConference(cmd.Context(), input)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrNotAuthenticated):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "You are not logged in. Run 'unconf login' to authenticate.")
				case errors.Is(err, client.ErrSessionExpired):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Your session has expired. Please run 'unconf login' to re-authenticate.")
				case errors.Is(err, client.ErrOrganizerForbidden):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Organizer permissions required to create conferences.")
				case errors.Is(err, client.ErrConferenceExists):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Conference slug already exists. Choose a different slug.")
				default:
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Failed to create conference. Please try again.")
				}
				return fmt.Errorf("failed to create conference: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Conference created successfully: %s\n", created.Slug)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Share this slug with attendees: %s\n", created.Slug)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Next step: run 'unconf checkout %s'\n", created.Slug)
			return nil
		},
	}
}

func promptCreateConferenceInput(cmd *cobra.Command) (client.CreateConferenceRequest, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "Conference setup")
	_, _ = fmt.Fprintln(out, "--------------")

	name, err := promptField(reader, out, "Name", "")
	if err != nil || strings.TrimSpace(name) == "" {
		return client.CreateConferenceRequest{}, fmt.Errorf("invalid conference name")
	}

	slugInput, err := promptField(reader, out, "Slug", "")
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("failed to read slug: %w", err)
	}
	slug := normalizeConferenceSlug(slugInput)
	if !conferenceSlugPattern.MatchString(slug) {
		return client.CreateConferenceRequest{}, fmt.Errorf("invalid slug format: use lowercase letters, numbers, and hyphens")
	}

	description, err := promptField(reader, out, "Description", "")
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("failed to read description: %w", err)
	}

	location, err := promptField(reader, out, "Location", "")
	if err != nil || strings.TrimSpace(location) == "" {
		return client.CreateConferenceRequest{}, fmt.Errorf("invalid location")
	}

	startDateRaw, err := promptField(reader, out, "Start date (YYYY-MM-DD)", "")
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("failed to read start date: %w", err)
	}
	startDate, err := parseConferenceDate(startDateRaw)
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("invalid start date")
	}

	endDateRaw, err := promptField(reader, out, "End date (YYYY-MM-DD)", "")
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("failed to read end date: %w", err)
	}
	endDate, err := parseConferenceDate(endDateRaw)
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("invalid end date")
	}
	if endDate.Before(startDate) {
		return client.CreateConferenceRequest{}, fmt.Errorf("end date cannot be before start date")
	}

	capacityRaw, err := promptField(reader, out, "Capacity", "")
	if err != nil {
		return client.CreateConferenceRequest{}, fmt.Errorf("failed to read capacity: %w", err)
	}
	capacity, err := parseConferenceCapacity(capacityRaw)
	if err != nil {
		return client.CreateConferenceRequest{}, err
	}

	return client.CreateConferenceRequest{
		Name:        strings.TrimSpace(name),
		Slug:        slug,
		Description: strings.TrimSpace(description),
		Location:    strings.TrimSpace(location),
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.Format("2006-01-02"),
		Capacity:    capacity,
	}, nil
}

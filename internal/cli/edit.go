package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type EditConferenceClient interface {
	GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error)
	UpdateConference(ctx context.Context, slug string, input client.UpdateConferenceRequest) (*client.ConferenceResponse, error)
}

type EditContextStore interface {
	GetActiveConference() (string, error)
}

func newEditCmd(editClient EditConferenceClient, ctxStore EditContextStore) *cobra.Command {
	return &cobra.Command{
		Use:   "edit [conference-slug]",
		Short: "Edit conference details",
		Long:  "Updates conference details for a conference you organize. If no slug is provided, uses active context.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveEditSlug(cmd, ctxStore, args)
			if err != nil || slug == "" {
				return err
			}

			current, err := editClient.GetConference(cmd.Context(), slug)
			if err != nil {
				if errors.Is(err, client.ErrConferenceNotFound) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Conference %q not found.\n", slug)
				}
				return fmt.Errorf("failed to load conference: %w", err)
			}

			input, err := promptEditConferenceInput(cmd, current)
			if err != nil {
				return err
			}

			updated, err := editClient.UpdateConference(cmd.Context(), slug, input)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrNotAuthenticated):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "You are not logged in. Run 'unconf login' to authenticate.")
				case errors.Is(err, client.ErrSessionExpired):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Your session has expired. Please run 'unconf login' to re-authenticate.")
				case errors.Is(err, client.ErrOrganizerForbidden):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Organizer permissions required to edit this conference.")
				case errors.Is(err, client.ErrConferenceNotFound):
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Conference not found.")
				default:
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Failed to update conference. Please try again.")
				}
				return fmt.Errorf("failed to update conference: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Conference updated successfully: %s\n", updated.Slug)
			return nil
		},
	}
}

func resolveEditSlug(cmd *cobra.Command, ctxStore EditContextStore, args []string) (string, error) {
	if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}
	if strings.TrimSpace(slug) == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Usage: unconf edit <slug> or set context with 'unconf checkout <slug>'")
		return "", nil
	}
	return strings.TrimSpace(slug), nil
}

func promptEditConferenceInput(cmd *cobra.Command, current *client.ConferenceResponse) (client.UpdateConferenceRequest, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(out, "Editing conference %s\n", current.Slug)

	name, err := promptField(reader, out, "Name", current.Name)
	if err != nil || strings.TrimSpace(name) == "" {
		return client.UpdateConferenceRequest{}, fmt.Errorf("invalid conference name")
	}
	description, err := promptField(reader, out, "Description", current.Description)
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("failed to read description: %w", err)
	}
	location, err := promptField(reader, out, "Location", current.Location)
	if err != nil || strings.TrimSpace(location) == "" {
		return client.UpdateConferenceRequest{}, fmt.Errorf("invalid location")
	}
	startRaw, err := promptField(reader, out, "Start date (YYYY-MM-DD)", current.StartDate)
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("failed to read start date: %w", err)
	}
	startDate, err := parseConferenceDate(startRaw)
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("invalid start date")
	}
	endRaw, err := promptField(reader, out, "End date (YYYY-MM-DD)", current.EndDate)
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("failed to read end date: %w", err)
	}
	endDate, err := parseConferenceDate(endRaw)
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("invalid end date")
	}
	if endDate.Before(startDate) {
		return client.UpdateConferenceRequest{}, fmt.Errorf("end date cannot be before start date")
	}
	capacityRaw, err := promptField(reader, out, "Capacity", strconv.Itoa(current.Capacity))
	if err != nil {
		return client.UpdateConferenceRequest{}, fmt.Errorf("failed to read capacity: %w", err)
	}
	capacity, err := parseConferenceCapacity(capacityRaw)
	if err != nil {
		return client.UpdateConferenceRequest{}, err
	}

	return client.UpdateConferenceRequest{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Location:    strings.TrimSpace(location),
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.Format("2006-01-02"),
		Capacity:    capacity,
	}, nil
}

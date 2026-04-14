package cli

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type roomInput struct {
	RoomNumber    string
	RoomType      string
	PricePerNight float64
	Capacity      int
}

func newRoomsAddCmd(roomsClient RoomsClient, ctxStore RoomsContextStore) *cobra.Command {
	var slug string
	var number string
	var roomType string
	var price float64
	var capacity int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a room to the active conference",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedSlug, err := resolveRoomManagementSlug(cmd, ctxStore, slug)
			if err != nil || resolvedSlug == "" {
				return err
			}

			input, err := collectRoomInput(cmd, roomInput{
				RoomNumber:    number,
				RoomType:      roomType,
				PricePerNight: price,
				Capacity:      capacity,
			}, false)
			if err != nil {
				return err
			}

			created, err := roomsClient.CreateRoom(cmd.Context(), resolvedSlug, toManageRoomRequest(input))
			if err != nil {
				handleRoomMutationError(cmd, "create", err)
				return fmt.Errorf("failed to create room: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Room %s created successfully in conference %s\n", created.RoomNumber, resolvedSlug)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "conference", "", "Conference slug (defaults to active context)")
	cmd.Flags().StringVar(&number, "number", "", "Room number")
	cmd.Flags().StringVar(&roomType, "type", "", "Room type: single|double|triple")
	cmd.Flags().Float64Var(&price, "price", 0, "Price per night")
	cmd.Flags().IntVar(&capacity, "capacity", 0, "Room capacity")

	return cmd
}

func newRoomsEditCmd(roomsClient RoomsClient, ctxStore RoomsContextStore) *cobra.Command {
	var slug string
	var number string
	var roomType string
	var price float64
	var capacity int

	cmd := &cobra.Command{
		Use:   "edit <number>",
		Short: "Edit room details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedSlug, err := resolveRoomManagementSlug(cmd, ctxStore, slug)
			if err != nil || resolvedSlug == "" {
				return err
			}

			targetNumber := strings.TrimSpace(args[0])
			if targetNumber == "" {
				return fmt.Errorf("room number is required")
			}

			defaults := roomInput{
				RoomNumber:    targetNumber,
				RoomType:      roomType,
				PricePerNight: price,
				Capacity:      capacity,
			}
			input, err := collectRoomInput(cmd, defaults, true)
			if err != nil {
				return err
			}

			updated, err := roomsClient.UpdateRoom(cmd.Context(), resolvedSlug, targetNumber, toManageRoomRequest(input))
			if err != nil {
				handleRoomMutationError(cmd, "update", err)
				return fmt.Errorf("failed to update room: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Room %s updated successfully in conference %s\n", updated.RoomNumber, resolvedSlug)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "conference", "", "Conference slug (defaults to active context)")
	cmd.Flags().StringVar(&number, "number", "", "Updated room number")
	cmd.Flags().StringVar(&roomType, "type", "", "Updated room type: single|double|triple")
	cmd.Flags().Float64Var(&price, "price", 0, "Updated price per night")
	cmd.Flags().IntVar(&capacity, "capacity", 0, "Updated room capacity")

	return cmd
}

func newRoomsRemoveCmd(roomsClient RoomsClient, ctxStore RoomsContextStore) *cobra.Command {
	var slug string

	cmd := &cobra.Command{
		Use:   "remove <number>",
		Short: "Remove a room from the active conference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedSlug, err := resolveRoomManagementSlug(cmd, ctxStore, slug)
			if err != nil || resolvedSlug == "" {
				return err
			}

			roomNumber := strings.TrimSpace(args[0])
			if roomNumber == "" {
				return fmt.Errorf("room number is required")
			}

			if err := roomsClient.DeleteRoom(cmd.Context(), resolvedSlug, roomNumber); err != nil {
				handleRoomMutationError(cmd, "remove", err)
				return fmt.Errorf("failed to remove room: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Room %s removed successfully from conference %s\n", roomNumber, resolvedSlug)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "conference", "", "Conference slug (defaults to active context)")
	return cmd
}

func newRoomsImportCmd(roomsClient RoomsClient, ctxStore RoomsContextStore) *cobra.Command {
	var slug string
	cmd := &cobra.Command{
		Use:   "import <csv>",
		Short: "Bulk import rooms from CSV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedSlug, err := resolveRoomManagementSlug(cmd, ctxStore, slug)
			if err != nil || resolvedSlug == "" {
				return err
			}

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open CSV file: %w", err)
			}
			defer func() { _ = file.Close() }()

			parsedRows, err := parseRoomImportCSV(file)
			if err != nil {
				return err
			}

			success := 0
			failures := 0
			for idx, row := range parsedRows {
				if _, createErr := roomsClient.CreateRoom(cmd.Context(), resolvedSlug, toManageRoomRequest(row)); createErr != nil {
					failures++
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Row %d failed: %v\n", idx+1, createErr)
					continue
				}
				success++
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Import summary: %d succeeded, %d failed\n", success, failures)
			if failures > 0 {
				return fmt.Errorf("room import completed with failures")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "conference", "", "Conference slug (defaults to active context)")
	return cmd
}

func resolveRoomManagementSlug(cmd *cobra.Command, ctxStore RoomsContextStore, explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit), nil
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}
	if strings.TrimSpace(slug) == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Use --conference <slug> or set context with 'unconf checkout <slug>'")
		return "", nil
	}

	return strings.TrimSpace(slug), nil
}

func collectRoomInput(cmd *cobra.Command, defaults roomInput, includeDefaults bool) (roomInput, error) {
	if defaults.RoomNumber != "" && defaults.RoomType != "" && defaults.PricePerNight > 0 && defaults.Capacity > 0 {
		return validateRoomInput(defaults)
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	roomNumberDefault := ""
	roomTypeDefault := ""
	priceDefault := ""
	capacityDefault := ""
	if includeDefaults {
		roomNumberDefault = defaults.RoomNumber
		roomTypeDefault = defaults.RoomType
		if defaults.PricePerNight > 0 {
			priceDefault = fmt.Sprintf("%.2f", defaults.PricePerNight)
		}
		if defaults.Capacity > 0 {
			capacityDefault = strconv.Itoa(defaults.Capacity)
		}
	}

	roomNumber, err := promptField(reader, out, "Room number", roomNumberDefault)
	if err != nil {
		return roomInput{}, fmt.Errorf("failed to read room number: %w", err)
	}
	roomType, err := promptField(reader, out, "Room type (single|double|triple)", roomTypeDefault)
	if err != nil {
		return roomInput{}, fmt.Errorf("failed to read room type: %w", err)
	}
	priceRaw, err := promptField(reader, out, "Price per night", priceDefault)
	if err != nil {
		return roomInput{}, fmt.Errorf("failed to read room price: %w", err)
	}
	capacityRaw, err := promptField(reader, out, "Capacity", capacityDefault)
	if err != nil {
		return roomInput{}, fmt.Errorf("failed to read room capacity: %w", err)
	}

	priceVal, err := strconv.ParseFloat(strings.TrimSpace(priceRaw), 64)
	if err != nil {
		return roomInput{}, fmt.Errorf("price must be a positive number")
	}
	capacityVal, err := strconv.Atoi(strings.TrimSpace(capacityRaw))
	if err != nil {
		return roomInput{}, fmt.Errorf("capacity must be a positive integer")
	}

	return validateRoomInput(roomInput{
		RoomNumber:    roomNumber,
		RoomType:      roomType,
		PricePerNight: priceVal,
		Capacity:      capacityVal,
	})
}

func validateRoomInput(input roomInput) (roomInput, error) {
	normalized := roomInput{
		RoomNumber:    strings.TrimSpace(input.RoomNumber),
		RoomType:      strings.ToLower(strings.TrimSpace(input.RoomType)),
		PricePerNight: input.PricePerNight,
		Capacity:      input.Capacity,
	}

	if normalized.RoomNumber == "" {
		return roomInput{}, fmt.Errorf("room number is required")
	}
	if normalized.RoomType != "single" && normalized.RoomType != "double" && normalized.RoomType != "triple" {
		return roomInput{}, fmt.Errorf("room type must be single, double, or triple")
	}
	if normalized.PricePerNight <= 0 {
		return roomInput{}, fmt.Errorf("price must be a positive number")
	}
	if normalized.Capacity <= 0 {
		return roomInput{}, fmt.Errorf("capacity must be a positive integer")
	}

	return normalized, nil
}

func parseRoomImportCSV(reader io.Reader) ([]roomInput, error) {
	csvReader := csv.NewReader(reader)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV is empty")
	}

	startRow := 0
	if looksLikeRoomHeader(rows[0]) {
		startRow = 1
	}

	parsed := make([]roomInput, 0, len(rows)-startRow)
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) != 4 {
			return nil, fmt.Errorf("row %d must have 4 columns: room_number,room_type,price_per_night,capacity", i+1)
		}

		priceVal, priceErr := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		if priceErr != nil {
			return nil, fmt.Errorf("row %d has invalid price_per_night", i+1)
		}
		capacityVal, capacityErr := strconv.Atoi(strings.TrimSpace(row[3]))
		if capacityErr != nil {
			return nil, fmt.Errorf("row %d has invalid capacity", i+1)
		}

		normalized, validationErr := validateRoomInput(roomInput{
			RoomNumber:    row[0],
			RoomType:      row[1],
			PricePerNight: priceVal,
			Capacity:      capacityVal,
		})
		if validationErr != nil {
			return nil, fmt.Errorf("row %d invalid: %w", i+1, validationErr)
		}
		parsed = append(parsed, normalized)
	}

	return parsed, nil
}

func looksLikeRoomHeader(row []string) bool {
	if len(row) != 4 {
		return false
	}
	header := []string{
		"room_number",
		"room_type",
		"price_per_night",
		"capacity",
	}
	for i := range row {
		if strings.ToLower(strings.TrimSpace(row[i])) != header[i] {
			return false
		}
	}

	return true
}

func toManageRoomRequest(input roomInput) client.ManageRoomRequest {
	return client.ManageRoomRequest{
		RoomNumber:    input.RoomNumber,
		RoomType:      input.RoomType,
		PricePerNight: input.PricePerNight,
		Capacity:      input.Capacity,
	}
}

func handleRoomMutationError(cmd *cobra.Command, action string, err error) {
	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "You are not logged in. Run 'unconf login' to authenticate.")
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Your session has expired. Please run 'unconf login' to re-authenticate.")
	case errors.Is(err, client.ErrOrganizerForbidden):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Organizer permissions required for room management.")
	case errors.Is(err, client.ErrConferenceNotFound):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Conference not found.")
	case errors.Is(err, client.ErrRoomNotFound):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Room not found.")
	case errors.Is(err, client.ErrRoomExists):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Room number already exists in this conference.")
	case errors.Is(err, client.ErrRoomHasBookings):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cannot remove room: bookings already exist for this room.")
	default:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Failed to %s room. Please try again.\n", action)
	}
}

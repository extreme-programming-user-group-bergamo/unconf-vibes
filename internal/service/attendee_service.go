package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

// AttendeeRoomResponse represents room details shown in attendee listings.
type AttendeeRoomResponse struct {
	RoomNumber string `json:"room_number"`
	RoomType   string `json:"room_type"`
}

// AttendeeResponse represents a public attendee in conference listings.
type AttendeeResponse struct {
	DisplayName    string                `json:"display_name"`
	GitHubUsername string                `json:"github_username,omitempty"`
	Room           *AttendeeRoomResponse `json:"room,omitempty"`
}

// AttendeeListResponse is the attendee endpoint payload for a conference.
type AttendeeListResponse struct {
	Attendees             []AttendeeResponse `json:"attendees"`
	PrivateAttendeesCount int                `json:"private_attendees_count"`
}

// OrganizerDashboardFilters defines optional attendee filtering for organizer dashboard queries.
type OrganizerDashboardFilters struct {
	RoomType           string
	BookingStatus      string
	HasSpecialRequests bool
	Search             string
}

// OrganizerDashboardRoomFillRateResponse represents occupancy summary by room type.
type OrganizerDashboardRoomFillRateResponse struct {
	RoomType      string  `json:"room_type"`
	Registrations int     `json:"registrations"`
	Capacity      int     `json:"capacity"`
	FillRatePct   float64 `json:"fill_rate_pct"`
}

// OrganizerDashboardAttendeeResponse represents an attendee row in the organizer dashboard.
type OrganizerDashboardAttendeeResponse struct {
	Name                     string `json:"name"`
	Email                    string `json:"email"`
	RoomNumber               string `json:"room_number,omitempty"`
	RoomType                 string `json:"room_type,omitempty"`
	BookingStatus            string `json:"booking_status"`
	DietaryAccessibilityNote string `json:"dietary_accessibility_note,omitempty"`
	PrivacySetting           string `json:"privacy_setting"`
}

// OrganizerDashboardResponse is the dashboard payload for organizers.
type OrganizerDashboardResponse struct {
	ConferenceSlug      string                                   `json:"conference_slug"`
	TotalRegistrations  int                                      `json:"total_registrations"`
	Capacity            int                                      `json:"capacity"`
	CapacityUsagePct    float64                                  `json:"capacity_usage_pct"`
	RoomFillRates       []OrganizerDashboardRoomFillRateResponse `json:"room_fill_rates"`
	Attendees           []OrganizerDashboardAttendeeResponse     `json:"attendees"`
	AppliedRoomType     string                                   `json:"applied_room_type,omitempty"`
	AppliedBookingState string                                   `json:"applied_booking_status,omitempty"`
	AppliedSearch       string                                   `json:"applied_search,omitempty"`
}

// ExportBookingRow represents one booking row in organizer CSV export.
type ExportBookingRow struct {
	Name  string
	Email string
	Room  string
	Dates string
	Notes string
}

// AttendeeService handles attendee listing logic.
type AttendeeService struct {
	confRepo      repository.ConferenceRepository
	organizerRepo repository.ConferenceOrganizerRepository
	bookingRepo   repository.BookingRepository
	roomRepo      repository.RoomRepository
	userRepo      repository.UserRepository
}

// NewAttendeeService creates a new AttendeeService.
func NewAttendeeService(
	confRepo repository.ConferenceRepository,
	organizerRepo repository.ConferenceOrganizerRepository,
	bookingRepo repository.BookingRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
) *AttendeeService {
	return &AttendeeService{
		confRepo:      confRepo,
		organizerRepo: organizerRepo,
		bookingRepo:   bookingRepo,
		roomRepo:      roomRepo,
		userRepo:      userRepo,
	}
}

// ListAttendees returns public attendees for a conference and an aggregate count of private attendees.
func (s *AttendeeService) ListAttendees(ctx context.Context, slug string, requesterUserID int64) (*AttendeeListResponse, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to list attendees: %w", ErrConferenceNotFound)
		}

		return nil, fmt.Errorf("failed to list attendees: %w", err)
	}

	bookings, err := s.bookingRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attendees: %w", err)
	}

	rooms, err := s.roomRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attendees: %w", err)
	}

	roomsByID := make(map[int64]*models.Room, len(rooms))
	for i := range rooms {
		roomsByID[rooms[i].ID] = rooms[i]
	}

	result := &AttendeeListResponse{
		Attendees: make([]AttendeeResponse, 0, len(bookings)),
	}

	includePrivate := false
	if requesterUserID > 0 && s.organizerRepo != nil {
		isOrganizer, organizerErr := s.organizerRepo.IsOrganizer(ctx, conf.ID, requesterUserID)
		if organizerErr != nil {
			return nil, fmt.Errorf("failed to list attendees: %w", organizerErr)
		}
		includePrivate = isOrganizer
	}

	for i := range bookings {
		booking := bookings[i]
		if !includePrivate && strings.EqualFold(strings.TrimSpace(booking.PrivacySetting), "private") {
			result.PrivateAttendeesCount++
			continue
		}

		user, userErr := s.userRepo.GetByID(ctx, booking.UserID)
		if userErr != nil {
			slog.Error("failed to fetch attendee user", "error", userErr, "conference_id", conf.ID, "user_id", booking.UserID)
			continue
		}

		if !includePrivate && strings.EqualFold(strings.TrimSpace(user.PrivacySetting), "private") {
			result.PrivateAttendeesCount++
			continue
		}

		attendee := AttendeeResponse{
			DisplayName:    displayNameOrFallback(user.DisplayName),
			GitHubUsername: strings.TrimSpace(user.GitHubID),
		}

		if room, ok := roomsByID[booking.RoomID]; ok {
			attendee.Room = &AttendeeRoomResponse{
				RoomNumber: room.RoomNumber,
				RoomType:   room.RoomType,
			}
		}

		result.Attendees = append(result.Attendees, attendee)
	}

	sort.SliceStable(result.Attendees, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(result.Attendees[i].DisplayName))
		right := strings.ToLower(strings.TrimSpace(result.Attendees[j].DisplayName))
		if left == right {
			return strings.ToLower(strings.TrimSpace(result.Attendees[i].GitHubUsername)) < strings.ToLower(strings.TrimSpace(result.Attendees[j].GitHubUsername))
		}

		return left < right
	})

	return result, nil
}

// GetOrganizerDashboard returns organizer metrics and attendee logistics rows for dashboard rendering.
func (s *AttendeeService) GetOrganizerDashboard(
	ctx context.Context,
	slug string,
	requesterUserID int64,
	filters OrganizerDashboardFilters,
) (*OrganizerDashboardResponse, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to get organizer dashboard: %w", ErrConferenceNotFound)
		}

		return nil, fmt.Errorf("failed to get organizer dashboard: %w", err)
	}

	if s.organizerRepo == nil {
		return nil, fmt.Errorf("failed to get organizer dashboard: %w", ErrOrganizerForbidden)
	}

	isOrganizer, err := s.organizerRepo.IsOrganizer(ctx, conf.ID, requesterUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer dashboard: %w", err)
	}
	if !isOrganizer {
		return nil, fmt.Errorf("failed to get organizer dashboard: %w", ErrOrganizerForbidden)
	}

	bookings, err := s.bookingRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer dashboard: %w", err)
	}

	rooms, err := s.roomRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizer dashboard: %w", err)
	}

	roomsByID := make(map[int64]*models.Room, len(rooms))
	roomTypeCapacity := map[string]int{}
	for i := range rooms {
		room := rooms[i]
		roomsByID[room.ID] = room
		roomType := strings.ToLower(strings.TrimSpace(room.RoomType))
		roomTypeCapacity[roomType] += room.Capacity
	}

	roomTypeRegistrations := map[string]int{}
	rows := make([]OrganizerDashboardAttendeeResponse, 0, len(bookings))
	for i := range bookings {
		booking := bookings[i]
		room := roomsByID[booking.RoomID]
		roomType := ""
		roomNumber := ""
		if room != nil {
			roomType = strings.ToLower(strings.TrimSpace(room.RoomType))
			roomNumber = strings.TrimSpace(room.RoomNumber)
			roomTypeRegistrations[roomType]++
		}

		user, userErr := s.userRepo.GetByID(ctx, booking.UserID)
		if userErr != nil {
			slog.Error("failed to fetch attendee user for organizer dashboard", "error", userErr, "conference_id", conf.ID, "user_id", booking.UserID)
			continue
		}

		privacySetting := strings.TrimSpace(booking.PrivacySetting)
		if privacySetting == "" {
			privacySetting = strings.TrimSpace(user.PrivacySetting)
		}

		row := OrganizerDashboardAttendeeResponse{
			Name:                     displayNameOrFallback(user.DisplayName),
			Email:                    strings.TrimSpace(user.Email),
			RoomNumber:               roomNumber,
			RoomType:                 roomType,
			BookingStatus:            strings.ToLower(strings.TrimSpace(string(booking.Status))),
			DietaryAccessibilityNote: strings.TrimSpace(booking.Notes),
			PrivacySetting:           strings.ToLower(privacySetting),
		}
		if organizerDashboardRowMatches(row, filters) {
			rows = append(rows, row)
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(rows[i].Name))
		right := strings.ToLower(strings.TrimSpace(rows[j].Name))
		if left == right {
			return strings.ToLower(strings.TrimSpace(rows[i].Email)) < strings.ToLower(strings.TrimSpace(rows[j].Email))
		}
		return left < right
	})

	roomTypeKeys := make([]string, 0, len(roomTypeCapacity))
	for roomType := range roomTypeCapacity {
		roomTypeKeys = append(roomTypeKeys, roomType)
	}
	sort.Strings(roomTypeKeys)

	fillRates := make([]OrganizerDashboardRoomFillRateResponse, 0, len(roomTypeKeys))
	for i := range roomTypeKeys {
		roomType := roomTypeKeys[i]
		registered := roomTypeRegistrations[roomType]
		capacity := roomTypeCapacity[roomType]
		fillRates = append(fillRates, OrganizerDashboardRoomFillRateResponse{
			RoomType:      roomType,
			Registrations: registered,
			Capacity:      capacity,
			FillRatePct:   toPercentage(registered, capacity),
		})
	}

	return &OrganizerDashboardResponse{
		ConferenceSlug:      conf.Slug,
		TotalRegistrations:  len(bookings),
		Capacity:            conf.Capacity,
		CapacityUsagePct:    toPercentage(len(bookings), conf.Capacity),
		RoomFillRates:       fillRates,
		Attendees:           rows,
		AppliedRoomType:     strings.ToLower(strings.TrimSpace(filters.RoomType)),
		AppliedBookingState: strings.ToLower(strings.TrimSpace(filters.BookingStatus)),
		AppliedSearch:       strings.TrimSpace(filters.Search),
	}, nil
}

// ExportConferenceBookingsCSV exports conference bookings in CSV format for organizers.
func (s *AttendeeService) ExportConferenceBookingsCSV(
	ctx context.Context,
	slug string,
	requesterUserID int64,
	includeCancelled bool,
) ([]byte, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to export conference bookings: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}

	if s.organizerRepo == nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", ErrOrganizerForbidden)
	}

	isOrganizer, err := s.organizerRepo.IsOrganizer(ctx, conf.ID, requesterUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}
	if !isOrganizer {
		return nil, fmt.Errorf("failed to export conference bookings: %w", ErrOrganizerForbidden)
	}

	var bookings []*models.Booking
	if includeCancelled {
		bookings, err = s.bookingRepo.ListByConferenceIncludingCancelled(ctx, conf.ID)
	} else {
		bookings, err = s.bookingRepo.ListByConference(ctx, conf.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}

	rooms, err := s.roomRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}
	roomsByID := make(map[int64]*models.Room, len(rooms))
	for i := range rooms {
		roomsByID[rooms[i].ID] = rooms[i]
	}

	dates := fmt.Sprintf("%s to %s", conf.StartDate.Format("2006-01-02"), conf.EndDate.Format("2006-01-02"))
	rows := make([]ExportBookingRow, 0, len(bookings))
	for i := range bookings {
		booking := bookings[i]
		user, userErr := s.userRepo.GetByID(ctx, booking.UserID)
		if userErr != nil {
			slog.Error("failed to fetch attendee user for csv export", "error", userErr, "conference_id", conf.ID, "user_id", booking.UserID)
			continue
		}

		room := ""
		if roomModel, ok := roomsByID[booking.RoomID]; ok {
			room = strings.TrimSpace(roomModel.RoomNumber)
		}

		rows = append(rows, ExportBookingRow{
			Name:  displayNameOrFallback(user.DisplayName),
			Email: strings.TrimSpace(user.Email),
			Room:  room,
			Dates: dates,
			Notes: strings.TrimSpace(booking.Notes),
		})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(rows[i].Name))
		right := strings.ToLower(strings.TrimSpace(rows[j].Name))
		if left == right {
			return strings.ToLower(strings.TrimSpace(rows[i].Email)) < strings.ToLower(strings.TrimSpace(rows[j].Email))
		}
		return left < right
	})

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"name", "email", "room", "dates", "dietary/special notes"}); err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}
	for i := range rows {
		row := rows[i]
		if err := writer.Write([]string{row.Name, row.Email, row.Room, row.Dates, row.Notes}); err != nil {
			return nil, fmt.Errorf("failed to export conference bookings: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}

	return buf.Bytes(), nil
}

func organizerDashboardRowMatches(row OrganizerDashboardAttendeeResponse, filters OrganizerDashboardFilters) bool {
	roomTypeFilter := strings.ToLower(strings.TrimSpace(filters.RoomType))
	if roomTypeFilter != "" && roomTypeFilter != "all" && row.RoomType != roomTypeFilter {
		return false
	}

	bookingStatusFilter := strings.ToLower(strings.TrimSpace(filters.BookingStatus))
	if bookingStatusFilter != "" && bookingStatusFilter != "all" && row.BookingStatus != bookingStatusFilter {
		return false
	}

	if filters.HasSpecialRequests && strings.TrimSpace(row.DietaryAccessibilityNote) == "" {
		return false
	}

	search := strings.ToLower(strings.TrimSpace(filters.Search))
	if search != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(row.Name)), search) {
		return false
	}

	return true
}

func toPercentage(numerator, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}

	return (float64(numerator) / float64(denominator)) * 100
}

func displayNameOrFallback(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Attendee"
	}

	return trimmed
}

package service

import (
	"context"
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

// AttendeeService handles attendee listing logic.
type AttendeeService struct {
	confRepo    repository.ConferenceRepository
	bookingRepo repository.BookingRepository
	roomRepo    repository.RoomRepository
	userRepo    repository.UserRepository
}

// NewAttendeeService creates a new AttendeeService.
func NewAttendeeService(
	confRepo repository.ConferenceRepository,
	bookingRepo repository.BookingRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
) *AttendeeService {
	return &AttendeeService{
		confRepo:    confRepo,
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		userRepo:    userRepo,
	}
}

// ListAttendees returns public attendees for a conference and an aggregate count of private attendees.
func (s *AttendeeService) ListAttendees(ctx context.Context, slug string) (*AttendeeListResponse, error) {
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

	for i := range bookings {
		booking := bookings[i]
		if strings.EqualFold(strings.TrimSpace(booking.PrivacySetting), "private") {
			result.PrivateAttendeesCount++
			continue
		}

		user, userErr := s.userRepo.GetByID(ctx, booking.UserID)
		if userErr != nil {
			slog.Error("failed to fetch attendee user", "error", userErr, "conference_id", conf.ID, "user_id", booking.UserID)
			continue
		}

		if strings.EqualFold(strings.TrimSpace(user.PrivacySetting), "private") {
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

func displayNameOrFallback(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Attendee"
	}

	return trimmed
}

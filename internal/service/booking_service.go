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

type bookingRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Booking, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error)
	ListByUser(ctx context.Context, userID int64) ([]*models.Booking, error)
	Cancel(ctx context.Context, bookingID int64) (*models.Booking, error)
	CancelAndCancelPendingOutgoingRequests(ctx context.Context, bookingID, userID int64) (*models.Booking, int64, error)
}

type BookingService struct {
	bookingRepo bookingRepository
	roomRepo    repository.RoomRepository
	confRepo    repository.ConferenceRepository
	userRepo    repository.UserRepository
}

type BookingRoomResponse struct {
	ID            int64   `json:"id"`
	RoomNumber    string  `json:"room_number"`
	RoomType      string  `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
}

type BookingConferenceResponse struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type BookingRoommateResponse struct {
	DisplayName    string `json:"display_name"`
	PrivacySetting string `json:"privacy_setting"`
}

type BookingResponse struct {
	ID             int64                     `json:"id"`
	RoomID         int64                     `json:"room_id"`
	ConferenceID   int64                     `json:"conference_id"`
	ConferenceSlug string                    `json:"conference_slug"`
	Status         string                    `json:"status"`
	PrivacySetting string                    `json:"privacy_setting"`
	Notes          string                    `json:"notes,omitempty"`
	CreatedAt      string                    `json:"created_at"`
	ConfirmedAt    string                    `json:"confirmed_at,omitempty"`
	CancelledAt    string                    `json:"cancelled_at,omitempty"`
	Room           BookingRoomResponse       `json:"room"`
	Conference     BookingConferenceResponse `json:"conference"`
	Roommates      []BookingRoommateResponse `json:"roommates"`
}

func NewBookingService(
	bookingRepo bookingRepository,
	roomRepo repository.RoomRepository,
	confRepo repository.ConferenceRepository,
	userRepo repository.UserRepository,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		confRepo:    confRepo,
		userRepo:    userRepo,
	}
}

func (s *BookingService) ListBookings(ctx context.Context, userID int64) ([]BookingResponse, error) {
	bookings, err := s.bookingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	if len(bookings) == 0 {
		return []BookingResponse{}, nil
	}

	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: failed to load conferences: %w", err)
	}

	conferencesByID := make(map[int64]*models.Conference, len(conferences))
	for i := range conferences {
		conferencesByID[conferences[i].ID] = conferences[i]
	}

	response := make([]BookingResponse, 0, len(bookings))
	for i := range bookings {
		booking, buildErr := s.buildBookingResponse(ctx, bookings[i], conferencesByID)
		if buildErr != nil {
			return nil, fmt.Errorf("failed to list bookings: %w", buildErr)
		}
		response = append(response, booking)
	}

	sort.SliceStable(response, func(i, j int) bool {
		return response[i].CreatedAt > response[j].CreatedAt
	})

	return response, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, userID int64, bookingID int64) (*BookingResponse, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingNotFound)
		}
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	if booking.UserID != userID {
		return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingForbidden)
	}

	if booking.Status == models.BookingStatusCancelled {
		return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingNotFound)
	}

	cancelledBooking, cancelledCount, err := s.bookingRepo.CancelAndCancelPendingOutgoingRequests(ctx, bookingID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingNotFound)
		}
		return nil, fmt.Errorf("failed to cancel booking transactionally: %w", err)
	}

	roommates, err := s.bookingRepo.ListByRoom(ctx, cancelledBooking.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: failed to load remaining roommates: %w", err)
	}

	remainingRoommates := int64(len(roommates))
	slog.Info("roommate notification event (conceptual)",
		"booking_id", bookingID,
		"room_id", cancelledBooking.RoomID,
		"remaining_roommates", remainingRoommates,
		"cancelled_pending_outgoing_requests", cancelledCount,
	)

	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: failed to load conferences: %w", err)
	}
	conferencesByID := make(map[int64]*models.Conference, len(conferences))
	for i := range conferences {
		conferencesByID[conferences[i].ID] = conferences[i]
	}

	result, err := s.buildBookingResponse(ctx, cancelledBooking, conferencesByID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	return &result, nil
}

func (s *BookingService) buildBookingResponse(
	ctx context.Context,
	booking *models.Booking,
	conferencesByID map[int64]*models.Conference,
) (BookingResponse, error) {
	room, err := s.roomRepo.GetByID(ctx, booking.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return BookingResponse{}, fmt.Errorf("failed to load room for booking %d: %w", booking.ID, ErrRoomNotFound)
		}
		return BookingResponse{}, fmt.Errorf("failed to load room for booking %d: %w", booking.ID, err)
	}

	conference := conferencesByID[booking.ConferenceID]
	if conference == nil {
		return BookingResponse{}, fmt.Errorf("failed to load conference for booking %d: %w", booking.ID, ErrConferenceNotFound)
	}

	roomOccupants, err := s.bookingRepo.ListByRoom(ctx, booking.RoomID)
	if err != nil {
		return BookingResponse{}, fmt.Errorf("failed to load roommates for booking %d: %w", booking.ID, err)
	}

	roommates := make([]BookingRoommateResponse, 0)
	for i := range roomOccupants {
		if roomOccupants[i].UserID == booking.UserID {
			continue
		}

		displayName := "Attendee"
		if strings.EqualFold(strings.TrimSpace(roomOccupants[i].PrivacySetting), "private") {
			displayName = "Private attendee"
		} else {
			user, userErr := s.userRepo.GetByID(ctx, roomOccupants[i].UserID)
			if userErr == nil {
				if strings.TrimSpace(user.DisplayName) != "" {
					displayName = strings.TrimSpace(user.DisplayName)
				}
			}
		}

		roommates = append(roommates, BookingRoommateResponse{
			DisplayName:    displayName,
			PrivacySetting: strings.TrimSpace(roomOccupants[i].PrivacySetting),
		})
	}

	result := BookingResponse{
		ID:             booking.ID,
		RoomID:         booking.RoomID,
		ConferenceID:   booking.ConferenceID,
		ConferenceSlug: conference.Slug,
		Status:         string(booking.Status),
		PrivacySetting: booking.PrivacySetting,
		Notes:          strings.TrimSpace(booking.Notes),
		CreatedAt:      booking.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Room: BookingRoomResponse{
			ID:            room.ID,
			RoomNumber:    room.RoomNumber,
			RoomType:      room.RoomType,
			PricePerNight: room.PricePerNight,
		},
		Conference: BookingConferenceResponse{
			ID:        conference.ID,
			Slug:      conference.Slug,
			Name:      conference.Name,
			StartDate: conference.StartDate.Format("2006-01-02"),
			EndDate:   conference.EndDate.Format("2006-01-02"),
		},
		Roommates: roommates,
	}

	if booking.ConfirmedAt != nil {
		result.ConfirmedAt = booking.ConfirmedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	if booking.CancelledAt != nil {
		result.CancelledAt = booking.CancelledAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	return result, nil
}

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

var validBookingPrivacySettings = map[string]bool{
	"public":  true,
	"private": true,
}

type bookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetByID(ctx context.Context, id int64) (*models.Booking, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error)
	ListByUser(ctx context.Context, userID int64) ([]*models.Booking, error)
	GetActiveByUserAndConference(ctx context.Context, userID, conferenceID int64) (*models.Booking, error)
	Cancel(ctx context.Context, bookingID int64) (*models.Booking, error)
	CancelAndCancelPendingOutgoingRequests(ctx context.Context, bookingID, userID int64) (*models.Booking, int64, error)
}

type BookingService struct {
	bookingRepo bookingRepository
	roomRepo    repository.RoomRepository
	confRepo    repository.ConferenceRepository
	userRepo    repository.UserRepository
	notifier    HotelEmailNotifier
}

type BookingRoomResponse struct {
	ID             int64   `json:"id"`
	RoomNumber     string  `json:"room_number"`
	RoomType       string  `json:"room_type"`
	PricePerNight  float64 `json:"price_per_night"`
	Capacity       int     `json:"capacity"`
	SpotsTaken     int     `json:"spots_taken"`
	SpotsAvailable int     `json:"spots_available"`
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

type CreateBookingInput struct {
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	PrivacySetting string `json:"privacy_setting"`
	Notes          string `json:"notes,omitempty"`
}

func NewBookingService(
	bookingRepo bookingRepository,
	roomRepo repository.RoomRepository,
	confRepo repository.ConferenceRepository,
	userRepo repository.UserRepository,
	notifier ...HotelEmailNotifier,
) *BookingService {
	var hotelNotifier HotelEmailNotifier
	if len(notifier) > 0 {
		hotelNotifier = notifier[0]
	}

	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		confRepo:    confRepo,
		userRepo:    userRepo,
		notifier:    hotelNotifier,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID int64, input CreateBookingInput) (*BookingResponse, error) {
	privacySetting := strings.ToLower(strings.TrimSpace(input.PrivacySetting))
	if !validBookingPrivacySettings[privacySetting] {
		return nil, fmt.Errorf("failed to create booking: %w", ErrInvalidPrivacySetting)
	}

	room, err := s.roomRepo.GetByID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return nil, fmt.Errorf("failed to create booking: %w", ErrRoomNotFound)
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	if room.ConferenceID != input.ConferenceID {
		return nil, fmt.Errorf("failed to create booking: %w", ErrRoomNotFound)
	}

	activeBooking, err := s.bookingRepo.GetActiveByUserAndConference(ctx, userID, input.ConferenceID)
	if err != nil && !errors.Is(err, repository.ErrBookingNotFound) {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	if activeBooking != nil {
		return nil, fmt.Errorf("failed to create booking: %w", ErrAlreadyBooked)
	}

	roomBookings, err := s.bookingRepo.ListByRoom(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	if len(roomBookings) >= room.Capacity {
		return nil, fmt.Errorf("failed to create booking: %w", ErrRoomFull)
	}

	created, err := s.bookingRepo.Create(ctx, &models.Booking{
		RoomID:         input.RoomID,
		UserID:         userID,
		ConferenceID:   input.ConferenceID,
		Status:         models.BookingStatusRequested,
		PrivacySetting: privacySetting,
		Notes:          strings.TrimSpace(input.Notes),
	})
	if err != nil {
		if errors.Is(err, repository.ErrBookingExists) {
			return nil, fmt.Errorf("failed to create booking: %w", ErrAlreadyBooked)
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	if s.notifier != nil {
		if notifyErr := s.notifier.NotifyBookingCreated(ctx, created); notifyErr != nil {
			slog.Error("failed to send hotel booking creation email",
				"error", notifyErr,
				"booking_id", created.ID,
				"conference_id", created.ConferenceID,
				"user_id", userID,
			)
		}
	}

	conferencesByID, err := s.loadConferencesByID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	result, err := s.buildBookingResponse(ctx, created, conferencesByID)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return &result, nil
}

func (s *BookingService) ListBookings(ctx context.Context, userID int64) ([]BookingResponse, error) {
	bookings, err := s.bookingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	if len(bookings) == 0 {
		return []BookingResponse{}, nil
	}

	conferencesByID, err := s.loadConferencesByID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	response := make([]BookingResponse, 0, len(bookings))
	for i := range bookings {
		if bookings[i].Status == models.BookingStatusCancelled {
			continue
		}

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

	if s.notifier != nil {
		if notifyErr := s.notifier.NotifyBookingCancelled(ctx, cancelledBooking); notifyErr != nil {
			slog.Error("failed to send hotel cancellation email",
				"error", notifyErr,
				"booking_id", cancelledBooking.ID,
				"conference_id", cancelledBooking.ConferenceID,
			)
		}
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

	conferencesByID, err := s.loadConferencesByID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	result, err := s.buildBookingResponse(ctx, cancelledBooking, conferencesByID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	return &result, nil
}

func (s *BookingService) loadConferencesByID(ctx context.Context) (map[int64]*models.Conference, error) {
	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load conferences: %w", err)
	}

	conferencesByID := make(map[int64]*models.Conference, len(conferences))
	for i := range conferences {
		conferencesByID[conferences[i].ID] = conferences[i]
	}

	return conferencesByID, nil
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
			ID:             room.ID,
			RoomNumber:     room.RoomNumber,
			RoomType:       room.RoomType,
			PricePerNight:  room.PricePerNight,
			Capacity:       room.Capacity,
			SpotsTaken:     len(roomOccupants),
			SpotsAvailable: max(room.Capacity-len(roomOccupants), 0),
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

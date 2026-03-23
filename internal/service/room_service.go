package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

// RoomResponse wraps a room with computed availability fields.
type RoomResponse struct {
	ID             int64              `json:"id"`
	ConferenceID   int64              `json:"conference_id"`
	RoomNumber     string             `json:"room_number"`
	RoomType       string             `json:"room_type"`
	PricePerNight  float64            `json:"price_per_night"`
	Capacity       int                `json:"capacity"`
	SpotsTaken     int                `json:"spots_taken"`
	SpotsAvailable int                `json:"spots_available"`
	Occupants      []OccupantResponse `json:"occupants"`
}

// OccupantResponse represents a room occupant with privacy-aware fields.
type OccupantResponse struct {
	UserID      *int64 `json:"user_id,omitempty"`
	DisplayName string `json:"display_name"`
}

// RoomService handles room business logic.
type RoomService struct {
	roomRepo    repository.RoomRepository
	bookingRepo repository.BookingRepository
	confRepo    repository.ConferenceRepository
	userRepo    repository.UserRepository
}

// NewRoomService creates a new RoomService.
func NewRoomService(roomRepo repository.RoomRepository, bookingRepo repository.BookingRepository, confRepo repository.ConferenceRepository, userRepo repository.UserRepository) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		confRepo:    confRepo,
		userRepo:    userRepo,
	}
}

// ListRooms returns all rooms for a conference with computed availability.
func (s *RoomService) ListRooms(ctx context.Context, slug string) ([]*RoomResponse, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to list rooms: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	rooms, err := s.roomRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	bookings, err := s.bookingRepo.ListByConference(ctx, conf.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	// Group bookings by room ID
	bookingsByRoom := make(map[int64][]*models.Booking)
	for _, b := range bookings {
		bookingsByRoom[b.RoomID] = append(bookingsByRoom[b.RoomID], b)
	}

	responses := make([]*RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		roomBookings := bookingsByRoom[room.ID]
		spotsTaken := len(roomBookings)

		occupants := make([]OccupantResponse, 0, len(roomBookings))
		for _, b := range roomBookings {
			occupant := OccupantResponse{}
			if b.PrivacySetting == "public" {
				user, userErr := s.userRepo.GetByID(ctx, b.UserID)
				if userErr != nil {
					slog.Error("failed to fetch user for occupant", "error", userErr, "user_id", b.UserID)
					occupant.DisplayName = "Unknown"
				} else {
					occupant.UserID = &user.ID
					occupant.DisplayName = user.DisplayName
				}
			} else {
				occupant.DisplayName = "Private attendee"
			}
			occupants = append(occupants, occupant)
		}

		responses = append(responses, &RoomResponse{
			ID:             room.ID,
			ConferenceID:   room.ConferenceID,
			RoomNumber:     room.RoomNumber,
			RoomType:       room.RoomType,
			PricePerNight:  room.PricePerNight,
			Capacity:       room.Capacity,
			SpotsTaken:     spotsTaken,
			SpotsAvailable: room.Capacity - spotsTaken,
			Occupants:      occupants,
		})
	}

	return responses, nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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

// ManageRoomInput contains create/update payload fields for organizer room management.
type ManageRoomInput struct {
	RoomNumber    string
	RoomType      string
	PricePerNight float64
	Capacity      int
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

// CreateRoom creates a room under a conference slug.
func (s *RoomService) CreateRoom(ctx context.Context, slug string, input ManageRoomInput) (*models.Room, error) {
	roomModel, err := validateManageRoomInput(input)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to create room: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to create room: %w", err)
	}
	roomModel.ConferenceID = conf.ID

	created, err := s.roomRepo.Create(ctx, roomModel)
	if err != nil {
		if errors.Is(err, repository.ErrRoomExists) {
			return nil, fmt.Errorf("failed to create room: %w", ErrRoomExists)
		}
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return created, nil
}

// UpdateRoom updates an existing room identified by conference slug + room number.
func (s *RoomService) UpdateRoom(ctx context.Context, slug string, roomNumber string, input ManageRoomInput) (*models.Room, error) {
	roomNumber = strings.TrimSpace(roomNumber)
	if roomNumber == "" {
		return nil, fmt.Errorf("failed to update room: %w", ErrInvalidRoomInput)
	}

	roomModel, err := validateManageRoomInput(input)
	if err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to update room: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to update room: %w", err)
	}
	roomModel.ConferenceID = conf.ID

	updated, err := s.roomRepo.UpdateByConferenceAndNumber(ctx, conf.ID, roomNumber, roomModel)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return nil, fmt.Errorf("failed to update room: %w", ErrRoomNotFound)
		}
		if errors.Is(err, repository.ErrRoomExists) {
			return nil, fmt.Errorf("failed to update room: %w", ErrRoomExists)
		}
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return updated, nil
}

// DeleteRoom removes a room identified by conference slug + room number.
func (s *RoomService) DeleteRoom(ctx context.Context, slug string, roomNumber string) error {
	roomNumber = strings.TrimSpace(roomNumber)
	if roomNumber == "" {
		return fmt.Errorf("failed to delete room: %w", ErrInvalidRoomInput)
	}

	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return fmt.Errorf("failed to delete room: %w", ErrConferenceNotFound)
		}
		return fmt.Errorf("failed to delete room: %w", err)
	}

	if err := s.roomRepo.DeleteByConferenceAndNumber(ctx, conf.ID, roomNumber); err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return fmt.Errorf("failed to delete room: %w", ErrRoomNotFound)
		}
		if errors.Is(err, repository.ErrRoomHasBookings) {
			return fmt.Errorf("failed to delete room: %w", ErrRoomHasBookings)
		}
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
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

		spotsAvailable := room.Capacity - spotsTaken
		if spotsAvailable < 0 {
			spotsAvailable = 0
		}

		responses = append(responses, &RoomResponse{
			ID:             room.ID,
			ConferenceID:   room.ConferenceID,
			RoomNumber:     room.RoomNumber,
			RoomType:       room.RoomType,
			PricePerNight:  room.PricePerNight,
			Capacity:       room.Capacity,
			SpotsTaken:     spotsTaken,
			SpotsAvailable: spotsAvailable,
			Occupants:      occupants,
		})
	}

	return responses, nil
}

func validateManageRoomInput(input ManageRoomInput) (*models.Room, error) {
	roomNumber := strings.TrimSpace(input.RoomNumber)
	roomType := strings.ToLower(strings.TrimSpace(input.RoomType))

	if roomNumber == "" {
		return nil, ErrInvalidRoomInput
	}
	if roomType != "single" && roomType != "double" && roomType != "triple" {
		return nil, ErrInvalidRoomInput
	}
	if input.PricePerNight <= 0 {
		return nil, ErrInvalidRoomInput
	}
	if input.Capacity <= 0 {
		return nil, ErrInvalidRoomInput
	}

	return &models.Room{
		RoomNumber:    roomNumber,
		RoomType:      roomType,
		PricePerNight: input.PricePerNight,
		Capacity:      input.Capacity,
	}, nil
}

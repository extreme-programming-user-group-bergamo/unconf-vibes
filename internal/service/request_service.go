package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

type requestRepository interface {
	Create(ctx context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error)
	GetByID(ctx context.Context, id int64) (*models.RoommateRequest, error)
	ListByUser(ctx context.Context, userID int64) ([]*models.RoommateRequest, error)
	UpdateStatus(ctx context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error)
}

type requestBookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error)
	GetActiveByUserAndConference(ctx context.Context, userID, conferenceID int64) (*models.Booking, error)
	UpdateRoom(ctx context.Context, bookingID, roomID int64) (*models.Booking, error)
}

type requestRoomRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Room, error)
}

type RequestService struct {
	requestRepo requestRepository
	bookingRepo requestBookingRepository
	roomRepo    requestRoomRepository
}

type CreateRoommateRequestInput struct {
	TargetID int64 `json:"target_id"`
	RoomID   int64 `json:"room_id"`
}

type RoommateRequestView struct {
	ID          int64                        `json:"id"`
	RequesterID int64                        `json:"requester_id"`
	TargetID    int64                        `json:"target_id"`
	RoomID      int64                        `json:"room_id"`
	Status      models.RoommateRequestStatus `json:"status"`
	Direction   string                       `json:"direction"`
	CreatedAt   string                       `json:"created_at"`
}

func NewRequestService(
	requestRepo requestRepository,
	bookingRepo requestBookingRepository,
	roomRepo requestRoomRepository,
) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
	}
}

func (s *RequestService) CreateRequest(ctx context.Context, requesterID int64, input CreateRoommateRequestInput) (*models.RoommateRequest, error) {
	if requesterID == input.TargetID {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrCannotRequestSelf)
	}

	room, err := s.roomRepo.GetByID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrRoomNotFound)
		}
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	bookings, err := s.bookingRepo.ListByRoom(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	requesterInRoom := false
	for _, booking := range bookings {
		if booking.UserID == requesterID {
			requesterInRoom = true
			break
		}
	}
	if !requesterInRoom {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrRequesterNotInRoom)
	}

	if len(bookings) >= room.Capacity {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrRoomFull)
	}

	created, err := s.requestRepo.Create(ctx, &models.RoommateRequest{
		RequesterID: requesterID,
		TargetID:    input.TargetID,
		RoomID:      input.RoomID,
		Status:      models.RoommateRequestStatusPending,
	})
	if err != nil {
		if errors.Is(err, repository.ErrRoommateRequestExists) {
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrDuplicateRequest)
		}
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	return created, nil
}

func (s *RequestService) ListRequests(ctx context.Context, userID int64) ([]RoommateRequestView, error) {
	requests, err := s.requestRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roommate requests: %w", err)
	}

	result := make([]RoommateRequestView, 0, len(requests))
	for _, req := range requests {
		direction := "incoming"
		if req.RequesterID == userID {
			direction = "outgoing"
		}
		result = append(result, RoommateRequestView{
			ID:          req.ID,
			RequesterID: req.RequesterID,
			TargetID:    req.TargetID,
			RoomID:      req.RoomID,
			Status:      req.Status,
			Direction:   direction,
			CreatedAt:   req.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, nil
}

func (s *RequestService) AcceptRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrRoommateRequestNotFound) {
			return nil, fmt.Errorf("failed to accept roommate request: %w", ErrRequestNotFound)
		}
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	if req.TargetID != userID {
		return nil, fmt.Errorf("failed to accept roommate request: %w", ErrRequestForbidden)
	}

	if req.Status != models.RoommateRequestStatusPending {
		return nil, fmt.Errorf("failed to accept roommate request: %w", ErrInvalidRequestState)
	}

	room, err := s.roomRepo.GetByID(ctx, req.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return nil, fmt.Errorf("failed to accept roommate request: %w", ErrRoomNotFound)
		}
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	roomBookings, err := s.bookingRepo.ListByRoom(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	var requesterBooking *models.Booking
	for _, booking := range roomBookings {
		if booking.UserID == req.RequesterID {
			requesterBooking = booking
			break
		}
	}
	if requesterBooking == nil {
		return nil, fmt.Errorf("failed to accept roommate request: %w", ErrRequesterNotInRoom)
	}

	targetBooking, err := s.bookingRepo.GetActiveByUserAndConference(ctx, req.TargetID, requesterBooking.ConferenceID)
	if err != nil && !errors.Is(err, repository.ErrBookingNotFound) {
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	targetAlreadyInRoom := targetBooking != nil && targetBooking.RoomID == req.RoomID
	if !targetAlreadyInRoom && len(roomBookings) >= room.Capacity {
		return nil, fmt.Errorf("failed to accept roommate request: %w", ErrRoomFull)
	}

	if targetBooking == nil {
		_, err = s.bookingRepo.Create(ctx, &models.Booking{
			RoomID:         req.RoomID,
			UserID:         req.TargetID,
			ConferenceID:   requesterBooking.ConferenceID,
			Status:         models.BookingStatusRequested,
			PrivacySetting: "public",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to accept roommate request: %w", err)
		}
	} else if targetBooking.RoomID != req.RoomID {
		_, err = s.bookingRepo.UpdateRoom(ctx, targetBooking.ID, req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("failed to accept roommate request: %w", err)
		}
	}

	updated, err := s.requestRepo.UpdateStatus(ctx, req.ID, models.RoommateRequestStatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	return updated, nil
}

func (s *RequestService) DeclineRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error) {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrRoommateRequestNotFound) {
			return nil, fmt.Errorf("failed to decline roommate request: %w", ErrRequestNotFound)
		}
		return nil, fmt.Errorf("failed to decline roommate request: %w", err)
	}

	if req.TargetID != userID {
		return nil, fmt.Errorf("failed to decline roommate request: %w", ErrRequestForbidden)
	}

	if req.Status != models.RoommateRequestStatusPending {
		return nil, fmt.Errorf("failed to decline roommate request: %w", ErrInvalidRequestState)
	}

	updated, err := s.requestRepo.UpdateStatus(ctx, req.ID, models.RoommateRequestStatusDeclined)
	if err != nil {
		return nil, fmt.Errorf("failed to decline roommate request: %w", err)
	}

	return updated, nil
}

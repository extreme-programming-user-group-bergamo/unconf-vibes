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
	Cancel(ctx context.Context, bookingID int64) (*models.Booking, error)
}

type requestRoomRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Room, error)
}

type requestUserRepository interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*models.User, error)
}

type RequestService struct {
	requestRepo requestRepository
	bookingRepo requestBookingRepository
	roomRepo    requestRoomRepository
	userRepo    requestUserRepository
	notifier    HotelEmailNotifier
}

type CreateRoommateRequestInput struct {
	TargetID       int64  `json:"target_id,omitempty"`
	TargetUsername string `json:"target_username,omitempty"`
	RoomID         int64  `json:"room_id"`
}

type RoommateRequestView struct {
	ID             int64                        `json:"id"`
	RequesterID    int64                        `json:"requester_id"`
	TargetID       int64                        `json:"target_id"`
	RoomID         int64                        `json:"room_id"`
	RoomNumber     string                       `json:"room_number,omitempty"`
	RoomType       string                       `json:"room_type,omitempty"`
	ConferenceID   int64                        `json:"conference_id"`
	ConferenceSlug string                       `json:"conference_slug,omitempty"`
	Status         models.RoommateRequestStatus `json:"status"`
	Direction      string                       `json:"direction"`
	RequesterName  string                       `json:"requester_name,omitempty"`
	TargetName     string                       `json:"target_name,omitempty"`
	CreatedAt      string                       `json:"created_at"`
}

func NewRequestService(
	requestRepo requestRepository,
	bookingRepo requestBookingRepository,
	roomRepo requestRoomRepository,
	userRepo requestUserRepository,
	notifier ...HotelEmailNotifier,
) *RequestService {
	var hotelNotifier HotelEmailNotifier
	if len(notifier) > 0 {
		hotelNotifier = notifier[0]
	}

	return &RequestService{
		requestRepo: requestRepo,
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		notifier:    hotelNotifier,
	}
}

func (s *RequestService) CreateRequest(ctx context.Context, requesterID int64, input CreateRoommateRequestInput) (*models.RoommateRequest, error) {
	if input.TargetID > 0 && requesterID == input.TargetID {
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

	var requesterBooking *models.Booking
	for _, booking := range bookings {
		if booking.UserID == requesterID {
			requesterBooking = booking
			break
		}
	}
	if requesterBooking == nil {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrRequesterNotInRoom)
	}

	if len(bookings) >= room.Capacity {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrRoomFull)
	}

	target, err := s.resolveTargetUser(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	if target.ID == requesterID {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrCannotRequestSelf)
	}

	targetBooking, err := s.bookingRepo.GetActiveByUserAndConference(ctx, target.ID, requesterBooking.ConferenceID)
	if err != nil && !errors.Is(err, repository.ErrBookingNotFound) {
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}
	if targetBooking != nil {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrTargetAlreadyBooked)
	}

	created, err := s.requestRepo.Create(ctx, &models.RoommateRequest{
		RequesterID: requesterID,
		TargetID:    target.ID,
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

func (s *RequestService) resolveTargetUser(ctx context.Context, input CreateRoommateRequestInput) (*models.User, error) {
	if input.TargetID > 0 {
		target, err := s.userRepo.GetByID(ctx, input.TargetID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return nil, ErrTargetNotFound
			}
			return nil, err
		}
		return target, nil
	}

	username := strings.TrimSpace(strings.TrimPrefix(input.TargetUsername, "@"))
	if username == "" {
		return nil, ErrTargetNotFound
	}

	target, err := s.userRepo.GetByGitHubID(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	return target, nil
}

func (s *RequestService) ListRequests(ctx context.Context, userID int64) ([]RoommateRequestView, error) {
	requests, err := s.requestRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roommate requests: %w", err)
	}

	result := make([]RoommateRequestView, 0, len(requests))
	for _, req := range requests {
		room, roomErr := s.roomRepo.GetByID(ctx, req.RoomID)
		if roomErr != nil {
			return nil, fmt.Errorf("failed to list roommate requests: failed to load room %d: %w", req.RoomID, roomErr)
		}

		requesterName, err := s.lookupUserDisplayName(ctx, req.RequesterID)
		if err != nil {
			return nil, fmt.Errorf("failed to list roommate requests: failed to load requester %d: %w", req.RequesterID, err)
		}

		targetName, err := s.lookupUserDisplayName(ctx, req.TargetID)
		if err != nil {
			return nil, fmt.Errorf("failed to list roommate requests: failed to load target %d: %w", req.TargetID, err)
		}

		direction := "incoming"
		if req.RequesterID == userID {
			direction = "outgoing"
		}
		result = append(result, RoommateRequestView{
			ID:            req.ID,
			RequesterID:   req.RequesterID,
			TargetID:      req.TargetID,
			RoomID:        req.RoomID,
			RoomNumber:    strings.TrimSpace(room.RoomNumber),
			RoomType:      strings.TrimSpace(room.RoomType),
			ConferenceID:  room.ConferenceID,
			Status:        req.Status,
			Direction:     direction,
			RequesterName: requesterName,
			TargetName:    targetName,
			CreatedAt:     req.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, nil
}

func (s *RequestService) lookupUserDisplayName(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	displayName := strings.TrimSpace(user.DisplayName)
	if displayName != "" {
		return displayName, nil
	}

	githubID := strings.TrimSpace(user.GitHubID)
	if githubID != "" {
		return githubID, nil
	}

	return fmt.Sprintf("user #%d", userID), nil
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

	var createdBooking *models.Booking
	var movedBookingID int64
	var originalRoomID int64

	if targetBooking == nil {
		created, createErr := s.bookingRepo.Create(ctx, &models.Booking{
			RoomID:         req.RoomID,
			UserID:         req.TargetID,
			ConferenceID:   requesterBooking.ConferenceID,
			Status:         models.BookingStatusRequested,
			PrivacySetting: "public",
		})
		if createErr != nil {
			return nil, fmt.Errorf("failed to accept roommate request: %w", createErr)
		}
		createdBooking = created
		if s.notifier != nil {
			if notifyErr := s.notifier.NotifyBookingCreated(ctx, created); notifyErr != nil {
				slog.Error("failed to send hotel booking creation email",
					"error", notifyErr,
					"booking_id", created.ID,
					"conference_id", created.ConferenceID,
					"request_id", req.ID,
				)
			}
		}
	} else if targetBooking.RoomID != req.RoomID {
		movedBookingID = targetBooking.ID
		originalRoomID = targetBooking.RoomID
		_, err = s.bookingRepo.UpdateRoom(ctx, targetBooking.ID, req.RoomID)
		if err != nil {
			return nil, fmt.Errorf("failed to accept roommate request: %w", err)
		}
	}

	updated, err := s.requestRepo.UpdateStatus(ctx, req.ID, models.RoommateRequestStatusAccepted)
	if err != nil {
		if rollbackErr := s.rollbackAcceptedRequestBookingChange(ctx, createdBooking, movedBookingID, originalRoomID); rollbackErr != nil {
			return nil, fmt.Errorf("failed to accept roommate request: %w (rollback failed: %v)", err, rollbackErr)
		}
		return nil, fmt.Errorf("failed to accept roommate request: %w", err)
	}

	return updated, nil
}

func (s *RequestService) rollbackAcceptedRequestBookingChange(
	ctx context.Context,
	createdBooking *models.Booking,
	movedBookingID int64,
	originalRoomID int64,
) error {
	if createdBooking != nil {
		_, err := s.bookingRepo.Cancel(ctx, createdBooking.ID)
		if err != nil && !errors.Is(err, repository.ErrBookingNotFound) {
			return fmt.Errorf("failed to rollback created booking %d: %w", createdBooking.ID, err)
		}
		return nil
	}

	if movedBookingID > 0 && originalRoomID > 0 {
		if _, err := s.bookingRepo.UpdateRoom(ctx, movedBookingID, originalRoomID); err != nil {
			return fmt.Errorf("failed to rollback moved booking %d to room %d: %w", movedBookingID, originalRoomID, err)
		}
	}

	return nil
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

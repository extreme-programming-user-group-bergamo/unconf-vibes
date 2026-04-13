package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRequestRepository struct {
	createFn       func(ctx context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error)
	getByIDFn      func(ctx context.Context, id int64) (*models.RoommateRequest, error)
	listByUserFn   func(ctx context.Context, userID int64) ([]*models.RoommateRequest, error)
	updateStatusFn func(ctx context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error)
}

func (m *mockRequestRepository) Create(ctx context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error) {
	return m.createFn(ctx, request)
}
func (m *mockRequestRepository) GetByID(ctx context.Context, id int64) (*models.RoommateRequest, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRequestRepository) ListByUser(ctx context.Context, userID int64) ([]*models.RoommateRequest, error) {
	return m.listByUserFn(ctx, userID)
}
func (m *mockRequestRepository) UpdateStatus(ctx context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error) {
	return m.updateStatusFn(ctx, id, status)
}

type mockRequestBookingRepository struct {
	createFn                       func(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	listByRoomFn                   func(ctx context.Context, roomID int64) ([]*models.Booking, error)
	getActiveByUserAndConferenceFn func(ctx context.Context, userID, conferenceID int64) (*models.Booking, error)
	updateRoomFn                   func(ctx context.Context, bookingID, roomID int64) (*models.Booking, error)
}

func (m *mockRequestBookingRepository) Create(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	return m.createFn(ctx, booking)
}
func (m *mockRequestBookingRepository) ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error) {
	return m.listByRoomFn(ctx, roomID)
}
func (m *mockRequestBookingRepository) GetActiveByUserAndConference(ctx context.Context, userID, conferenceID int64) (*models.Booking, error) {
	return m.getActiveByUserAndConferenceFn(ctx, userID, conferenceID)
}
func (m *mockRequestBookingRepository) UpdateRoom(ctx context.Context, bookingID, roomID int64) (*models.Booking, error) {
	return m.updateRoomFn(ctx, bookingID, roomID)
}

type mockRequestRoomRepository struct {
	getByIDFn func(ctx context.Context, id int64) (*models.Room, error)
}

func (m *mockRequestRoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	return m.getByIDFn(ctx, id)
}

func TestRequestService_CreateRequest_SelfRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{},
		&mockRequestRoomRepository{},
	)

	_, err := svc.CreateRequest(context.Background(), 10, CreateRoommateRequestInput{
		TargetID: 10,
		RoomID:   1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRequestSelf)
}

func TestRequestService_CreateRequest_RoomFullRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{UserID: 1}, {UserID: 2}}, nil
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 7, Capacity: 2}, nil
			},
		},
	)

	_, err := svc.CreateRequest(context.Background(), 1, CreateRoommateRequestInput{
		TargetID: 3,
		RoomID:   7,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRoomFull)
}

func TestRequestService_AcceptRequest_CreatesTargetBookingWhenMissing(t *testing.T) {
	now := time.Now().UTC()
	updateCalled := false
	createCalled := false

	svc := NewRequestService(
		&mockRequestRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.RoommateRequest, error) {
				return &models.RoommateRequest{
					ID:          55,
					RequesterID: 11,
					TargetID:    22,
					RoomID:      33,
					Status:      models.RoommateRequestStatusPending,
					CreatedAt:   now,
				}, nil
			},
			updateStatusFn: func(_ context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error) {
				updateCalled = true
				return &models.RoommateRequest{ID: id, Status: status}, nil
			},
		},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{UserID: 11, RoomID: 33, ConferenceID: 44}}, nil
			},
			getActiveByUserAndConferenceFn: func(_ context.Context, _, _ int64) (*models.Booking, error) {
				return nil, repository.ErrBookingNotFound
			},
			createFn: func(_ context.Context, booking *models.Booking) (*models.Booking, error) {
				createCalled = true
				assert.Equal(t, int64(22), booking.UserID)
				assert.Equal(t, int64(33), booking.RoomID)
				assert.Equal(t, int64(44), booking.ConferenceID)
				return booking, nil
			},
			updateRoomFn: func(_ context.Context, _, _ int64) (*models.Booking, error) {
				return nil, errors.New("should not update room")
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 33, Capacity: 2}, nil
			},
		},
	)

	updated, err := svc.AcceptRequest(context.Background(), 55, 22)
	require.NoError(t, err)
	assert.True(t, createCalled)
	assert.True(t, updateCalled)
	assert.Equal(t, models.RoommateRequestStatusAccepted, updated.Status)
}

func TestRequestService_DeclineRequest_UpdatesStatus(t *testing.T) {
	now := time.Now().UTC()

	svc := NewRequestService(
		&mockRequestRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.RoommateRequest, error) {
				return &models.RoommateRequest{
					ID:          7,
					RequesterID: 10,
					TargetID:    11,
					RoomID:      20,
					Status:      models.RoommateRequestStatusPending,
					CreatedAt:   now,
				}, nil
			},
			updateStatusFn: func(_ context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error) {
				return &models.RoommateRequest{ID: id, Status: status}, nil
			},
		},
		&mockRequestBookingRepository{},
		&mockRequestRoomRepository{},
	)

	result, err := svc.DeclineRequest(context.Background(), 7, 11)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusDeclined, result.Status)
}

func TestRequestService_CreateRequest_RequesterNotInRoomRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				// room is full but requester is not one of the occupants
				return []*models.Booking{{UserID: 2}, {UserID: 3}}, nil
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 7, Capacity: 2}, nil
			},
		},
	)

	_, err := svc.CreateRequest(context.Background(), 1, CreateRoommateRequestInput{
		TargetID: 4,
		RoomID:   7,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequesterNotInRoom)
}

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

type mockRequestUserRepository struct {
	getByIDFn       func(ctx context.Context, id int64) (*models.User, error)
	getByGitHubIDFn func(ctx context.Context, githubID string) (*models.User, error)
}

func (m *mockRequestUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockRequestUserRepository) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	if m.getByGitHubIDFn != nil {
		return m.getByGitHubIDFn(ctx, githubID)
	}
	return nil, repository.ErrUserNotFound
}

func TestRequestService_CreateRequest_SelfRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{},
		&mockRequestRoomRepository{},
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id}, nil
			},
		},
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
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id}, nil
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
		&mockRequestUserRepository{},
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
		&mockRequestUserRepository{},
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
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id}, nil
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

func TestRequestService_CreateRequest_TargetUsernameNotFoundRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{UserID: 1, ConferenceID: 99}}, nil
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 7, Capacity: 2}, nil
			},
		},
		&mockRequestUserRepository{
			getByGitHubIDFn: func(_ context.Context, _ string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		},
	)

	_, err := svc.CreateRequest(context.Background(), 1, CreateRoommateRequestInput{
		TargetUsername: "@missing",
		RoomID:         7,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTargetNotFound)
}

func TestRequestService_CreateRequest_TargetAlreadyBookedRejected(t *testing.T) {
	svc := NewRequestService(
		&mockRequestRepository{},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{UserID: 1, ConferenceID: 77}}, nil
			},
			getActiveByUserAndConferenceFn: func(_ context.Context, userID, conferenceID int64) (*models.Booking, error) {
				return &models.Booking{UserID: userID, ConferenceID: conferenceID, RoomID: 999}, nil
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 7, Capacity: 3}, nil
			},
		},
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id}, nil
			},
		},
	)

	_, err := svc.CreateRequest(context.Background(), 1, CreateRoommateRequestInput{
		TargetID: 2,
		RoomID:   7,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTargetAlreadyBooked)
}

func TestRequestService_CreateRequest_AllowsOpenSpotAfterRoommateDeparture(t *testing.T) {
	createCalled := false

	svc := NewRequestService(
		&mockRequestRepository{
			createFn: func(_ context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error) {
				createCalled = true
				assert.Equal(t, int64(1), request.RequesterID)
				assert.Equal(t, int64(2), request.TargetID)
				assert.Equal(t, int64(7), request.RoomID)
				assert.Equal(t, models.RoommateRequestStatusPending, request.Status)
				return &models.RoommateRequest{ID: 99, RequesterID: 1, TargetID: 2, RoomID: 7, Status: models.RoommateRequestStatusPending}, nil
			},
		},
		&mockRequestBookingRepository{
			listByRoomFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				// requester remains in room after roommate departure
				return []*models.Booking{{UserID: 1, ConferenceID: 77}}, nil
			},
			getActiveByUserAndConferenceFn: func(_ context.Context, _, _ int64) (*models.Booking, error) {
				return nil, repository.ErrBookingNotFound
			},
		},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
				return &models.Room{ID: 7, Capacity: 2}, nil
			},
		},
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id}, nil
			},
		},
	)

	created, err := svc.CreateRequest(context.Background(), 1, CreateRoommateRequestInput{
		TargetID: 2,
		RoomID:   7,
	})
	require.NoError(t, err)
	assert.True(t, createCalled)
	require.NotNil(t, created)
	assert.Equal(t, int64(99), created.ID)
}

func TestRequestService_ListRequests_EnrichesNamesAndRoomDetails(t *testing.T) {
	now := time.Now().UTC()
	svc := NewRequestService(
		&mockRequestRepository{
			listByUserFn: func(_ context.Context, userID int64) ([]*models.RoommateRequest, error) {
				assert.Equal(t, int64(9), userID)
				return []*models.RoommateRequest{
					{
						ID:          1,
						RequesterID: 9,
						TargetID:    10,
						RoomID:      5,
						Status:      models.RoommateRequestStatusPending,
						CreatedAt:   now,
					},
					{
						ID:          2,
						RequesterID: 11,
						TargetID:    9,
						RoomID:      6,
						Status:      models.RoommateRequestStatusAccepted,
						CreatedAt:   now,
					},
				}, nil
			},
		},
		&mockRequestBookingRepository{},
		&mockRequestRoomRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.Room, error) {
				switch id {
				case 5:
					return &models.Room{ID: 5, ConferenceID: 77, RoomNumber: "501", RoomType: "double"}, nil
				case 6:
					return &models.Room{ID: 6, ConferenceID: 77, RoomNumber: "601", RoomType: "single"}, nil
				default:
					return nil, errors.New("unexpected room id")
				}
			},
		},
		&mockRequestUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				switch id {
				case 9:
					return &models.User{ID: 9, DisplayName: "Requester"}, nil
				case 10:
					return &models.User{ID: 10, DisplayName: "Target"}, nil
				case 11:
					return &models.User{ID: 11, DisplayName: "Other"}, nil
				default:
					return nil, errors.New("unexpected user id")
				}
			},
		},
	)

	rows, err := svc.ListRequests(context.Background(), 9)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	assert.Equal(t, "outgoing", rows[0].Direction)
	assert.Equal(t, "Requester", rows[0].RequesterName)
	assert.Equal(t, "Target", rows[0].TargetName)
	assert.Equal(t, "501", rows[0].RoomNumber)
	assert.Equal(t, "double", rows[0].RoomType)
	assert.Equal(t, int64(77), rows[0].ConferenceID)

	assert.Equal(t, "incoming", rows[1].Direction)
	assert.Equal(t, "Other", rows[1].RequesterName)
	assert.Equal(t, "Requester", rows[1].TargetName)
	assert.Equal(t, "601", rows[1].RoomNumber)
}

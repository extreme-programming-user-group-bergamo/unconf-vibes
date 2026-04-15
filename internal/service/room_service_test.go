package service

import (
	"context"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock repositories ---

type mockRoomRepository struct {
	listByConferenceFn func(ctx context.Context, conferenceID int64) ([]*models.Room, error)
	getByIDFn          func(ctx context.Context, id int64) (*models.Room, error)
	getByConfAndNumFn  func(ctx context.Context, conferenceID int64, roomNumber string) (*models.Room, error)
	createFn           func(ctx context.Context, room *models.Room) (*models.Room, error)
	updateFn           func(ctx context.Context, conferenceID int64, roomNumber string, room *models.Room) (*models.Room, error)
	deleteFn           func(ctx context.Context, conferenceID int64, roomNumber string) error
}

func (m *mockRoomRepository) ListByConference(ctx context.Context, conferenceID int64) ([]*models.Room, error) {
	if m.listByConferenceFn != nil {
		return m.listByConferenceFn(ctx, conferenceID)
	}
	return []*models.Room{}, nil
}

func (m *mockRoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRoomRepository) GetByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) (*models.Room, error) {
	if m.getByConfAndNumFn != nil {
		return m.getByConfAndNumFn(ctx, conferenceID, roomNumber)
	}
	return nil, nil
}

func (m *mockRoomRepository) Create(ctx context.Context, room *models.Room) (*models.Room, error) {
	if m.createFn != nil {
		return m.createFn(ctx, room)
	}
	return nil, nil
}

func (m *mockRoomRepository) UpdateByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string, room *models.Room) (*models.Room, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, conferenceID, roomNumber, room)
	}
	return nil, nil
}

func (m *mockRoomRepository) DeleteByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, conferenceID, roomNumber)
	}
	return nil
}

type mockBookingRepository struct {
	listByConferenceFn                   func(ctx context.Context, conferenceID int64) ([]*models.Booking, error)
	listByConferenceIncludingCancelledFn func(ctx context.Context, conferenceID int64) ([]*models.Booking, error)
	listByRoomFn                         func(ctx context.Context, roomID int64) ([]*models.Booking, error)
	countByConferenceFn                  func(ctx context.Context, conferenceID int64) (int, error)
	getByIDFn                            func(ctx context.Context, id int64) (*models.Booking, error)
	createFn                             func(ctx context.Context, booking *models.Booking) (*models.Booking, error)
}

func (m *mockBookingRepository) ListByConference(ctx context.Context, conferenceID int64) ([]*models.Booking, error) {
	if m.listByConferenceFn != nil {
		return m.listByConferenceFn(ctx, conferenceID)
	}
	return []*models.Booking{}, nil
}

func (m *mockBookingRepository) ListByConferenceIncludingCancelled(ctx context.Context, conferenceID int64) ([]*models.Booking, error) {
	if m.listByConferenceIncludingCancelledFn != nil {
		return m.listByConferenceIncludingCancelledFn(ctx, conferenceID)
	}
	return []*models.Booking{}, nil
}

func (m *mockBookingRepository) ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error) {
	if m.listByRoomFn != nil {
		return m.listByRoomFn(ctx, roomID)
	}
	return []*models.Booking{}, nil
}

func (m *mockBookingRepository) CountByConference(ctx context.Context, conferenceID int64) (int, error) {
	if m.countByConferenceFn != nil {
		return m.countByConferenceFn(ctx, conferenceID)
	}
	return 0, nil
}

func (m *mockBookingRepository) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockBookingRepository) Create(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	if m.createFn != nil {
		return m.createFn(ctx, booking)
	}
	return nil, nil
}

type mockUserRepository struct {
	getByIDFn       func(ctx context.Context, id int64) (*models.User, error)
	getByGitHubIDFn func(ctx context.Context, githubID string) (*models.User, error)
	createFn        func(ctx context.Context, user *models.User) (*models.User, error)
	updateFn        func(ctx context.Context, user *models.User) (*models.User, error)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	if m.getByGitHubIDFn != nil {
		return m.getByGitHubIDFn(ctx, githubID)
	}
	return nil, nil
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil, nil
}

// --- helpers ---

func testConference() *models.Conference {
	return &models.Conference{
		ID:       1,
		Slug:     "socrates-26",
		Name:     "SoCraTes 2026",
		Capacity: 100,
	}
}

func testRoom(id int64, confID int64, roomNumber string, capacity int) *models.Room {
	return &models.Room{
		ID:            id,
		ConferenceID:  confID,
		RoomNumber:    roomNumber,
		RoomType:      "double",
		PricePerNight: 120.0,
		Capacity:      capacity,
	}
}

func testBooking(id, roomID, userID, confID int64, privacy string) *models.Booking {
	return &models.Booking{
		ID:             id,
		RoomID:         roomID,
		UserID:         userID,
		ConferenceID:   confID,
		Status:         models.BookingStatusConfirmed,
		PrivacySetting: privacy,
	}
}

func testUser(id int64, name string) *models.User {
	return &models.User{
		ID:          id,
		DisplayName: name,
	}
}

func newTestRoomService(confRepo *mockConferenceRepository, roomRepo *mockRoomRepository, bookingRepo *mockBookingRepository, userRepo *mockUserRepository) *RoomService {
	return NewRoomService(roomRepo, bookingRepo, confRepo, userRepo)
}

// --- tests ---

func TestRoomService_ListRooms_Success(t *testing.T) {
	conf := testConference()
	room1 := testRoom(1, conf.ID, "101", 2)
	room2 := testRoom(2, conf.ID, "102", 3)
	booking1 := testBooking(1, room1.ID, 10, conf.ID, "public")
	user10 := testUser(10, "Alice")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, slug string) (*models.Conference, error) {
				assert.Equal(t, "socrates-26", slug)
				return conf, nil
			},
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, confID int64) ([]*models.Room, error) {
				return []*models.Room{room1, room2}, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, confID int64) ([]*models.Booking, error) {
				return []*models.Booking{booking1}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				assert.Equal(t, int64(10), id)
				return user10, nil
			},
		},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Room 101: 1 booking, capacity 2
	assert.Equal(t, int64(1), results[0].ID)
	assert.Equal(t, 1, results[0].SpotsTaken)
	assert.Equal(t, 1, results[0].SpotsAvailable)
	require.Len(t, results[0].Occupants, 1)
	assert.Equal(t, "Alice", results[0].Occupants[0].DisplayName)
	require.NotNil(t, results[0].Occupants[0].UserID)
	assert.Equal(t, int64(10), *results[0].Occupants[0].UserID)

	// Room 102: 0 bookings, capacity 3
	assert.Equal(t, int64(2), results[1].ID)
	assert.Equal(t, 0, results[1].SpotsTaken)
	assert.Equal(t, 3, results[1].SpotsAvailable)
	assert.Empty(t, results[1].Occupants)
}

func TestRoomService_ListRooms_NoBookings(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 4)

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room}, nil
			},
		},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 0, results[0].SpotsTaken)
	assert.Equal(t, 4, results[0].SpotsAvailable)
	assert.Empty(t, results[0].Occupants)
}

func TestRoomService_ListRooms_PrivateBookingOmitsUserID(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 2)
	booking := testBooking(1, room.ID, 10, conf.ID, "private")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room}, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{booking}, nil
			},
		},
		&mockUserRepository{},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].SpotsTaken)
	require.Len(t, results[0].Occupants, 1)
	assert.Nil(t, results[0].Occupants[0].UserID)
	assert.Equal(t, "Private attendee", results[0].Occupants[0].DisplayName)
}

func TestRoomService_ListRooms_MixedPublicPrivate(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 3)
	publicBooking := testBooking(1, room.ID, 10, conf.ID, "public")
	privateBooking := testBooking(2, room.ID, 20, conf.ID, "private")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room}, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{publicBooking, privateBooking}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return testUser(id, "Alice"), nil
			},
		},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 2, results[0].SpotsTaken)
	assert.Equal(t, 1, results[0].SpotsAvailable)
	require.Len(t, results[0].Occupants, 2)

	// Public occupant
	assert.NotNil(t, results[0].Occupants[0].UserID)
	assert.Equal(t, "Alice", results[0].Occupants[0].DisplayName)

	// Private occupant
	assert.Nil(t, results[0].Occupants[1].UserID)
	assert.Equal(t, "Private attendee", results[0].Occupants[1].DisplayName)
}

func TestRoomService_ListRooms_ConferenceNotFound(t *testing.T) {
	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
				return nil, repository.ErrConferenceNotFound
			},
		},
		&mockRoomRepository{},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListRooms(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}

func TestRoomService_ListRooms_RoomRepoError(t *testing.T) {
	conf := testConference()
	repoErr := errors.New("room repo failure")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return nil, repoErr
			},
		},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListRooms(context.Background(), "socrates-26")
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestRoomService_ListRooms_BookingRepoError(t *testing.T) {
	conf := testConference()
	repoErr := errors.New("booking repo failure")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{}, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return nil, repoErr
			},
		},
		&mockUserRepository{},
	)

	_, err := svc.ListRooms(context.Background(), "socrates-26")
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestRoomService_ListRooms_EmptyRooms(t *testing.T) {
	conf := testConference()

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{}, nil
			},
		},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.NotNil(t, results)
	assert.Empty(t, results)
}

func TestRoomService_ListRooms_UserFetchError_FallsBackToUnknown(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 2)
	booking := testBooking(1, room.ID, 10, conf.ID, "public")

	svc := newTestRoomService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room}, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{booking}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
				return nil, errors.New("user not found")
			},
		},
	)

	results, err := svc.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Len(t, results[0].Occupants, 1)
	assert.Nil(t, results[0].Occupants[0].UserID)
	assert.Equal(t, "Unknown", results[0].Occupants[0].DisplayName)
}

func TestRoomService_CreateRoom_Success(t *testing.T) {
	conf := testConference()
	svc := newTestRoomService(
		&mockConferenceRepository{getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil }},
		&mockRoomRepository{createFn: func(_ context.Context, room *models.Room) (*models.Room, error) {
			assert.Equal(t, conf.ID, room.ConferenceID)
			assert.Equal(t, "101", room.RoomNumber)
			return &models.Room{ID: 1, ConferenceID: conf.ID, RoomNumber: "101", RoomType: "double", PricePerNight: 100, Capacity: 2}, nil
		}},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	created, err := svc.CreateRoom(context.Background(), conf.Slug, ManageRoomInput{
		RoomNumber:    "101",
		RoomType:      "double",
		PricePerNight: 100,
		Capacity:      2,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "101", created.RoomNumber)
}

func TestRoomService_UpdateRoom_Success(t *testing.T) {
	conf := testConference()
	svc := newTestRoomService(
		&mockConferenceRepository{getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil }},
		&mockRoomRepository{updateFn: func(_ context.Context, conferenceID int64, roomNumber string, room *models.Room) (*models.Room, error) {
			assert.Equal(t, conf.ID, conferenceID)
			assert.Equal(t, "101", roomNumber)
			assert.Equal(t, "102", room.RoomNumber)
			return &models.Room{
				ID:            1,
				ConferenceID:  conf.ID,
				RoomNumber:    "102",
				RoomType:      room.RoomType,
				PricePerNight: room.PricePerNight,
				Capacity:      room.Capacity,
			}, nil
		}},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	updated, err := svc.UpdateRoom(context.Background(), conf.Slug, "101", ManageRoomInput{
		RoomNumber:    "102",
		RoomType:      "triple",
		PricePerNight: 180,
		Capacity:      3,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "102", updated.RoomNumber)
}

func TestRoomService_UpdateRoom_NotFoundMapsToServiceError(t *testing.T) {
	conf := testConference()
	svc := newTestRoomService(
		&mockConferenceRepository{getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil }},
		&mockRoomRepository{updateFn: func(_ context.Context, _ int64, _ string, _ *models.Room) (*models.Room, error) {
			return nil, repository.ErrRoomNotFound
		}},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	_, err := svc.UpdateRoom(context.Background(), conf.Slug, "999", ManageRoomInput{
		RoomNumber:    "999",
		RoomType:      "double",
		PricePerNight: 120,
		Capacity:      2,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomService_DeleteRoom_HasBookings(t *testing.T) {
	conf := testConference()
	svc := newTestRoomService(
		&mockConferenceRepository{getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil }},
		&mockRoomRepository{deleteFn: func(_ context.Context, _ int64, _ string) error { return repository.ErrRoomHasBookings }},
		&mockBookingRepository{},
		&mockUserRepository{},
	)

	err := svc.DeleteRoom(context.Background(), conf.Slug, "101")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRoomHasBookings)
}

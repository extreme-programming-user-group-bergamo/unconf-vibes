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

type mockBookingServiceBookingRepo struct {
	getByIDFn              func(context.Context, int64) (*models.Booking, error)
	listByRoom             func(context.Context, int64) ([]*models.Booking, error)
	listByUser             func(context.Context, int64) ([]*models.Booking, error)
	cancelFn               func(context.Context, int64) (*models.Booking, error)
	cancelAndRequestsTxnFn func(context.Context, int64, int64) (*models.Booking, int64, error)
}

func (m *mockBookingServiceBookingRepo) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockBookingServiceBookingRepo) ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error) {
	return m.listByRoom(ctx, roomID)
}
func (m *mockBookingServiceBookingRepo) ListByUser(ctx context.Context, userID int64) ([]*models.Booking, error) {
	return m.listByUser(ctx, userID)
}
func (m *mockBookingServiceBookingRepo) Cancel(ctx context.Context, bookingID int64) (*models.Booking, error) {
	return m.cancelFn(ctx, bookingID)
}
func (m *mockBookingServiceBookingRepo) CancelAndCancelPendingOutgoingRequests(
	ctx context.Context,
	bookingID int64,
	userID int64,
) (*models.Booking, int64, error) {
	return m.cancelAndRequestsTxnFn(ctx, bookingID, userID)
}

type mockBookingServiceRoomRepo struct {
	getByIDFn func(context.Context, int64) (*models.Room, error)
}

func (m *mockBookingServiceRoomRepo) Create(context.Context, *models.Room) (*models.Room, error) {
	panic("not implemented")
}
func (m *mockBookingServiceRoomRepo) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockBookingServiceRoomRepo) ListByConference(context.Context, int64) ([]*models.Room, error) {
	panic("not implemented")
}

type mockBookingServiceConferenceRepo struct {
	listFn func(context.Context) ([]*models.Conference, error)
}

func (m *mockBookingServiceConferenceRepo) Create(context.Context, *models.Conference) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockBookingServiceConferenceRepo) CreateWithOwner(context.Context, *models.Conference, int64) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockBookingServiceConferenceRepo) GetBySlug(context.Context, string) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockBookingServiceConferenceRepo) List(ctx context.Context) ([]*models.Conference, error) {
	return m.listFn(ctx)
}

type mockBookingServiceUserRepo struct {
	getByIDFn func(context.Context, int64) (*models.User, error)
}

func (m *mockBookingServiceUserRepo) Create(context.Context, *models.User) (*models.User, error) {
	panic("not implemented")
}
func (m *mockBookingServiceUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockBookingServiceUserRepo) GetByGitHubID(context.Context, string) (*models.User, error) {
	panic("not implemented")
}
func (m *mockBookingServiceUserRepo) Update(context.Context, *models.User) (*models.User, error) {
	panic("not implemented")
}

func TestBookingService_CancelBooking_Success(t *testing.T) {
	now := time.Now().UTC()
	cancelledAt := now.Add(time.Minute)
	svc := NewBookingService(
		&mockBookingServiceBookingRepo{
			getByIDFn: func(_ context.Context, id int64) (*models.Booking, error) {
				return &models.Booking{ID: id, UserID: 7, RoomID: 11, ConferenceID: 3, Status: models.BookingStatusConfirmed, CreatedAt: now}, nil
			},
			cancelFn: func(_ context.Context, id int64) (*models.Booking, error) {
				return &models.Booking{ID: id, UserID: 7, RoomID: 11, ConferenceID: 3, Status: models.BookingStatusCancelled, CreatedAt: now, CancelledAt: &cancelledAt}, nil
			},
			cancelAndRequestsTxnFn: func(_ context.Context, bookingID, userID int64) (*models.Booking, int64, error) {
				assert.Equal(t, int64(1), bookingID)
				assert.Equal(t, int64(7), userID)
				return &models.Booking{ID: bookingID, UserID: userID, RoomID: 11, ConferenceID: 3, Status: models.BookingStatusCancelled, CreatedAt: now, CancelledAt: &cancelledAt}, 2, nil
			},
			listByRoom: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{ID: 2, UserID: 8, RoomID: 11, PrivacySetting: "public"}}, nil
			},
			listByUser: func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
		},
		&mockBookingServiceRoomRepo{
			getByIDFn: func(_ context.Context, id int64) (*models.Room, error) {
				return &models.Room{ID: id, RoomNumber: "204", RoomType: "double", PricePerNight: 120, Capacity: 2}, nil
			},
		},
		&mockBookingServiceConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 3, Slug: "socrates-26", Name: "SoCraTes", StartDate: now, EndDate: now.Add(24 * time.Hour)}}, nil
			},
		},
		&mockBookingServiceUserRepo{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				return &models.User{ID: id, DisplayName: "Roommate"}, nil
			},
		},
	)

	result, err := svc.CancelBooking(context.Background(), 7, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "cancelled", result.Status)
	assert.Equal(t, "204", result.Room.RoomNumber)
	assert.Equal(t, 2, result.Room.Capacity)
	assert.Equal(t, 1, result.Room.SpotsTaken)
	assert.Equal(t, 1, result.Room.SpotsAvailable)
	assert.Len(t, result.Roommates, 1)
}

func TestBookingService_CancelBooking_Forbidden(t *testing.T) {
	svc := NewBookingService(
		&mockBookingServiceBookingRepo{
			getByIDFn: func(_ context.Context, _ int64) (*models.Booking, error) {
				return &models.Booking{ID: 1, UserID: 42, Status: models.BookingStatusConfirmed}, nil
			},
			cancelFn:               func(_ context.Context, _ int64) (*models.Booking, error) { return nil, nil },
			cancelAndRequestsTxnFn: func(_ context.Context, _, _ int64) (*models.Booking, int64, error) { return nil, 0, nil },
			listByRoom:             func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
			listByUser:             func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
		},
		&mockBookingServiceRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) { return nil, nil }},
		&mockBookingServiceConferenceRepo{listFn: func(_ context.Context) ([]*models.Conference, error) { return nil, nil }},
		&mockBookingServiceUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) { return nil, nil }},
	)

	result, err := svc.CancelBooking(context.Background(), 7, 1)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBookingForbidden)
}

func TestBookingService_CancelBooking_NotFound(t *testing.T) {
	svc := NewBookingService(
		&mockBookingServiceBookingRepo{
			getByIDFn: func(_ context.Context, _ int64) (*models.Booking, error) {
				return nil, repository.ErrBookingNotFound
			},
			cancelFn:               func(_ context.Context, _ int64) (*models.Booking, error) { return nil, nil },
			cancelAndRequestsTxnFn: func(_ context.Context, _, _ int64) (*models.Booking, int64, error) { return nil, 0, nil },
			listByRoom:             func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
			listByUser:             func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
		},
		&mockBookingServiceRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) { return nil, nil }},
		&mockBookingServiceConferenceRepo{listFn: func(_ context.Context) ([]*models.Conference, error) { return nil, nil }},
		&mockBookingServiceUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) { return nil, nil }},
	)

	result, err := svc.CancelBooking(context.Background(), 7, 1)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBookingNotFound)
}

func TestBookingService_CancelBooking_TransactionalFailure(t *testing.T) {
	svc := NewBookingService(
		&mockBookingServiceBookingRepo{
			getByIDFn: func(_ context.Context, id int64) (*models.Booking, error) {
				return &models.Booking{ID: id, UserID: 7, Status: models.BookingStatusConfirmed}, nil
			},
			cancelFn: func(_ context.Context, _ int64) (*models.Booking, error) { return nil, errors.New("unused") },
			cancelAndRequestsTxnFn: func(_ context.Context, _, _ int64) (*models.Booking, int64, error) {
				return nil, 0, errors.New("boom")
			},
			listByRoom: func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
			listByUser: func(_ context.Context, _ int64) ([]*models.Booking, error) { return nil, nil },
		},
		&mockBookingServiceRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) { return nil, nil }},
		&mockBookingServiceConferenceRepo{listFn: func(_ context.Context) ([]*models.Conference, error) { return nil, nil }},
		&mockBookingServiceUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) { return nil, nil }},
	)

	result, err := svc.CancelBooking(context.Background(), 7, 1)
	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cancel booking transactionally")
}

func TestBookingService_ListBookings_UsesActiveBookings(t *testing.T) {
	now := time.Now().UTC()
	svc := NewBookingService(
		&mockBookingServiceBookingRepo{
			listByUser: func(_ context.Context, userID int64) ([]*models.Booking, error) {
				assert.Equal(t, int64(7), userID)
				return []*models.Booking{
					{ID: 1, UserID: 7, RoomID: 11, ConferenceID: 3, Status: models.BookingStatusConfirmed, PrivacySetting: "public", CreatedAt: now},
				}, nil
			},
			listByRoom: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{{ID: 1, UserID: 7}}, nil
			},
			getByIDFn: func(_ context.Context, _ int64) (*models.Booking, error) { return nil, errors.New("unused") },
			cancelFn:  func(_ context.Context, _ int64) (*models.Booking, error) { return nil, errors.New("unused") },
			cancelAndRequestsTxnFn: func(_ context.Context, _, _ int64) (*models.Booking, int64, error) {
				return nil, 0, errors.New("unused")
			},
		},
		&mockBookingServiceRoomRepo{
			getByIDFn: func(_ context.Context, id int64) (*models.Room, error) {
				return &models.Room{ID: id, RoomNumber: "204", RoomType: "double", PricePerNight: 120}, nil
			},
		},
		&mockBookingServiceConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 3, Slug: "socrates-26", Name: "SoCraTes", StartDate: now, EndDate: now.Add(24 * time.Hour)}}, nil
			},
		},
		&mockBookingServiceUserRepo{
			getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
				return &models.User{DisplayName: "ignored"}, nil
			},
		},
	)

	result, err := svc.ListBookings(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "socrates-26", result[0].ConferenceSlug)
}

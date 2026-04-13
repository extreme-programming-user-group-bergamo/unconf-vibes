package sqlite

import (
	"context"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bookingTestFixture struct {
	confRepo    *ConferenceRepository
	roomRepo    *RoomRepository
	bookingRepo *BookingRepository
	userRepo    *UserRepository
	conf        *models.Conference
	room        *models.Room
	user        *models.User
}

func setupBookingFixture(t *testing.T) *bookingTestFixture {
	t.Helper()
	db := setupTestDB(t)

	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)
	bookingRepo := NewBookingRepository(db)
	userRepo := NewUserRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("booking-conf"))
	require.NoError(t, err)

	room, err := roomRepo.Create(context.Background(), newTestRoom(conf.ID, "101"))
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "gh_booking_user",
		Email:       "booker@test.com",
		DisplayName: "Booker",
	})
	require.NoError(t, err)

	return &bookingTestFixture{
		confRepo:    confRepo,
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		conf:        conf,
		room:        room,
		user:        user,
	}
}

func newTestBooking(roomID, userID, conferenceID int64) *models.Booking {
	return &models.Booking{
		RoomID:         roomID,
		UserID:         userID,
		ConferenceID:   conferenceID,
		Status:         models.BookingStatusRequested,
		PrivacySetting: "public",
		Notes:          "test booking",
	}
}

func TestBookingRepository_Create_Success(t *testing.T) {
	f := setupBookingFixture(t)

	booking := newTestBooking(f.room.ID, f.user.ID, f.conf.ID)
	created, err := f.bookingRepo.Create(context.Background(), booking)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, f.room.ID, created.RoomID)
	assert.Equal(t, f.user.ID, created.UserID)
	assert.Equal(t, f.conf.ID, created.ConferenceID)
	assert.Equal(t, models.BookingStatusRequested, created.Status)
	assert.Equal(t, "public", created.PrivacySetting)
	assert.Equal(t, "test booking", created.Notes)
	assert.False(t, created.CreatedAt.IsZero())
	assert.Nil(t, created.ConfirmedAt)
	assert.Nil(t, created.CancelledAt)
}

func TestBookingRepository_Create_DuplicateActiveBooking(t *testing.T) {
	f := setupBookingFixture(t)

	_, err := f.bookingRepo.Create(context.Background(), newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	_, err = f.bookingRepo.Create(context.Background(), newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingExists)
}

func TestBookingRepository_Create_AllowsAfterCancelled(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	// Create and cancel a booking
	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	// Cancel via direct SQL (no UpdateStatus method yet)
	_, err = f.bookingRepo.db.ExecContext(ctx, "UPDATE bookings SET status = 'cancelled' WHERE id = ?", created.ID)
	require.NoError(t, err)

	// Should allow new booking after cancellation
	newBooking, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)
	assert.NotEqual(t, created.ID, newBooking.ID)
}

func TestBookingRepository_GetByID_Success(t *testing.T) {
	f := setupBookingFixture(t)

	created, err := f.bookingRepo.Create(context.Background(), newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	fetched, err := f.bookingRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, f.room.ID, fetched.RoomID)
	assert.Equal(t, f.user.ID, fetched.UserID)
}

func TestBookingRepository_GetByID_NotFound(t *testing.T) {
	f := setupBookingFixture(t)

	_, err := f.bookingRepo.GetByID(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)
}

func TestBookingRepository_ListByRoom_ReturnsNonCancelled(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	_, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByRoom(ctx, f.room.ID)
	require.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestBookingRepository_ListByRoom_ExcludesCancelled(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	// Cancel the booking
	_, err = f.bookingRepo.db.ExecContext(ctx, "UPDATE bookings SET status = 'cancelled' WHERE id = ?", created.ID)
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByRoom(ctx, f.room.ID)
	require.NoError(t, err)
	assert.Empty(t, bookings)
}

func TestBookingRepository_ListByRoom_Empty(t *testing.T) {
	f := setupBookingFixture(t)

	bookings, err := f.bookingRepo.ListByRoom(context.Background(), f.room.ID)
	require.NoError(t, err)
	require.NotNil(t, bookings)
	assert.Empty(t, bookings)
}

func TestBookingRepository_ListByConference_ReturnsNonCancelled(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	_, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByConference(ctx, f.conf.ID)
	require.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestBookingRepository_ListByConference_ExcludesCancelled(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	_, err = f.bookingRepo.db.ExecContext(ctx, "UPDATE bookings SET status = 'cancelled' WHERE id = ?", created.ID)
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByConference(ctx, f.conf.ID)
	require.NoError(t, err)
	assert.Empty(t, bookings)
}

func TestBookingRepository_CountByConference_DistinctUsers(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	// Create a second user
	user2, err := f.userRepo.Create(ctx, &models.User{
		GitHubID:    "gh_user2",
		Email:       "user2@test.com",
		DisplayName: "User Two",
	})
	require.NoError(t, err)

	// Create a second room
	room2, err := f.roomRepo.Create(ctx, newTestRoom(f.conf.ID, "102"))
	require.NoError(t, err)

	// Two users booking different rooms
	_, err = f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)
	_, err = f.bookingRepo.Create(ctx, newTestBooking(room2.ID, user2.ID, f.conf.ID))
	require.NoError(t, err)

	count, err := f.bookingRepo.CountByConference(ctx, f.conf.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestBookingRepository_CountByConference_NoBookings(t *testing.T) {
	f := setupBookingFixture(t)

	count, err := f.bookingRepo.CountByConference(context.Background(), f.conf.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestBookingRepository_GetActiveByUserAndConference_Success(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	found, err := f.bookingRepo.GetActiveByUserAndConference(ctx, f.user.ID, f.conf.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestBookingRepository_GetActiveByUserAndConference_NotFound(t *testing.T) {
	f := setupBookingFixture(t)

	_, err := f.bookingRepo.GetActiveByUserAndConference(context.Background(), f.user.ID, f.conf.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)
}

func TestBookingRepository_UpdateRoom_Success(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	room2, err := f.roomRepo.Create(ctx, newTestRoom(f.conf.ID, "103"))
	require.NoError(t, err)

	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	updated, err := f.bookingRepo.UpdateRoom(ctx, created.ID, room2.ID)
	require.NoError(t, err)
	assert.Equal(t, room2.ID, updated.RoomID)
	assert.Equal(t, created.ID, updated.ID)
}

func TestBookingRepository_UpdateRoom_NotFound(t *testing.T) {
	f := setupBookingFixture(t)

	_, err := f.bookingRepo.UpdateRoom(context.Background(), 9999, f.room.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)
}

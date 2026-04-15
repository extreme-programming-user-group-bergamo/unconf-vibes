package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bookingTestFixture struct {
	db          *sql.DB
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
		db:          db,
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

func TestBookingRepository_ListByConferenceIncludingCancelled_ReturnsCancelledAndActive(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	activeBooking, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	user2, err := f.userRepo.Create(ctx, &models.User{
		GitHubID:    "gh_booking_cancelled_include",
		Email:       "cancelled-include@test.com",
		DisplayName: "Cancelled Include",
	})
	require.NoError(t, err)
	room2, err := f.roomRepo.Create(ctx, newTestRoom(f.conf.ID, "108"))
	require.NoError(t, err)
	cancelledBooking, err := f.bookingRepo.Create(ctx, newTestBooking(room2.ID, user2.ID, f.conf.ID))
	require.NoError(t, err)
	_, err = f.bookingRepo.db.ExecContext(ctx, "UPDATE bookings SET status = 'cancelled' WHERE id = ?", cancelledBooking.ID)
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByConferenceIncludingCancelled(ctx, f.conf.ID)
	require.NoError(t, err)
	require.Len(t, bookings, 2)

	foundActive := false
	foundCancelled := false
	for i := range bookings {
		if bookings[i].ID == activeBooking.ID {
			foundActive = true
		}
		if bookings[i].ID == cancelledBooking.ID {
			foundCancelled = true
			assert.Equal(t, models.BookingStatusCancelled, bookings[i].Status)
		}
	}
	assert.True(t, foundActive)
	assert.True(t, foundCancelled)
}

func TestBookingRepository_ListByUser_ReturnsActiveBookings(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	_, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	bookings, err := f.bookingRepo.ListByUser(ctx, f.user.ID)
	require.NoError(t, err)
	require.Len(t, bookings, 1)
	assert.Equal(t, f.user.ID, bookings[0].UserID)
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

func TestBookingRepository_Cancel_Success(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()

	created, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	cancelled, err := f.bookingRepo.Cancel(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.BookingStatusCancelled, cancelled.Status)
	require.NotNil(t, cancelled.CancelledAt)
}

func TestBookingRepository_Cancel_NotFound(t *testing.T) {
	f := setupBookingFixture(t)

	_, err := f.bookingRepo.Cancel(context.Background(), 9999)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)
}

func TestBookingRepository_CancelAndCancelPendingOutgoingRequests_Success(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()
	requestRepo := NewRoommateRequestRepository(f.bookingRepo.db)

	targetUser, err := f.userRepo.Create(ctx, &models.User{
		GitHubID:    "gh_booking_cancel_target",
		Email:       "booking-cancel-target@test.com",
		DisplayName: "Cancel Target",
	})
	require.NoError(t, err)

	otherConference, err := f.confRepo.Create(ctx, newTestConference("booking-cancel-other-conf"))
	require.NoError(t, err)
	otherRoom, err := f.roomRepo.Create(ctx, newTestRoom(otherConference.ID, "109"))
	require.NoError(t, err)

	createdBooking, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	targetConferencePending, err := requestRepo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.user.ID,
		TargetID:    targetUser.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	otherConferencePending, err := requestRepo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.user.ID,
		TargetID:    targetUser.ID,
		RoomID:      otherRoom.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	cancelledBooking, cancelledCount, err := f.bookingRepo.CancelAndCancelPendingOutgoingRequests(ctx, createdBooking.ID, f.user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cancelledCount)
	assert.Equal(t, models.BookingStatusCancelled, cancelledBooking.Status)
	require.NotNil(t, cancelledBooking.CancelledAt)

	updatedTargetConferenceRequest, err := requestRepo.GetByID(ctx, targetConferencePending.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusCancelled, updatedTargetConferenceRequest.Status)

	updatedOtherConferenceRequest, err := requestRepo.GetByID(ctx, otherConferencePending.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusPending, updatedOtherConferenceRequest.Status)
}

func TestBookingRepository_CancelAndCancelPendingOutgoingRequests_RollsBackOnRequestUpdateFailure(t *testing.T) {
	f := setupBookingFixture(t)
	ctx := context.Background()
	requestRepo := NewRoommateRequestRepository(f.bookingRepo.db)

	targetUser, err := f.userRepo.Create(ctx, &models.User{
		GitHubID:    "gh_booking_cancel_rollback_target",
		Email:       "booking-cancel-rollback-target@test.com",
		DisplayName: "Rollback Target",
	})
	require.NoError(t, err)

	createdBooking, err := f.bookingRepo.Create(ctx, newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	pendingRequest, err := requestRepo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.user.ID,
		TargetID:    targetUser.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	_, err = f.bookingRepo.db.ExecContext(ctx, `
		CREATE TRIGGER trg_block_roommate_request_cancel
		BEFORE UPDATE OF status ON roommate_requests
		WHEN NEW.status = 'cancelled'
		BEGIN
			SELECT RAISE(ABORT, 'blocked roommate request cancellation');
		END;
	`)
	require.NoError(t, err)

	_, _, err = f.bookingRepo.CancelAndCancelPendingOutgoingRequests(ctx, createdBooking.ID, f.user.ID)
	require.Error(t, err)

	bookingAfterFailure, err := f.bookingRepo.GetByID(ctx, createdBooking.ID)
	require.NoError(t, err)
	assert.Equal(t, models.BookingStatusRequested, bookingAfterFailure.Status)
	assert.Nil(t, bookingAfterFailure.CancelledAt)

	requestAfterFailure, err := requestRepo.GetByID(ctx, pendingRequest.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusPending, requestAfterFailure.Status)
}

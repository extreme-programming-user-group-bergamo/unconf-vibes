package sqlite

import (
	"context"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailLogRepository_CreateAndListByBookingID(t *testing.T) {
	f := setupBookingFixture(t)
	repo := NewEmailLogRepository(f.db)
	booking, err := f.bookingRepo.Create(context.Background(), newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	created, err := repo.Create(context.Background(), &models.EmailLog{
		BookingID:    booking.ID,
		EmailType:    "booking_created",
		Recipient:    "hotel@example.com",
		Subject:      "New Booking Notification",
		Status:       models.EmailLogStatusSent,
		Attempt:      1,
		ErrorDetails: "",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, models.EmailLogStatusSent, created.Status)

	logs, err := repo.ListByBookingID(context.Background(), booking.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "booking_created", logs[0].EmailType)
	assert.Equal(t, "hotel@example.com", logs[0].Recipient)
	assert.Equal(t, "New Booking Notification", logs[0].Subject)
	assert.Equal(t, 1, logs[0].Attempt)
}

func TestEmailLogRepository_Create_FailedAttempt(t *testing.T) {
	f := setupBookingFixture(t)
	repo := NewEmailLogRepository(f.db)
	booking, err := f.bookingRepo.Create(context.Background(), newTestBooking(f.room.ID, f.user.ID, f.conf.ID))
	require.NoError(t, err)

	created, err := repo.Create(context.Background(), &models.EmailLog{
		BookingID:    booking.ID,
		EmailType:    "booking_cancelled",
		Recipient:    "hotel@example.com",
		Subject:      "Booking Cancellation Notification",
		Status:       models.EmailLogStatusFailed,
		Attempt:      2,
		ErrorDetails: "smtp timeout",
	})
	require.NoError(t, err)
	assert.Equal(t, models.EmailLogStatusFailed, created.Status)
	assert.Equal(t, "smtp timeout", created.ErrorDetails)
}

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

func TestAttendeeService_ListAttendees_Success(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 2)

	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, slug string) (*models.Conference, error) {
				assert.Equal(t, "socrates-26", slug)
				return conf, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, conferenceID int64) ([]*models.Booking, error) {
				assert.Equal(t, conf.ID, conferenceID)
				return []*models.Booking{
					testBooking(1, room.ID, 10, conf.ID, "public"),
				}, nil
			},
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, conferenceID int64) ([]*models.Room, error) {
				assert.Equal(t, conf.ID, conferenceID)
				return []*models.Room{room}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				assert.Equal(t, int64(10), id)
				return &models.User{
					ID:             10,
					GitHubID:       "alice",
					DisplayName:    "Alice",
					PrivacySetting: "public",
				}, nil
			},
		},
	)

	result, err := svc.ListAttendees(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Attendees, 1)
	assert.Equal(t, "Alice", result.Attendees[0].DisplayName)
	assert.Equal(t, "alice", result.Attendees[0].GitHubUsername)
	require.NotNil(t, result.Attendees[0].Room)
	assert.Equal(t, "101", result.Attendees[0].Room.RoomNumber)
	assert.Equal(t, 0, result.PrivateAttendeesCount)
}

func TestAttendeeService_ListAttendees_PrivateCountsOnly(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 2)

	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
				return conf, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{
					testBooking(1, room.ID, 10, conf.ID, "private"),
					testBooking(2, room.ID, 20, conf.ID, "public"),
				}, nil
			},
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				if id == 20 {
					return &models.User{
						ID:             20,
						GitHubID:       "bob",
						DisplayName:    "Bob",
						PrivacySetting: "private",
					}, nil
				}
				return nil, repository.ErrUserNotFound
			},
		},
	)

	result, err := svc.ListAttendees(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Attendees)
	assert.Equal(t, 2, result.PrivateAttendeesCount)
}

func TestAttendeeService_ListAttendees_ConferenceNotFound(t *testing.T) {
	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
				return nil, repository.ErrConferenceNotFound
			},
		},
		&mockBookingRepository{},
		&mockRoomRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListAttendees(context.Background(), "missing-conf")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}

func TestAttendeeService_ListAttendees_BookingRepoError(t *testing.T) {
	conf := testConference()
	repoErr := errors.New("booking repo failure")

	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
				return conf, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return nil, repoErr
			},
		},
		&mockRoomRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListAttendees(context.Background(), "socrates-26")
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

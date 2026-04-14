package service

import (
	"context"
	"errors"
	"strings"
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
		&mockOrganizerRepository{
			isOrganizerFn: func(_ context.Context, conferenceID int64, userID int64) (bool, error) {
				assert.Equal(t, conf.ID, conferenceID)
				assert.Equal(t, int64(50), userID)
				return false, nil
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

	result, err := svc.ListAttendees(context.Background(), "socrates-26", 50)
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
		&mockOrganizerRepository{
			isOrganizerFn: func(_ context.Context, _ int64, _ int64) (bool, error) {
				return false, nil
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

	result, err := svc.ListAttendees(context.Background(), "socrates-26", 50)
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
		&mockOrganizerRepository{},
		&mockBookingRepository{},
		&mockRoomRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListAttendees(context.Background(), "missing-conf", 1)
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
		&mockOrganizerRepository{},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return nil, repoErr
			},
		},
		&mockRoomRepository{},
		&mockUserRepository{},
	)

	_, err := svc.ListAttendees(context.Background(), "socrates-26", 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestAttendeeService_ListAttendees_OrganizerSeesPrivateDetails(t *testing.T) {
	conf := testConference()
	room := testRoom(1, conf.ID, "101", 2)

	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
				return conf, nil
			},
		},
		&mockOrganizerRepository{
			isOrganizerFn: func(_ context.Context, _ int64, _ int64) (bool, error) {
				return true, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				return []*models.Booking{
					testBooking(1, room.ID, 10, conf.ID, "private"),
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
				return &models.User{
					ID:             id,
					GitHubID:       "secret-user",
					DisplayName:    "Secret User",
					PrivacySetting: "private",
				}, nil
			},
		},
	)

	result, err := svc.ListAttendees(context.Background(), "socrates-26", 99)
	require.NoError(t, err)
	require.Len(t, result.Attendees, 1)
	assert.Equal(t, "Secret User", result.Attendees[0].DisplayName)
	assert.Equal(t, 0, result.PrivateAttendeesCount)
}

func TestAttendeeService_GetOrganizerDashboard_ComputesMetricsAndFilters(t *testing.T) {
	conf := testConference()
	room1 := testRoom(1, conf.ID, "101", 2)
	room1.RoomType = "double"
	room2 := testRoom(2, conf.ID, "102", 1)
	room2.RoomType = "single"

	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockOrganizerRepository{
			isOrganizerFn: func(_ context.Context, conferenceID int64, userID int64) (bool, error) {
				assert.Equal(t, conf.ID, conferenceID)
				assert.Equal(t, int64(77), userID)
				return true, nil
			},
		},
		&mockBookingRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Booking, error) {
				b1 := testBooking(1, room1.ID, 10, conf.ID, "private")
				b1.Notes = "Vegan"
				b1.Status = models.BookingStatusConfirmed
				b2 := testBooking(2, room2.ID, 20, conf.ID, "public")
				b2.Status = models.BookingStatusRequested
				return []*models.Booking{b1, b2}, nil
			},
		},
		&mockRoomRepository{
			listByConferenceFn: func(_ context.Context, _ int64) ([]*models.Room, error) {
				return []*models.Room{room1, room2}, nil
			},
		},
		&mockUserRepository{
			getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
				if id == 10 {
					return &models.User{ID: 10, DisplayName: "Alice", Email: "alice@test.dev", PrivacySetting: "private"}, nil
				}
				return &models.User{ID: 20, DisplayName: "Bob", Email: "bob@test.dev", PrivacySetting: "public"}, nil
			},
		},
	)

	result, err := svc.GetOrganizerDashboard(context.Background(), conf.Slug, 77, OrganizerDashboardFilters{
		RoomType:           "double",
		BookingStatus:      "confirmed",
		HasSpecialRequests: true,
		Search:             "ali",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRegistrations)
	assert.Equal(t, toPercentage(2, conf.Capacity), result.CapacityUsagePct)
	require.Len(t, result.RoomFillRates, 2)
	require.Len(t, result.Attendees, 1)
	assert.Equal(t, "Alice", result.Attendees[0].Name)
	assert.Equal(t, "alice@test.dev", result.Attendees[0].Email)
	assert.Equal(t, "private", result.Attendees[0].PrivacySetting)
	assert.Equal(t, "vegan", strings.ToLower(result.Attendees[0].DietaryAccessibilityNote))
}

func TestAttendeeService_GetOrganizerDashboard_NonOrganizerForbidden(t *testing.T) {
	conf := testConference()
	svc := NewAttendeeService(
		&mockConferenceRepository{
			getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) { return conf, nil },
		},
		&mockOrganizerRepository{
			isOrganizerFn: func(_ context.Context, _ int64, _ int64) (bool, error) { return false, nil },
		},
		&mockBookingRepository{},
		&mockRoomRepository{},
		&mockUserRepository{},
	)

	_, err := svc.GetOrganizerDashboard(context.Background(), conf.Slug, 11, OrganizerDashboardFilters{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOrganizerForbidden)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/email"
	"github.com/katurdays/unconf/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHotelEmailSender struct {
	sendFn func(context.Context, email.Message) error
}

func (m *mockHotelEmailSender) Send(ctx context.Context, message email.Message) error {
	if m.sendFn == nil {
		return nil
	}
	return m.sendFn(ctx, message)
}

type mockHotelConferenceRepo struct {
	listFn func(context.Context) ([]*models.Conference, error)
}

func (m *mockHotelConferenceRepo) Create(context.Context, *models.Conference) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockHotelConferenceRepo) CreateWithOwner(context.Context, *models.Conference, int64) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockHotelConferenceRepo) GetBySlug(context.Context, string) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockHotelConferenceRepo) UpdateBySlug(context.Context, string, *models.Conference) (*models.Conference, error) {
	panic("not implemented")
}
func (m *mockHotelConferenceRepo) List(ctx context.Context) ([]*models.Conference, error) {
	return m.listFn(ctx)
}

type mockHotelRoomRepo struct {
	getByIDFn func(context.Context, int64) (*models.Room, error)
}

func (m *mockHotelRoomRepo) Create(context.Context, *models.Room) (*models.Room, error) {
	panic("not implemented")
}
func (m *mockHotelRoomRepo) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockHotelRoomRepo) GetByConferenceAndNumber(context.Context, int64, string) (*models.Room, error) {
	panic("not implemented")
}
func (m *mockHotelRoomRepo) ListByConference(context.Context, int64) ([]*models.Room, error) {
	panic("not implemented")
}
func (m *mockHotelRoomRepo) UpdateByConferenceAndNumber(context.Context, int64, string, *models.Room) (*models.Room, error) {
	panic("not implemented")
}
func (m *mockHotelRoomRepo) DeleteByConferenceAndNumber(context.Context, int64, string) error {
	panic("not implemented")
}

type mockHotelUserRepo struct {
	getByIDFn func(context.Context, int64) (*models.User, error)
}

func (m *mockHotelUserRepo) Create(context.Context, *models.User) (*models.User, error) {
	panic("not implemented")
}
func (m *mockHotelUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockHotelUserRepo) GetByGitHubID(context.Context, string) (*models.User, error) {
	panic("not implemented")
}
func (m *mockHotelUserRepo) Update(context.Context, *models.User) (*models.User, error) {
	panic("not implemented")
}

type mockHotelOrganizerRepo struct {
	emailFn func(context.Context, int64) ([]string, error)
}

func (m *mockHotelOrganizerRepo) Add(context.Context, *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error) {
	panic("not implemented")
}
func (m *mockHotelOrganizerRepo) GetByConferenceAndUser(context.Context, int64, int64) (*models.ConferenceOrganizer, error) {
	panic("not implemented")
}
func (m *mockHotelOrganizerRepo) ListEmailsByConference(ctx context.Context, conferenceID int64) ([]string, error) {
	return m.emailFn(ctx, conferenceID)
}
func (m *mockHotelOrganizerRepo) RemoveByConferenceAndUser(context.Context, int64, int64) error {
	panic("not implemented")
}
func (m *mockHotelOrganizerRepo) IsOrganizer(context.Context, int64, int64) (bool, error) {
	panic("not implemented")
}
func (m *mockHotelOrganizerRepo) IsOrganizerForAnyConference(context.Context, int64) (bool, error) {
	panic("not implemented")
}

type mockHotelEmailLogRepo struct {
	createFn func(context.Context, *models.EmailLog) (*models.EmailLog, error)
	logs     []*models.EmailLog
}

func (m *mockHotelEmailLogRepo) Create(ctx context.Context, log *models.EmailLog) (*models.EmailLog, error) {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	m.logs = append(m.logs, log)
	return log, nil
}

func (m *mockHotelEmailLogRepo) ListByBookingID(context.Context, int64) ([]*models.EmailLog, error) {
	return m.logs, nil
}

func TestHotelEmailService_NotifyBookingCreated_SendsWithOrganizerBCCAndLogs(t *testing.T) {
	var sentMessage email.Message
	renderer := email.NewTemplateRenderer()
	sender := &mockHotelEmailSender{
		sendFn: func(_ context.Context, message email.Message) error {
			sentMessage = message
			return nil
		},
	}
	emailSvc, err := email.NewService(renderer, sender, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	logRepo := &mockHotelEmailLogRepo{}
	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "hotel@example.com", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com", "admin@example.com"}, nil
		}},
		logRepo,
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCreated(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
		Notes:        "Late arrival",
	})
	require.NoError(t, err)
	assert.Equal(t, "hotel@example.com", sentMessage.To[0].Email)
	assert.Len(t, sentMessage.BCC, 2)
	assert.Equal(t, "owner@example.com", sentMessage.BCC[0].Email)
	assert.NotContains(t, sentMessage.Body, "BCC:")
	require.Len(t, logRepo.logs, 1)
	assert.Equal(t, models.EmailLogStatusSent, logRepo.logs[0].Status)
	assert.Equal(t, 1, logRepo.logs[0].Attempt)
}

func TestHotelEmailService_NotifyBookingCancelled_RetriesAndLogsFailures(t *testing.T) {
	attempt := 0
	renderer := email.NewTemplateRenderer()
	sender := &mockHotelEmailSender{
		sendFn: func(_ context.Context, _ email.Message) error {
			attempt++
			if attempt < 3 {
				return errors.New("transient smtp timeout")
			}
			return nil
		},
	}
	emailSvc, err := email.NewService(renderer, sender, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	logRepo := &mockHotelEmailLogRepo{}
	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "hotel@example.com", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com"}, nil
		}},
		logRepo,
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCancelled(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
	})
	require.NoError(t, err)
	require.Len(t, logRepo.logs, 3)
	assert.Equal(t, models.EmailLogStatusFailed, logRepo.logs[0].Status)
	assert.Equal(t, models.EmailLogStatusFailed, logRepo.logs[1].Status)
	assert.Equal(t, models.EmailLogStatusSent, logRepo.logs[2].Status)
	assert.Equal(t, 3, logRepo.logs[2].Attempt)
}

func TestHotelEmailService_NotifyBookingCancelled_ContinuesRetriesWhenFailedAttemptLogWriteFails(t *testing.T) {
	attempt := 0
	renderer := email.NewTemplateRenderer()
	sender := &mockHotelEmailSender{
		sendFn: func(_ context.Context, _ email.Message) error {
			attempt++
			if attempt < 3 {
				return errors.New("transient smtp timeout")
			}
			return nil
		},
	}
	emailSvc, err := email.NewService(renderer, sender, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	logRepo := &mockHotelEmailLogRepo{}
	logRepo.createFn = func(_ context.Context, log *models.EmailLog) (*models.EmailLog, error) {
		if log.Status == models.EmailLogStatusFailed {
			return nil, errors.New("db unavailable")
		}
		logRepo.logs = append(logRepo.logs, log)
		return log, nil
	}

	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "hotel@example.com", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com"}, nil
		}},
		logRepo,
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCancelled(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
	})
	require.NoError(t, err)
	require.Len(t, logRepo.logs, 1)
	assert.Equal(t, models.EmailLogStatusSent, logRepo.logs[0].Status)
	assert.Equal(t, 3, logRepo.logs[0].Attempt)
}

func TestHotelEmailService_NotifyBookingCreated_SucceedsWhenSuccessfulAttemptLogWriteFails(t *testing.T) {
	renderer := email.NewTemplateRenderer()
	sender := &mockHotelEmailSender{
		sendFn: func(_ context.Context, _ email.Message) error {
			return nil
		},
	}
	emailSvc, err := email.NewService(renderer, sender, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	logRepo := &mockHotelEmailLogRepo{
		createFn: func(_ context.Context, log *models.EmailLog) (*models.EmailLog, error) {
			if log.Status == models.EmailLogStatusSent {
				return nil, errors.New("db unavailable")
			}
			return log, nil
		},
	}

	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "hotel@example.com", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com"}, nil
		}},
		logRepo,
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCreated(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
	})
	require.NoError(t, err)
}

func TestHotelEmailService_NotifyBookingCancelled_ReturnsDeliveryFailedAfterMaxAttempts(t *testing.T) {
	renderer := email.NewTemplateRenderer()
	sender := &mockHotelEmailSender{
		sendFn: func(_ context.Context, _ email.Message) error {
			return errors.New("smtp down")
		},
	}
	emailSvc, err := email.NewService(renderer, sender, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	logRepo := &mockHotelEmailLogRepo{}
	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "hotel@example.com", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com"}, nil
		}},
		logRepo,
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCancelled(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHotelEmailDeliveryFailed)
	require.Len(t, logRepo.logs, 3)
	assert.Equal(t, models.EmailLogStatusFailed, logRepo.logs[0].Status)
	assert.Equal(t, 1, logRepo.logs[0].Attempt)
	assert.Equal(t, models.EmailLogStatusFailed, logRepo.logs[1].Status)
	assert.Equal(t, 2, logRepo.logs[1].Attempt)
	assert.Equal(t, models.EmailLogStatusFailed, logRepo.logs[2].Status)
	assert.Equal(t, 3, logRepo.logs[2].Attempt)
}

func TestHotelEmailService_NotifyBookingCreated_FailsWhenHotelEmailMissing(t *testing.T) {
	renderer := email.NewTemplateRenderer()
	emailSvc, err := email.NewService(renderer, &mockHotelEmailSender{}, email.Address{Email: "noreply@example.com"})
	require.NoError(t, err)

	service, err := NewHotelEmailService(
		emailSvc,
		&mockHotelConferenceRepo{
			listFn: func(_ context.Context) ([]*models.Conference, error) {
				return []*models.Conference{{ID: 44, HotelEmail: "", StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)}}, nil
			},
		},
		&mockHotelRoomRepo{getByIDFn: func(_ context.Context, _ int64) (*models.Room, error) {
			return &models.Room{ID: 33, RoomNumber: "101"}, nil
		}},
		&mockHotelUserRepo{getByIDFn: func(_ context.Context, _ int64) (*models.User, error) {
			return &models.User{ID: 22, DisplayName: "Guest", Email: "guest@example.com"}, nil
		}},
		&mockHotelOrganizerRepo{emailFn: func(_ context.Context, _ int64) ([]string, error) {
			return []string{"owner@example.com"}, nil
		}},
		&mockHotelEmailLogRepo{},
		3,
		0,
	)
	require.NoError(t, err)

	err = service.NotifyBookingCreated(context.Background(), &models.Booking{
		ID:           99,
		ConferenceID: 44,
		RoomID:       33,
		UserID:       22,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHotelEmailNotConfigured)
}

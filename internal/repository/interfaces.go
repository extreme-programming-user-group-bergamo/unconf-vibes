package repository

import (
	"context"

	"github.com/katurdays/unconf/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
}

type RefreshSessionRepository interface {
	Create(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
	Rotate(ctx context.Context, currentSessionID int64, replacement *models.RefreshSession) (*models.RefreshSession, error)
	RevokeByID(ctx context.Context, sessionID int64) error
}

type ConferenceRepository interface {
	Create(ctx context.Context, conf *models.Conference) (*models.Conference, error)
	CreateWithOwner(ctx context.Context, conf *models.Conference, ownerUserID int64) (*models.Conference, error)
	GetBySlug(ctx context.Context, slug string) (*models.Conference, error)
	UpdateBySlug(ctx context.Context, slug string, conf *models.Conference) (*models.Conference, error)
	List(ctx context.Context) ([]*models.Conference, error)
}

type ConferenceOrganizerRepository interface {
	Add(ctx context.Context, organizer *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error)
	GetByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error)
	RemoveByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) error
	IsOrganizer(ctx context.Context, conferenceID int64, userID int64) (bool, error)
	IsOrganizerForAnyConference(ctx context.Context, userID int64) (bool, error)
}

type RoomRepository interface {
	Create(ctx context.Context, room *models.Room) (*models.Room, error)
	GetByID(ctx context.Context, id int64) (*models.Room, error)
	GetByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) (*models.Room, error)
	ListByConference(ctx context.Context, conferenceID int64) ([]*models.Room, error)
	UpdateByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string, room *models.Room) (*models.Room, error)
	DeleteByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) error
}

type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetByID(ctx context.Context, id int64) (*models.Booking, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error)
	ListByConference(ctx context.Context, conferenceID int64) ([]*models.Booking, error)
	CountByConference(ctx context.Context, conferenceID int64) (int, error)
}

type RoommateRequestRepository interface {
	Create(ctx context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error)
	GetByID(ctx context.Context, id int64) (*models.RoommateRequest, error)
	ListByUser(ctx context.Context, userID int64) ([]*models.RoommateRequest, error)
	UpdateStatus(ctx context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error)
}

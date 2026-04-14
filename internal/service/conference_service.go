package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

// ConferenceResponse wraps a conference with computed fields.
type ConferenceResponse struct {
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Location      string `json:"location"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Capacity      int    `json:"capacity"`
	AttendeeCount int    `json:"attendee_count"`
	Status        string `json:"status"`
}

type CreateConferenceInput struct {
	Slug        string
	Name        string
	Description string
	Location    string
	StartDate   time.Time
	EndDate     time.Time
	Capacity    int
	HotelEmail  string
}

// ConferenceService handles conference business logic.
type ConferenceService struct {
	confRepo    repository.ConferenceRepository
	bookingRepo repository.BookingRepository
}

// NewConferenceService creates a new ConferenceService with the given repositories.
func NewConferenceService(confRepo repository.ConferenceRepository, bookingRepo repository.BookingRepository) *ConferenceService {
	return &ConferenceService{confRepo: confRepo, bookingRepo: bookingRepo}
}

// ListConferences returns all conferences with computed status and attendee count.
func (s *ConferenceService) ListConferences(ctx context.Context) ([]*ConferenceResponse, error) {
	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list conferences: %w", err)
	}

	now := time.Now()
	responses := make([]*ConferenceResponse, 0, len(conferences))
	for _, conf := range conferences {
		attendeeCount, countErr := s.bookingRepo.CountByConference(ctx, conf.ID)
		if countErr != nil {
			slog.Error("failed to count attendees for conference", "error", countErr, "conference_id", conf.ID)
			attendeeCount = 0
		}
		responses = append(responses, toResponseAt(conf, now, attendeeCount))
	}

	return responses, nil
}

// GetConference returns a single conference by slug with computed status and attendee count.
func (s *ConferenceService) GetConference(ctx context.Context, slug string) (*ConferenceResponse, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to get conference: %w", ErrConferenceNotFound)
		}

		return nil, fmt.Errorf("failed to get conference: %w", err)
	}

	attendeeCount, countErr := s.bookingRepo.CountByConference(ctx, conf.ID)
	if countErr != nil {
		slog.Error("failed to count attendees for conference", "error", countErr, "conference_id", conf.ID)
		attendeeCount = 0
	}

	return toResponseAt(conf, time.Now(), attendeeCount), nil
}

// CreateConference creates a conference and assigns the creator as organizer owner.
func (s *ConferenceService) CreateConference(ctx context.Context, creatorUserID int64, input CreateConferenceInput) (*ConferenceResponse, error) {
	if creatorUserID <= 0 {
		return nil, fmt.Errorf("failed to create conference: invalid creator user id")
	}
	if strings.TrimSpace(input.Slug) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Location) == "" {
		return nil, fmt.Errorf("failed to create conference: missing required fields")
	}
	if input.Capacity <= 0 {
		return nil, fmt.Errorf("failed to create conference: capacity must be positive")
	}
	if truncateToDate(input.EndDate).Before(truncateToDate(input.StartDate)) {
		return nil, fmt.Errorf("failed to create conference: end date cannot be before start date")
	}

	created, err := s.confRepo.CreateWithOwner(ctx, &models.Conference{
		Slug:        strings.TrimSpace(input.Slug),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Location:    strings.TrimSpace(input.Location),
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Capacity:    input.Capacity,
		HotelEmail:  strings.TrimSpace(input.HotelEmail),
	}, creatorUserID)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceExists) {
			return nil, fmt.Errorf("failed to create conference: %w", ErrConferenceExists)
		}
		return nil, fmt.Errorf("failed to create conference: %w", err)
	}

	return toResponseAt(created, time.Now(), 0), nil
}

// DeriveStatus computes the conference status from its dates using the current time.
func (s *ConferenceService) DeriveStatus(conf *models.Conference) string {
	return DeriveStatusAt(conf, time.Now())
}

// DeriveStatusAt computes the conference status from its dates relative to the given time.
// Dates are compared as date-only (normalized to midnight UTC) so that
// end_date is treated as inclusive (the conference is "active" through the entire end date).
func DeriveStatusAt(conf *models.Conference, now time.Time) string {
	today := truncateToDate(now)
	startDate := truncateToDate(conf.StartDate)
	endDate := truncateToDate(conf.EndDate)

	if startDate.After(today) {
		return string(models.ConferenceStatusUpcoming)
	}

	if endDate.Before(today) {
		return string(models.ConferenceStatusPast)
	}

	return string(models.ConferenceStatusActive)
}

// truncateToDate normalizes a time to midnight UTC for date-only comparisons.
// Converts to UTC first to avoid timezone-dependent off-by-one errors.
func truncateToDate(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func toResponseAt(conf *models.Conference, now time.Time, attendeeCount int) *ConferenceResponse {
	return &ConferenceResponse{
		ID:            conf.ID,
		Slug:          conf.Slug,
		Name:          conf.Name,
		Description:   conf.Description,
		Location:      conf.Location,
		StartDate:     conf.StartDate.Format("2006-01-02"),
		EndDate:       conf.EndDate.Format("2006-01-02"),
		Capacity:      conf.Capacity,
		AttendeeCount: attendeeCount,
		Status:        DeriveStatusAt(conf, now),
	}
}

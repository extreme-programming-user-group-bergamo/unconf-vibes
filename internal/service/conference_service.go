package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

var (
	conferenceSlugPattern  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	conferenceSlugCollapse = regexp.MustCompile(`-+`)
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

type UpdateConferenceInput struct {
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
	confRepo      repository.ConferenceRepository
	bookingRepo   repository.BookingRepository
	organizerRepo repository.ConferenceOrganizerRepository
}

// NewConferenceService creates a new ConferenceService with the given repositories.
func NewConferenceService(
	confRepo repository.ConferenceRepository,
	bookingRepo repository.BookingRepository,
	organizerRepo repository.ConferenceOrganizerRepository,
) *ConferenceService {
	return &ConferenceService{
		confRepo:      confRepo,
		bookingRepo:   bookingRepo,
		organizerRepo: organizerRepo,
	}
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

	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create conference: %w", err)
	}

	isOrganizer, err := s.organizerRepo.IsOrganizerForAnyConference(ctx, creatorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conference: %w", err)
	}
	if !isOrganizer && len(conferences) > 0 {
		return nil, fmt.Errorf("failed to create conference: %w", ErrOrganizerForbidden)
	}

	slug := normalizeConferenceSlug(input.Slug)
	if !conferenceSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("failed to create conference: %w", ErrInvalidConferenceInput)
	}

	if err := validateConferenceFields(strings.TrimSpace(input.Name), strings.TrimSpace(input.Location), input.StartDate, input.EndDate, input.Capacity); err != nil {
		return nil, fmt.Errorf("failed to create conference: %w", err)
	}

	created, err := s.confRepo.CreateWithOwner(ctx, &models.Conference{
		Slug:        slug,
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

func normalizeConferenceSlug(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = conferenceSlugCollapse.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")

	return normalized
}

// UpdateConference updates conference details by slug while preserving slug immutability.
func (s *ConferenceService) UpdateConference(ctx context.Context, slug string, input UpdateConferenceInput) (*ConferenceResponse, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, fmt.Errorf("failed to update conference: missing conference slug")
	}

	if err := validateConferenceFields(strings.TrimSpace(input.Name), strings.TrimSpace(input.Location), input.StartDate, input.EndDate, input.Capacity); err != nil {
		return nil, fmt.Errorf("failed to update conference: %w", err)
	}

	updated, err := s.confRepo.UpdateBySlug(ctx, slug, &models.Conference{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Location:    strings.TrimSpace(input.Location),
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Capacity:    input.Capacity,
		HotelEmail:  strings.TrimSpace(input.HotelEmail),
	})
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to update conference: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to update conference: %w", err)
	}

	attendeeCount, countErr := s.bookingRepo.CountByConference(ctx, updated.ID)
	if countErr != nil {
		slog.Error("failed to count attendees for conference", "error", countErr, "conference_id", updated.ID)
		attendeeCount = 0
	}

	return toResponseAt(updated, time.Now(), attendeeCount), nil
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

func validateConferenceFields(name, location string, startDate, endDate time.Time, capacity int) error {
	if name == "" || location == "" {
		return fmt.Errorf("%w: missing required fields", ErrInvalidConferenceInput)
	}
	if capacity <= 0 {
		return fmt.Errorf("%w: capacity must be positive", ErrInvalidConferenceInput)
	}
	if truncateToDate(endDate).Before(truncateToDate(startDate)) {
		return fmt.Errorf("%w: end date cannot be before start date", ErrInvalidConferenceInput)
	}

	return nil
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

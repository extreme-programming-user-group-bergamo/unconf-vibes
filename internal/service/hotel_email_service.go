package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/katurdays/unconf/internal/email"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

const (
	defaultHotelEmailRetryAttempts = 3
)

type HotelEmailNotifier interface {
	NotifyBookingCreated(ctx context.Context, booking *models.Booking) error
	NotifyBookingCancelled(ctx context.Context, booking *models.Booking) error
}

type HotelEmailService struct {
	emailService *email.Service

	confRepo      repository.ConferenceRepository
	roomRepo      repository.RoomRepository
	userRepo      repository.UserRepository
	organizerRepo repository.ConferenceOrganizerRepository
	emailLogRepo  repository.EmailLogRepository

	maxAttempts int
	retryDelay  time.Duration
}

func NewHotelEmailService(
	emailService *email.Service,
	confRepo repository.ConferenceRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	organizerRepo repository.ConferenceOrganizerRepository,
	emailLogRepo repository.EmailLogRepository,
	maxAttempts int,
	retryDelay time.Duration,
) (*HotelEmailService, error) {
	if emailService == nil {
		return nil, fmt.Errorf("hotel email service requires email service")
	}
	if confRepo == nil || roomRepo == nil || userRepo == nil || organizerRepo == nil || emailLogRepo == nil {
		return nil, fmt.Errorf("hotel email service requires repositories")
	}
	if maxAttempts <= 0 {
		maxAttempts = defaultHotelEmailRetryAttempts
	}

	return &HotelEmailService{
		emailService:  emailService,
		confRepo:      confRepo,
		roomRepo:      roomRepo,
		userRepo:      userRepo,
		organizerRepo: organizerRepo,
		emailLogRepo:  emailLogRepo,
		maxAttempts:   maxAttempts,
		retryDelay:    retryDelay,
	}, nil
}

func (s *HotelEmailService) NotifyBookingCreated(ctx context.Context, booking *models.Booking) error {
	return s.notify(ctx, booking, email.TemplateTypeNewBooking, "booking_created")
}

func (s *HotelEmailService) NotifyBookingCancelled(ctx context.Context, booking *models.Booking) error {
	return s.notify(ctx, booking, email.TemplateTypeCancellation, "booking_cancelled")
}

func (s *HotelEmailService) notify(
	ctx context.Context,
	booking *models.Booking,
	templateType email.TemplateType,
	emailType string,
) error {
	if booking == nil {
		return fmt.Errorf("failed to send hotel email notification: booking is nil")
	}
	conferences, err := s.confRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: load conference list: %w", err)
	}

	var conference *models.Conference
	for i := range conferences {
		if conferences[i].ID == booking.ConferenceID {
			conference = conferences[i]
			break
		}
	}
	if conference == nil {
		return fmt.Errorf("failed to send hotel email notification: %w", ErrConferenceNotFound)
	}

	hotelRecipient, err := parseHotelRecipient(conference.HotelEmail)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: %w", err)
	}

	room, err := s.roomRepo.GetByID(ctx, booking.RoomID)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: load room: %w", err)
	}

	guest, err := s.userRepo.GetByID(ctx, booking.UserID)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: load guest: %w", err)
	}

	data, err := email.MapBookingTemplateData(email.BookingTemplateSource{
		Booking:    booking,
		Conference: conference,
		Room:       room,
		Guest:      guest,
	})
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: map email template data: %w", err)
	}

	bccRecipients, err := s.loadOrganizerBCC(ctx, conference.ID, hotelRecipient.Email)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: load organizer bcc: %w", err)
	}

	message, err := s.emailService.ComposeBookingMessageWithBCC(templateType, []email.Address{hotelRecipient}, bccRecipients, data)
	if err != nil {
		return fmt.Errorf("failed to send hotel email notification: compose message: %w", err)
	}

	var sendErr error
	for attempt := 1; attempt <= s.maxAttempts; attempt++ {
		sendErr = s.emailService.Send(ctx, message)
		if sendErr != nil {
			if logErr := s.recordEmailAttempt(ctx, booking.ID, emailType, message.Subject, hotelRecipient.Email, models.EmailLogStatusFailed, attempt, sendErr.Error()); logErr != nil {
				slog.Error("failed to record failed hotel email attempt",
					"error", logErr,
					"booking_id", booking.ID,
					"email_type", emailType,
					"attempt", attempt,
				)
			}
			if attempt < s.maxAttempts && s.retryDelay > 0 {
				select {
				case <-ctx.Done():
					return fmt.Errorf("failed to send hotel email notification: %w", ctx.Err())
				case <-time.After(s.retryDelay):
				}
			}
			continue
		}

		if logErr := s.recordEmailAttempt(ctx, booking.ID, emailType, message.Subject, hotelRecipient.Email, models.EmailLogStatusSent, attempt, ""); logErr != nil {
			slog.Error("failed to record successful hotel email attempt",
				"error", logErr,
				"booking_id", booking.ID,
				"email_type", emailType,
				"attempt", attempt,
			)
		}
		return nil
	}

	return fmt.Errorf("%w: attempts=%d last_error=%v", ErrHotelEmailDeliveryFailed, s.maxAttempts, sendErr)
}

func parseHotelRecipient(hotelEmail string) (email.Address, error) {
	recipient := email.Address{Email: strings.TrimSpace(hotelEmail), Name: "Hotel"}
	if recipient.Email == "" {
		return email.Address{}, ErrHotelEmailNotConfigured
	}
	if err := recipient.Validate(); err != nil {
		return email.Address{}, fmt.Errorf("%w: %v", ErrHotelEmailInvalid, err)
	}

	return recipient, nil
}

func (s *HotelEmailService) loadOrganizerBCC(ctx context.Context, conferenceID int64, hotelEmail string) ([]email.Address, error) {
	emails, err := s.organizerRepo.ListEmailsByConference(ctx, conferenceID)
	if err != nil {
		return nil, err
	}

	hotelEmailNormalized := strings.ToLower(strings.TrimSpace(hotelEmail))
	bcc := make([]email.Address, 0, len(emails))
	seen := map[string]struct{}{}
	for i := range emails {
		normalized := strings.ToLower(strings.TrimSpace(emails[i]))
		if normalized == "" || normalized == hotelEmailNormalized {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		addr := email.Address{Email: strings.TrimSpace(emails[i])}
		if err := addr.Validate(); err != nil {
			return nil, fmt.Errorf("invalid organizer email %q: %w", emails[i], err)
		}
		bcc = append(bcc, addr)
	}

	return bcc, nil
}

func (s *HotelEmailService) recordEmailAttempt(
	ctx context.Context,
	bookingID int64,
	emailType string,
	subject string,
	recipient string,
	status models.EmailLogStatus,
	attempt int,
	errorDetails string,
) error {
	_, err := s.emailLogRepo.Create(ctx, &models.EmailLog{
		BookingID:    bookingID,
		EmailType:    emailType,
		Recipient:    recipient,
		Subject:      subject,
		Status:       status,
		Attempt:      attempt,
		ErrorDetails: strings.TrimSpace(errorDetails),
	})
	if err != nil {
		return fmt.Errorf("create email log: %w", err)
	}

	return nil
}

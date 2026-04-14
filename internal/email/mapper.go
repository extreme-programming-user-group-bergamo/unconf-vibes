package email

import (
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/models"
)

type BookingTemplateSource struct {
	Booking    *models.Booking
	Conference *models.Conference
	Room       *models.Room
	Guest      *models.User
}

func MapBookingTemplateData(source BookingTemplateSource) (TemplateData, error) {
	if source.Booking == nil {
		return TemplateData{}, fmt.Errorf("booking is required")
	}
	if source.Conference == nil {
		return TemplateData{}, fmt.Errorf("conference is required")
	}
	if source.Room == nil {
		return TemplateData{}, fmt.Errorf("room is required")
	}
	if source.Guest == nil {
		return TemplateData{}, fmt.Errorf("guest is required")
	}

	guestName := strings.TrimSpace(source.Guest.DisplayName)
	if guestName == "" {
		guestName = strings.TrimSpace(source.Guest.GitHubID)
	}
	if guestName == "" {
		guestName = "Guest"
	}

	specialRequests := strings.TrimSpace(source.Booking.Notes)
	if specialRequests == "" {
		specialRequests = "None"
	}

	data := TemplateData{
		GuestName:       guestName,
		RoomNumber:      strings.TrimSpace(source.Room.RoomNumber),
		StartDate:       source.Conference.StartDate.Format("2006-01-02"),
		EndDate:         source.Conference.EndDate.Format("2006-01-02"),
		SpecialRequests: specialRequests,
	}
	if err := data.Validate(); err != nil {
		return TemplateData{}, fmt.Errorf("invalid mapped template data: %w", err)
	}

	return data, nil
}

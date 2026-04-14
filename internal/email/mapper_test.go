package email

import (
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapBookingTemplateData_MapsFields(t *testing.T) {
	start := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	end := start.Add(48 * time.Hour)

	data, err := MapBookingTemplateData(BookingTemplateSource{
		Booking: &models.Booking{
			Notes: "Wheelchair-accessible room",
		},
		Conference: &models.Conference{
			StartDate: start,
			EndDate:   end,
		},
		Room: &models.Room{
			RoomNumber: "307",
		},
		Guest: &models.User{
			DisplayName: "Grace Hopper",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Grace Hopper", data.GuestName)
	assert.Equal(t, "307", data.RoomNumber)
	assert.Equal(t, "2026-05-20", data.StartDate)
	assert.Equal(t, "2026-05-22", data.EndDate)
	assert.Equal(t, "Wheelchair-accessible room", data.SpecialRequests)
}

func TestMapBookingTemplateData_DefaultsSpecialRequests(t *testing.T) {
	start := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	data, err := MapBookingTemplateData(BookingTemplateSource{
		Booking: &models.Booking{},
		Conference: &models.Conference{
			StartDate: start,
			EndDate:   end,
		},
		Room: &models.Room{
			RoomNumber: "101",
		},
		Guest: &models.User{
			DisplayName: "Linus Torvalds",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "None", data.SpecialRequests)
}

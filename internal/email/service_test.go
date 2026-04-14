package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSender struct {
	message Message
}

func (s *stubSender) Send(_ context.Context, message Message) error {
	s.message = message
	return nil
}

func TestService_ComposeAndSendBookingEmail(t *testing.T) {
	renderer := NewTemplateRenderer()
	sender := &stubSender{}
	svc, err := NewService(renderer, sender, Address{Email: "from@example.com", Name: "UNCONF"})
	require.NoError(t, err)

	err = svc.SendBookingEmail(context.Background(), TemplateTypeNewBooking, []Address{{Email: "to@example.com"}}, TemplateData{
		GuestName:       "Test Guest",
		GuestEmail:      "guest@example.com",
		RoomNumber:      "101",
		StartDate:       "2026-01-01",
		EndDate:         "2026-01-02",
		SpecialRequests: "None",
	})
	require.NoError(t, err)
	assert.Equal(t, "to@example.com", sender.message.To[0].Email)
	assert.Contains(t, sender.message.Subject, "New Booking")
}

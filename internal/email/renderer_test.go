package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateRenderer_Render_AllBookingTemplates(t *testing.T) {
	renderer := NewTemplateRenderer()
	data := TemplateData{
		GuestName:       "Ada Lovelace",
		RoomNumber:      "204",
		StartDate:       "2026-05-20",
		EndDate:         "2026-05-22",
		SpecialRequests: "Late check-in",
	}

	tests := []struct {
		name        string
		template    TemplateType
		subjectHint string
		bodyHint    string
	}{
		{
			name:        "new booking",
			template:    TemplateTypeNewBooking,
			subjectHint: "New Booking",
			bodyHint:    "created",
		},
		{
			name:        "cancellation",
			template:    TemplateTypeCancellation,
			subjectHint: "Cancellation",
			bodyHint:    "cancelled",
		},
		{
			name:        "modification",
			template:    TemplateTypeModification,
			subjectHint: "Modification",
			bodyHint:    "modified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := renderer.Render(tt.template, data)
			require.NoError(t, err)
			assert.Contains(t, rendered.Subject, tt.subjectHint)
			assert.Contains(t, rendered.Body, data.GuestName)
			assert.Contains(t, rendered.Body, data.RoomNumber)
			assert.Contains(t, rendered.Body, data.StartDate)
			assert.Contains(t, rendered.Body, data.EndDate)
			assert.Contains(t, rendered.Body, data.SpecialRequests)
			assert.Contains(t, rendered.Body, tt.bodyHint)
		})
	}
}

func TestTemplateRenderer_Render_RejectsInvalidInput(t *testing.T) {
	renderer := NewTemplateRenderer()

	_, err := renderer.Render(TemplateTypeNewBooking, TemplateData{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "guest name is required")
}

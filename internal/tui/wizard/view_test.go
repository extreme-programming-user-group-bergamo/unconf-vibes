package wizard

import (
	"context"
	"regexp"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiSeq.ReplaceAllString(s, "")
}

func TestView_ConfirmStepContainsRoomSummary(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	out := stripANSI(model.View())
	assert.Contains(t, out, "Step 1/4 - Confirm Room")
	assert.Contains(t, out, "Room: 205")
	assert.Contains(t, out, "Type: double")
	assert.Contains(t, out, "$149.00/night")
	assert.Contains(t, out, "Occupants: Alex")
}

func TestView_PrivacyAndReviewText(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())
	model.step = StepPrivacy
	model.privacySetting = privacyPrivate

	out := stripANSI(model.View())
	assert.Contains(t, out, "Select privacy setting")
	assert.Contains(t, out, "Private - room occupancy is shown without your display name")

	model.step = StepReview
	model.notes = "wheelchair access"

	out = stripANSI(model.View())
	assert.Contains(t, out, "Review booking before submission")
	assert.Contains(t, out, "Privacy: "+privacyPrivate)
	assert.Contains(t, out, "Notes: wheelchair access")
}

func TestView_SuccessStepShowsConfirmationDetails(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())
	model.step = StepSuccess
	model.booking = &client.BookingResponse{ID: 77, RoomID: 12, ConferenceID: 4, Status: "requested", PrivacySetting: privacyPublic}

	out := stripANSI(model.View())
	assert.Contains(t, out, "Booking created successfully")
	assert.Contains(t, out, "Booking ID: 77")
	assert.Contains(t, out, "Room ID: 12")
	assert.Contains(t, out, "Conference ID: 4")
}

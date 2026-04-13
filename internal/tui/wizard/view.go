package wizard

import (
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
)

// View renders the wizard screen.
func (m *Model) View() string {
	var b strings.Builder

	b.WriteString(m.styles.Header.Render("UNCONF Booking Wizard"))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Conference: %s", m.roomSel.ConferenceSlug)))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render(m.step.title()))
	b.WriteString("\n\n")

	if m.submitting {
		b.WriteString(m.styles.Subtle.Render("Submitting booking..."))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"q: quit"}))
		b.WriteString("\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.styles.Error.Render("Could not create booking."))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"r: retry", "b: back", "q: quit"}))
		b.WriteString("\n")
		return b.String()
	}

	switch m.step {
	case StepConfirm:
		b.WriteString(renderConfirmStep(m.roomSel.Room))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"enter: continue", "b: exit wizard", "q: quit"}))
	case StepPrivacy:
		b.WriteString(renderPrivacyStep(m.privacySetting))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"left/right: toggle", "p/v: public/private", "enter: continue", "b: back", "q: quit"}))
	case StepNotes:
		b.WriteString(renderNotesStep(m.notes))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"type: add notes", "backspace: delete", "enter: continue", "esc/left: back", "q: quit"}))
	case StepReview:
		b.WriteString(renderReviewStep(m.roomSel.Room, m.privacySetting, m.notes))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"enter: confirm booking", "b: back", "q: quit"}))
	case StepSuccess:
		b.WriteString(renderSuccessStep(m.booking))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, []string{"enter: finish", "q: quit"}))
	}

	b.WriteString("\n")
	return b.String()
}

func renderConfirmStep(room client.RoomResponse) string {
	occupants := occupantNames(room.Occupants)
	return strings.Join([]string{
		"Confirm your room selection:",
		fmt.Sprintf("  Room: %s", room.RoomNumber),
		fmt.Sprintf("  Type: %s", room.RoomType),
		fmt.Sprintf("  Price: $%.2f/night", room.PricePerNight),
		fmt.Sprintf("  Occupants: %s", strings.Join(occupants, ", ")),
		fmt.Sprintf("  Availability: %d/%d spots", room.SpotsAvailable, room.Capacity),
	}, "\n")
}

func renderPrivacyStep(privacySetting string) string {
	publicMarker := " "
	privateMarker := " "
	if privacySetting == privacyPublic {
		publicMarker = "x"
	} else {
		privateMarker = "x"
	}

	return strings.Join([]string{
		"Select privacy setting:",
		fmt.Sprintf("  [%s] Public  - your display name appears in room occupant lists", publicMarker),
		fmt.Sprintf("  [%s] Private - room occupancy is shown without your display name", privateMarker),
	}, "\n")
}

func renderNotesStep(notes string) string {
	displayNotes := notes
	if strings.TrimSpace(displayNotes) == "" {
		displayNotes = "(none)"
	}

	return strings.Join([]string{
		"Optional hotel notes (dietary/accessibility):",
		"  Notes:",
		fmt.Sprintf("  %s", displayNotes),
	}, "\n")
}

func renderReviewStep(room client.RoomResponse, privacySetting string, notes string) string {
	displayNotes := strings.TrimSpace(notes)
	if displayNotes == "" {
		displayNotes = "(none)"
	}

	return strings.Join([]string{
		"Review booking before submission:",
		fmt.Sprintf("  Room: %s (%s)", room.RoomNumber, room.RoomType),
		fmt.Sprintf("  Price: $%.2f/night", room.PricePerNight),
		fmt.Sprintf("  Privacy: %s", privacySetting),
		fmt.Sprintf("  Notes: %s", displayNotes),
	}, "\n")
}

func renderSuccessStep(booking *client.BookingResponse) string {
	if booking == nil {
		return "Booking confirmed."
	}

	return strings.Join([]string{
		"Booking created successfully.",
		fmt.Sprintf("  Booking ID: %d", booking.ID),
		fmt.Sprintf("  Room ID: %d", booking.RoomID),
		fmt.Sprintf("  Conference ID: %d", booking.ConferenceID),
		fmt.Sprintf("  Status: %s", booking.Status),
		fmt.Sprintf("  Privacy: %s", booking.PrivacySetting),
	}, "\n")
}

func occupantNames(occupants []client.RoomOccupantResponse) []string {
	if len(occupants) == 0 {
		return []string{"none"}
	}

	names := make([]string, 0, len(occupants))
	for _, occupant := range occupants {
		name := strings.TrimSpace(occupant.DisplayName)
		if name == "" {
			name = "Private attendee"
		}
		names = append(names, name)
	}

	return names
}

package wizard

// Step identifies the active booking wizard page.
type Step int

const (
	StepConfirm Step = iota
	StepPrivacy
	StepNotes
	StepReview
	StepSuccess
)

func (s Step) title() string {
	switch s {
	case StepConfirm:
		return "Step 1/4 - Confirm Room"
	case StepPrivacy:
		return "Step 2/4 - Privacy"
	case StepNotes:
		return "Step 3/4 - Notes"
	case StepReview:
		return "Step 4/4 - Review"
	case StepSuccess:
		return "Booking Confirmed"
	default:
		return "Booking Wizard"
	}
}

package wizard

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
)

// RoomSelection carries the selected room context from room explorer into booking wizard.
type RoomSelection struct {
	ConferenceSlug string
	Room           client.RoomResponse
}

// SubmitBookingFunc submits the final booking payload.
type SubmitBookingFunc func(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error)

const (
	notesMaxRunes  = 280
	privacyPublic  = "public"
	privacyPrivate = "private"
)

// Model is the Bubble Tea model for the booking wizard.
type Model struct {
	ctx     context.Context
	styles  common.Styles
	submit  SubmitBookingFunc
	roomSel RoomSelection

	step           Step
	privacySetting string
	notes          string
	err            error
	submitting     bool
	booking        *client.BookingResponse
}

// BookingResult returns the created booking if submission completed successfully.
func (m *Model) BookingResult() (*client.BookingResponse, bool) {
	if m.step != StepSuccess || m.booking == nil {
		return nil, false
	}

	return m.booking, true
}

// NewModel creates a booking wizard model for one room selection.
func NewModel(ctx context.Context, roomSel RoomSelection, submit SubmitBookingFunc, styles common.Styles) *Model {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Model{
		ctx:            ctx,
		styles:         styles,
		submit:         submit,
		roomSel:        roomSel,
		step:           StepConfirm,
		privacySetting: privacyPublic,
		notes:          "",
	}
}

// Init satisfies Bubble Tea Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

type bookingSubmittedMsg struct {
	booking *client.BookingResponse
	err     error
}

func (m *Model) submitBookingCmd() tea.Cmd {
	return func() tea.Msg {
		if m.submit == nil {
			return bookingSubmittedMsg{err: fmt.Errorf("booking submitter is not configured")}
		}

		input := client.CreateBookingRequest{
			RoomID:         m.roomSel.Room.ID,
			ConferenceID:   m.roomSel.Room.ConferenceID,
			PrivacySetting: m.privacySetting,
			Notes:          strings.TrimSpace(m.notes),
		}

		booking, err := m.submit(m.ctx, input)
		if err != nil {
			return bookingSubmittedMsg{err: fmt.Errorf("submit booking: %w", err)}
		}

		return bookingSubmittedMsg{booking: booking}
	}
}

// Update processes keyboard input and submit completion events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bookingSubmittedMsg:
		m.submitting = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.err = nil
		m.booking = msg.booking
		m.step = StepSuccess
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC || isRuneKey(msg, 'q') {
		return m, tea.Quit
	}

	if m.submitting {
		return m, nil
	}

	if m.step == StepSuccess {
		if msg.Type == tea.KeyEnter {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.err != nil {
		switch {
		case isRuneKey(msg, 'r'):
			m.err = nil
			m.submitting = true
			return m, m.submitBookingCmd()
		case isRuneKey(msg, 'b'):
			m.err = nil
			m.step = StepNotes
			return m, nil
		}

		return m, nil
	}

	switch m.step {
	case StepConfirm:
		switch {
		case msg.Type == tea.KeyEnter, msg.Type == tea.KeyRight, isRuneKey(msg, 'l', 'n'):
			m.step = StepPrivacy
		case msg.Type == tea.KeyEsc, msg.Type == tea.KeyLeft, isRuneKey(msg, 'h', 'b'):
			return m, tea.Quit
		}
	case StepPrivacy:
		switch {
		case msg.Type == tea.KeyLeft, msg.Type == tea.KeyRight, msg.Type == tea.KeyUp, msg.Type == tea.KeyDown, msg.Type == tea.KeyTab:
			if m.privacySetting == privacyPublic {
				m.privacySetting = privacyPrivate
			} else {
				m.privacySetting = privacyPublic
			}
		case isRuneKey(msg, 'p'):
			m.privacySetting = privacyPublic
		case isRuneKey(msg, 'v'):
			m.privacySetting = privacyPrivate
		case msg.Type == tea.KeyEnter, isRuneKey(msg, 'n'):
			m.step = StepNotes
		case isRuneKey(msg, 'b', 'h'), msg.Type == tea.KeyEsc:
			m.step = StepConfirm
		}
	case StepNotes:
		switch {
		case msg.Type == tea.KeyEnter:
			m.step = StepReview
		case msg.Type == tea.KeyLeft, msg.Type == tea.KeyEsc:
			m.step = StepPrivacy
		case msg.Type == tea.KeyBackspace:
			m.notes = trimLastRune(m.notes)
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0:
			currentRunes := utf8.RuneCountInString(m.notes)
			for _, r := range msg.Runes {
				if !unicode.IsPrint(r) {
					continue
				}
				if currentRunes >= notesMaxRunes {
					break
				}
				m.notes += string(r)
				currentRunes++
			}
		}
	case StepReview:
		switch {
		case msg.Type == tea.KeyEnter, isRuneKey(msg, 'c'):
			m.submitting = true
			m.err = nil
			return m, m.submitBookingCmd()
		case isRuneKey(msg, 'b'), msg.Type == tea.KeyEsc:
			m.step = StepNotes
		}
	}

	return m, nil
}

func trimLastRune(s string) string {
	if s == "" {
		return s
	}

	_, size := utf8.DecodeLastRuneInString(s)
	if size <= 0 || size > len(s) {
		return ""
	}

	return s[:len(s)-size]
}

func isRuneKey(msg tea.KeyMsg, runes ...rune) bool {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return false
	}

	r := msg.Runes[0]
	for _, candidate := range runes {
		if r == candidate {
			return true
		}
	}

	return false
}

package wizard

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRoomSelection() RoomSelection {
	return RoomSelection{
		ConferenceSlug: "socrates-26",
		Room: client.RoomResponse{
			ID:             12,
			ConferenceID:   4,
			RoomNumber:     "205",
			RoomType:       "double",
			PricePerNight:  149,
			SpotsAvailable: 1,
			Capacity:       2,
			Occupants: []client.RoomOccupantResponse{
				{DisplayName: "Alex"},
			},
		},
	}
}

func TestModel_StepProgressionAndBackNavigation(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepPrivacy, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepReview, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, StepNotes, next.(*Model).step)
}

func TestModel_PrivacyAndNotesStatePersistedIntoPayload(t *testing.T) {
	var captured client.CreateBookingRequest
	model := NewModel(context.Background(), testRoomSelection(), func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		captured = input
		return &client.BookingResponse{ID: 99, RoomID: input.RoomID, ConferenceID: input.ConferenceID, Status: "requested", PrivacySetting: input.PrivacySetting}, nil
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, privacyPrivate, next.(*Model).privacySetting)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dietary")})
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)

	review := next.(*Model)
	require.Equal(t, StepReview, review.step)

	next, cmd = review.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	next, _ = next.(*Model).Update(msg)

	final := next.(*Model)
	assert.Equal(t, StepSuccess, final.step)
	assert.Equal(t, int64(12), captured.RoomID)
	assert.Equal(t, int64(4), captured.ConferenceID)
	assert.Equal(t, privacyPrivate, captured.PrivacySetting)
	assert.Equal(t, "dietary", captured.Notes)
}

func TestModel_SubmitErrorAllowsRetryAndBack(t *testing.T) {
	attempts := 0
	model := NewModel(context.Background(), testRoomSelection(), func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("room_full")
		}
		return &client.BookingResponse{ID: 1, Status: "requested"}, nil
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	next, _ = next.(*Model).Update(cmd())

	failed := next.(*Model)
	require.Error(t, failed.err)
	assert.Equal(t, StepReview, failed.step)

	next, _ = failed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, StepNotes, next.(*Model).step)
}

func TestModel_NotesStepAllowsTypingNavigationLetters(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bnb")})
	notes := next.(*Model)
	assert.Equal(t, StepNotes, notes.step)
	assert.Equal(t, "bnb", notes.notes)
}

func TestModel_NotesStepSingleLetterInputDoesNotNavigate(t *testing.T) {
	t.Run("n stays in notes step", func(t *testing.T) {
		model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.Equal(t, StepNotes, next.(*Model).step)

		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		notes := next.(*Model)

		assert.Equal(t, StepNotes, notes.step)
		assert.Equal(t, "n", notes.notes)
	})

	t.Run("b stays in notes step", func(t *testing.T) {
		model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.Equal(t, StepNotes, next.(*Model).step)

		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		notes := next.(*Model)

		assert.Equal(t, StepNotes, notes.step)
		assert.Equal(t, "b", notes.notes)
	})
}

func TestModel_NotesBackspaceRemovesLastRune(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("naive😄")})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyBackspace})
	notes := next.(*Model)

	assert.Equal(t, "naive", notes.notes)
	assert.Equal(t, StepNotes, notes.step)
}

func TestModel_NotesRuneLimit_AllowsMultibyteContentUnderRuneLimit(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	multibyte := strings.Repeat("😄", 200)
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(multibyte)})
	notes := next.(*Model)

	assert.Equal(t, 200, utf8.RuneCountInString(notes.notes))
	assert.Equal(t, multibyte, notes.notes)
}

func TestModel_NotesRuneLimit_CapsInputAtMaxRunes(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	overLimit := strings.Repeat("😄", notesMaxRunes+20)
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(overLimit)})
	notes := next.(*Model)

	assert.Equal(t, notesMaxRunes, utf8.RuneCountInString(notes.notes))
	assert.Equal(t, strings.Repeat("😄", notesMaxRunes), notes.notes)
}

func TestModel_SubmitCanceledContextShowsErrorState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	model := NewModel(ctx, testRoomSelection(), func(ctx context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, ctx.Err()
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})

	cancel()
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	next, _ = next.(*Model).Update(cmd())

	failed := next.(*Model)
	require.Error(t, failed.err)
	assert.ErrorIs(t, failed.err, context.Canceled)
	assert.Contains(t, failed.err.Error(), "submit booking")
	assert.Equal(t, StepReview, failed.step)
	assert.False(t, failed.submitting)
}

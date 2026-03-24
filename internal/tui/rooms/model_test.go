package rooms

import (
	"context"
	"errors"
	"regexp"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiSeq.ReplaceAllString(s, "")
}

func TestModel_InitStartsLoadingAndTransitionsToLoaded(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, slug string) ([]client.RoomResponse, error) {
		assert.Equal(t, "socrates-26", slug)
		return []client.RoomResponse{{
			ID:             11,
			RoomNumber:     "101",
			RoomType:       "double",
			SpotsAvailable: 1,
			Capacity:       2,
			PricePerNight:  120,
		}}, nil
	}, common.NewStyles())

	assert.True(t, model.loading)

	cmd := model.Init()
	require.NotNil(t, cmd)

	msg := cmd()
	updated, followUp := model.Update(msg)
	require.Nil(t, followUp)

	updatedModel := updated.(*Model)
	assert.False(t, updatedModel.loading)
	assert.NoError(t, updatedModel.err)
	assert.Len(t, updatedModel.rooms, 1)
	assert.Len(t, updatedModel.visible, 1)
	assert.Contains(t, updatedModel.View(), "101")
	assert.Contains(t, updatedModel.View(), "double")
	assert.Contains(t, updatedModel.View(), "$120.00")
	assert.Contains(t, updatedModel.View(), "1/2 spots")
}

func TestModel_UpdateTransitionsToErrorState(t *testing.T) {
	expectedErr := errors.New("boom")
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return nil, expectedErr
	}, common.NewStyles())

	msg := model.Init()()
	updated, followUp := model.Update(msg)
	require.Nil(t, followUp)

	updatedModel := updated.(*Model)
	assert.False(t, updatedModel.loading)
	assert.Error(t, updatedModel.err)
	assert.ErrorIs(t, updatedModel.err, expectedErr)
	assert.Contains(t, updatedModel.View(), "Failed to load rooms")
}

func TestModel_UpdateNavigation(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{
			{ID: 1, RoomNumber: "101", RoomType: "double", SpotsAvailable: 1, Capacity: 2, PricePerNight: 120},
			{ID: 2, RoomNumber: "102", RoomType: "single", SpotsAvailable: 0, Capacity: 1, PricePerNight: 90},
		}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)

	next, _ := state.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, next.(*Model).selected)

	prev, _ := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, prev.(*Model).selected)
}

func TestModel_UpdateFilteringByRoomType(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{
			{ID: 1, RoomNumber: "101", RoomType: "single", SpotsAvailable: 1, Capacity: 1, PricePerNight: 100},
			{ID: 2, RoomNumber: "201", RoomType: "double", SpotsAvailable: 2, Capacity: 2, PricePerNight: 130},
			{ID: 3, RoomNumber: "301", RoomType: "triple", SpotsAvailable: 3, Capacity: 3, PricePerNight: 160},
		}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)

	filteredSingle, _ := state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, RoomTypeFilterSingle, filteredSingle.(*Model).filter)
	assert.Len(t, filteredSingle.(*Model).visible, 1)
	assert.Equal(t, "101", filteredSingle.(*Model).visible[0].RoomNumber)

	filteredDouble, _ := filteredSingle.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.Equal(t, RoomTypeFilterDouble, filteredDouble.(*Model).filter)
	assert.Len(t, filteredDouble.(*Model).visible, 1)
	assert.Equal(t, "201", filteredDouble.(*Model).visible[0].RoomNumber)

	allRooms, _ := filteredDouble.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.Equal(t, RoomTypeFilterAll, allRooms.(*Model).filter)
	assert.Len(t, allRooms.(*Model).visible, 3)
}

func TestModel_UpdateFilteringClampsSelectedIndexWhenVisibleShrinks(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{
			{ID: 1, RoomNumber: "101", RoomType: "single", SpotsAvailable: 1, Capacity: 1, PricePerNight: 100},
			{ID: 2, RoomNumber: "201", RoomType: "double", SpotsAvailable: 2, Capacity: 2, PricePerNight: 130},
			{ID: 3, RoomNumber: "301", RoomType: "triple", SpotsAvailable: 3, Capacity: 3, PricePerNight: 160},
		}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)

	navigated, _ := state.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, _ = navigated.(*Model).Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, navigated.(*Model).selected)

	filteredDouble, _ := navigated.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.Len(t, filteredDouble.(*Model).visible, 1)
	assert.Equal(t, 0, filteredDouble.(*Model).selected)
	assert.Equal(t, "201", filteredDouble.(*Model).visible[filteredDouble.(*Model).selected].RoomNumber)
}

func TestModel_UpdateSorting(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{
			{ID: 2, RoomNumber: "202", RoomType: "double", SpotsAvailable: 1, Capacity: 2, PricePerNight: 150},
			{ID: 1, RoomNumber: "101", RoomType: "single", SpotsAvailable: 0, Capacity: 1, PricePerNight: 100},
			{ID: 3, RoomNumber: "303", RoomType: "triple", SpotsAvailable: 2, Capacity: 3, PricePerNight: 200},
		}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)
	assert.Equal(t, "303", state.visible[0].RoomNumber)

	byPrice, _ := state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Equal(t, SortModePrice, byPrice.(*Model).sortMode)
	assert.Equal(t, "101", byPrice.(*Model).visible[0].RoomNumber)

	byAvailability, _ := byPrice.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	assert.Equal(t, SortModeAvailability, byAvailability.(*Model).sortMode)
	assert.Equal(t, "303", byAvailability.(*Model).visible[0].RoomNumber)
}

func TestModel_ViewRendersOccupantsAndPrivacyMasking(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{
			ID:             9,
			RoomNumber:     "401",
			RoomType:       "double",
			SpotsAvailable: 1,
			Capacity:       2,
			PricePerNight:  180,
			Occupants: []client.RoomOccupantResponse{
				{DisplayName: "Alice"},
				{DisplayName: "Private attendee"},
				{},
			},
		}}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	out := stripANSI(loaded.(*Model).View())

	assert.Contains(t, out, "  Occupants:")
	assert.Contains(t, out, "Alice")
	assert.Contains(t, out, "Private attendee")
}

func TestModel_ViewUsesReadableFilterAndSortLabels(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 1, RoomNumber: "101", RoomType: "single", SpotsAvailable: 1, Capacity: 1, PricePerNight: 99}}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)

	assert.Contains(t, stripANSI(state.View()), "Filter: All rooms | Sort: Availability")

	sorted, _ := state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.Contains(t, stripANSI(sorted.(*Model).View()), "Filter: All rooms | Sort: Price")
}

func TestModel_EnterInitiatesBookingSelection(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 7, RoomNumber: "205", RoomType: "double", SpotsAvailable: 1, Capacity: 2, PricePerNight: 140}}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	next, cmd := loaded.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	selection, ok := next.(*Model).BookingSelection()
	require.True(t, ok)
	assert.Equal(t, "socrates-26", selection.ConferenceSlug)
	assert.Equal(t, "205", selection.Room.RoomNumber)
}

func TestModel_ViewShowsExtendedFooterHelp(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 1, RoomNumber: "101", RoomType: "single", SpotsAvailable: 1, Capacity: 1, PricePerNight: 99}}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	out := loaded.(*Model).View()

	assert.Contains(t, out, "s/d/t/f")
	assert.Contains(t, out, "p/v")
	assert.Contains(t, out, "enter")
	assert.Contains(t, out, "q")
}

func TestModel_UpdateQuitKeys(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", nil, common.NewStyles())

	_, quitCmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, quitCmd)

	_, quitCmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, quitCmd)
}

func TestModel_InitPropagatesProvidedContext(t *testing.T) {
	type ctxKey string
	const requestIDKey ctxKey = "request-id"

	ctx := context.WithValue(context.Background(), requestIDKey, "req-123")
	model := NewModel(ctx, "socrates-26", func(fetchCtx context.Context, _ string) ([]client.RoomResponse, error) {
		assert.Equal(t, "req-123", fetchCtx.Value(requestIDKey))
		return []client.RoomResponse{}, nil
	}, common.NewStyles())

	_ = model.Init()()
}

package rooms

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModel_InitStartsLoadingAndTransitionsToLoaded(t *testing.T) {
	model := NewModel("socrates-26", func(_ context.Context, slug string) ([]client.RoomResponse, error) {
		assert.Equal(t, "socrates-26", slug)
		return []client.RoomResponse{{RoomNumber: "101", RoomType: "double", SpotsAvailable: 1, Capacity: 2}}, nil
	}, common.NewStyles())

	assert.True(t, model.loading)

	cmd := model.Init()
	require.NotNil(t, cmd)

	msg := cmd()
	updated, followUp := model.Update(msg)
	require.Nil(t, followUp)

	updatedModel := updated.(Model)
	assert.False(t, updatedModel.loading)
	assert.NoError(t, updatedModel.err)
	assert.Len(t, updatedModel.rooms, 1)
	assert.Contains(t, updatedModel.View(), "101")
}

func TestModel_UpdateTransitionsToErrorState(t *testing.T) {
	expectedErr := errors.New("boom")
	model := NewModel("socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return nil, expectedErr
	}, common.NewStyles())

	msg := model.Init()()
	updated, followUp := model.Update(msg)
	require.Nil(t, followUp)

	updatedModel := updated.(Model)
	assert.False(t, updatedModel.loading)
	assert.Error(t, updatedModel.err)
	assert.ErrorIs(t, updatedModel.err, expectedErr)
	assert.Contains(t, updatedModel.View(), "Failed to load rooms")
}

func TestModel_UpdateNavigation(t *testing.T) {
	model := NewModel("socrates-26", func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{
			{RoomNumber: "101", RoomType: "double", SpotsAvailable: 1, Capacity: 2},
			{RoomNumber: "102", RoomType: "single", SpotsAvailable: 0, Capacity: 1},
		}, nil
	}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(Model)

	next, _ := state.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, next.(Model).selected)

	prev, _ := next.(Model).Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, prev.(Model).selected)
}

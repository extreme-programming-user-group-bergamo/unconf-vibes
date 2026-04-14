package dashboard

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModel_InitLoadAndRender(t *testing.T) {
	model := NewModel(context.Background(), "socrates-26", func(_ context.Context, slug string, _ client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
		assert.Equal(t, "socrates-26", slug)
		return &client.OrganizerDashboardResponse{
			TotalRegistrations: 3,
			Capacity:           10,
			CapacityUsagePct:   30,
			RoomFillRates: []client.OrganizerDashboardRoomFillRateResponse{
				{RoomType: "double", Registrations: 2, Capacity: 4, FillRatePct: 50},
			},
			Attendees: []client.OrganizerDashboardAttendeeResponse{
				{Name: "Alice", Email: "alice@test.dev", RoomNumber: "101", RoomType: "double", BookingStatus: "confirmed", DietaryAccessibilityNote: "Vegan", PrivacySetting: "private"},
			},
		}, nil
	}, client.DashboardQuery{}, common.NewStyles())

	msg := model.Init()()
	next, _ := model.Update(msg)
	out := next.(*Model).View()
	assert.Contains(t, out, "Organizer dashboard")
	assert.Contains(t, out, "Registrations: 3")
	assert.Contains(t, out, "Alice")
	assert.Contains(t, out, "Vegan")
}

func TestModel_LoadError(t *testing.T) {
	model := NewModel(context.Background(), "slug", func(_ context.Context, _ string, _ client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
		return nil, errors.New("boom")
	}, client.DashboardQuery{}, common.NewStyles())

	next, _ := model.Update(model.Init()())
	assert.Contains(t, next.(*Model).View(), "Failed to load organizer dashboard")
}

func TestModel_FilterAndSearchInteractionsTriggerReload(t *testing.T) {
	var received []client.DashboardQuery
	model := NewModel(context.Background(), "slug", func(_ context.Context, _ string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
		received = append(received, query)
		return &client.OrganizerDashboardResponse{}, nil
	}, client.DashboardQuery{}, common.NewStyles())

	loaded, _ := model.Update(model.Init()())
	state := loaded.(*Model)

	next, cmd := state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	require.NotNil(t, cmd)
	_, _ = next.(*Model).Update(cmd())

	next, cmd = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	require.NotNil(t, cmd)
	_, _ = next.(*Model).Update(cmd())

	next, cmd = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	require.NotNil(t, cmd)
	_, _ = next.(*Model).Update(cmd())

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	editing := next.(*Model)
	require.True(t, editing.editingSearch)

	typed, _ := editing.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ali")})
	applied, cmd := typed.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	_, _ = applied.(*Model).Update(cmd())

	require.GreaterOrEqual(t, len(received), 4)
	last := received[len(received)-1]
	assert.Equal(t, "single", strings.ToLower(last.RoomType))
	assert.Equal(t, "confirmed", strings.ToLower(last.BookingStatus))
	assert.True(t, last.HasSpecialRequests)
	assert.Equal(t, "ali", last.Search)
}

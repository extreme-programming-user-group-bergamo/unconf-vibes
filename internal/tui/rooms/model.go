package rooms

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
)

// FetchRoomsFunc loads rooms for a conference.
type FetchRoomsFunc func(ctx context.Context, conferenceSlug string) ([]client.RoomResponse, error)

// Model is the Bubble Tea model for room exploration.
type Model struct {
	ctx            context.Context
	conferenceSlug string
	fetchRooms     FetchRoomsFunc
	styles         common.Styles

	loading  bool
	err      error
	rooms    []client.RoomResponse
	selected int
}

// NewModel creates a room explorer model with loading state enabled.
func NewModel(ctx context.Context, conferenceSlug string, fetchRooms FetchRoomsFunc, styles common.Styles) *Model {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Model{
		ctx:            ctx,
		conferenceSlug: conferenceSlug,
		fetchRooms:     fetchRooms,
		styles:         styles,
		loading:        true,
		rooms:          []client.RoomResponse{},
	}
}

// Init starts asynchronous room loading.
func (m *Model) Init() tea.Cmd {
	return m.loadRoomsCmd()
}

func (m *Model) loadRoomsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.fetchRooms == nil {
			return RoomsLoadErrMsg{Err: fmt.Errorf("rooms loader is not configured")}
		}

		rooms, err := m.fetchRooms(m.ctx, m.conferenceSlug)
		if err != nil {
			return RoomsLoadErrMsg{Err: err}
		}

		return RoomsLoadedMsg{Rooms: rooms}
	}
}

// Update handles async load events and keyboard navigation.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RoomsLoadedMsg:
		m.loading = false
		m.err = nil
		m.rooms = msg.Rooms
		if m.selected >= len(m.rooms) {
			m.selected = max(0, len(m.rooms)-1)
		}
		return m, nil
	case RoomsLoadErrMsg:
		m.loading = false
		m.err = msg.Err
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.rooms)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

// View renders the room explorer screen.
func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(common.RenderHeader(m.styles, m.conferenceSlug))
	b.WriteString("\n\n")

	switch {
	case m.loading:
		b.WriteString(m.styles.Subtle.Render("Loading rooms..."))
	case m.err != nil:
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Failed to load rooms: %v", m.err)))
	case len(m.rooms) == 0:
		b.WriteString(m.styles.Subtle.Render("No rooms found for this conference."))
	default:
		for i, room := range m.rooms {
			line := fmt.Sprintf("%s  %s  [%d/%d available]", room.RoomNumber, room.RoomType, room.SpotsAvailable, room.Capacity)
			if i == m.selected {
				b.WriteString(m.styles.Selected.Render(line))
			} else {
				b.WriteString(m.styles.Normal.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(common.RenderFooter(m.styles, []string{"up/down: navigate", "q: quit"}))
	b.WriteString("\n")

	return b.String()
}

package rooms

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
)

const privateAttendeeDisplayName = "Private attendee"

// RoomTypeFilter defines the active room type filter.
type RoomTypeFilter string

const (
	RoomTypeFilterAll    RoomTypeFilter = "all"
	RoomTypeFilterSingle RoomTypeFilter = "single"
	RoomTypeFilterDouble RoomTypeFilter = "double"
	RoomTypeFilterTriple RoomTypeFilter = "triple"
)

// SortMode defines the active room sort mode.
type SortMode string

const (
	SortModeAvailability SortMode = "availability"
	SortModePrice        SortMode = "price"
)

// BookingSelection is returned to the CLI as the handoff contract for booking flow.
type BookingSelection struct {
	ConferenceSlug string
	Room           client.RoomResponse
}

// InviteSelection is returned to the CLI as the handoff contract for invite flow.
type InviteSelection struct {
	ConferenceSlug string
	Room           client.RoomResponse
}

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
	visible  []client.RoomResponse
	selected int

	filter             RoomTypeFilter
	sortMode           SortMode
	bookingInitiated   bool
	selectedForBooking client.RoomResponse
	inviteInitiated    bool
	selectedForInvite  client.RoomResponse
	ownRoomID          int64
}

// NewModel creates a room explorer model with loading state enabled.
func NewModel(ctx context.Context, conferenceSlug string, fetchRooms FetchRoomsFunc, ownRoomID int64, styles common.Styles) *Model {
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
		visible:        []client.RoomResponse{},
		filter:         RoomTypeFilterAll,
		sortMode:       SortModeAvailability,
		ownRoomID:      ownRoomID,
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
		m.rebuildVisibleRooms()
		return m, nil
	case RoomsLoadErrMsg:
		m.loading = false
		m.err = msg.Err
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			m.filter = RoomTypeFilterSingle
			m.rebuildVisibleRooms()
		case "d":
			m.filter = RoomTypeFilterDouble
			m.rebuildVisibleRooms()
		case "t":
			m.filter = RoomTypeFilterTriple
			m.rebuildVisibleRooms()
		case "f":
			m.filter = RoomTypeFilterAll
			m.rebuildVisibleRooms()
		case "p":
			m.sortMode = SortModePrice
			m.rebuildVisibleRooms()
		case "v":
			m.sortMode = SortModeAvailability
			m.rebuildVisibleRooms()
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.visible)-1 {
				m.selected++
			}
		case "enter":
			if len(m.visible) > 0 {
				m.bookingInitiated = true
				m.selectedForBooking = m.visible[m.selected]
				return m, tea.Quit
			}
		case "i":
			if len(m.visible) > 0 {
				selected := m.visible[m.selected]
				if m.ownRoomID > 0 && selected.ID == m.ownRoomID && selected.SpotsAvailable > 0 {
					m.inviteInitiated = true
					m.selectedForInvite = selected
					return m, tea.Quit
				}
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
	case len(m.visible) == 0:
		b.WriteString(m.styles.Subtle.Render("No rooms found for this conference."))
	default:
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Filter: %s | Sort: %s", filterLabel(m.filter), sortLabel(m.sortMode))))
		b.WriteString("\n\n")

		for i, room := range m.visible {
			line := fmt.Sprintf("%s  %s  $%.2f  %s", room.RoomNumber, room.RoomType, room.PricePerNight, availabilityText(room))
			if i == m.selected {
				b.WriteString(m.styles.Selected.Render(line))
			} else {
				b.WriteString(m.styles.Normal.Render(line))
			}
			b.WriteString("\n")

			occupantsLine := "  Occupants: " + strings.Join(occupantNames(room.Occupants), ", ")
			if i == m.selected {
				b.WriteString(m.styles.Selected.Render(occupantsLine))
			} else {
				b.WriteString(m.styles.Subtle.Render(occupantsLine))
			}
			b.WriteString("\n")

			if room.ID == m.ownRoomID {
				ownRoomLine := "  Your room"
				if room.SpotsAvailable > 0 {
					ownRoomLine += " (press i to invite roommate)"
				}
				if i == m.selected {
					b.WriteString(m.styles.Selected.Render(ownRoomLine))
				} else {
					b.WriteString(m.styles.Subtle.Render(ownRoomLine))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n\n")
	b.WriteString(common.RenderFooter(m.styles, []string{"up/down: navigate", "s/d/t/f: filter", "p/v: sort", "enter: select", "i: invite from own room", "q: quit"}))
	b.WriteString("\n")

	return b.String()
}

// BookingSelection returns the selected room when Enter was used to initiate booking.
func (m *Model) BookingSelection() (BookingSelection, bool) {
	if !m.bookingInitiated {
		return BookingSelection{}, false
	}

	return BookingSelection{
		ConferenceSlug: m.conferenceSlug,
		Room:           m.selectedForBooking,
	}, true
}

// InviteSelection returns the selected room when i was used to initiate roommate invite.
func (m *Model) InviteSelection() (InviteSelection, bool) {
	if !m.inviteInitiated {
		return InviteSelection{}, false
	}

	return InviteSelection{
		ConferenceSlug: m.conferenceSlug,
		Room:           m.selectedForInvite,
	}, true
}

func (m *Model) rebuildVisibleRooms() {
	visible := make([]client.RoomResponse, 0, len(m.rooms))
	for _, room := range m.rooms {
		if m.filter != RoomTypeFilterAll && strings.ToLower(room.RoomType) != string(m.filter) {
			continue
		}

		visible = append(visible, room)
	}

	sortRooms(visible, m.sortMode)
	m.visible = visible

	if len(m.visible) == 0 {
		m.selected = 0
		return
	}

	if m.selected < 0 {
		m.selected = 0
	}

	if m.selected >= len(m.visible) {
		m.selected = len(m.visible) - 1
	}
}

func sortRooms(rooms []client.RoomResponse, mode SortMode) {
	if len(rooms) < 2 {
		return
	}

	sort.SliceStable(rooms, func(i, j int) bool {
		left := rooms[i]
		right := rooms[j]

		switch mode {
		case SortModePrice:
			if left.PricePerNight != right.PricePerNight {
				return left.PricePerNight < right.PricePerNight
			}
		case SortModeAvailability:
			if left.SpotsAvailable != right.SpotsAvailable {
				return left.SpotsAvailable > right.SpotsAvailable
			}
		}

		if left.RoomNumber != right.RoomNumber {
			return left.RoomNumber < right.RoomNumber
		}

		return left.ID < right.ID
	})
}

func availabilityText(room client.RoomResponse) string {
	return fmt.Sprintf("%d/%d spots", room.SpotsAvailable, room.Capacity)
}

func filterLabel(filter RoomTypeFilter) string {
	switch filter {
	case RoomTypeFilterSingle:
		return "Single"
	case RoomTypeFilterDouble:
		return "Double"
	case RoomTypeFilterTriple:
		return "Triple"
	default:
		return "All rooms"
	}
}

func sortLabel(mode SortMode) string {
	switch mode {
	case SortModePrice:
		return "Price"
	default:
		return "Availability"
	}
}

func occupantNames(occupants []client.RoomOccupantResponse) []string {
	if len(occupants) == 0 {
		return []string{"none"}
	}

	names := make([]string, 0, len(occupants))
	for _, occupant := range occupants {
		name := strings.TrimSpace(occupant.DisplayName)
		if name == "" {
			name = privateAttendeeDisplayName
		}
		names = append(names, name)
	}

	return names
}

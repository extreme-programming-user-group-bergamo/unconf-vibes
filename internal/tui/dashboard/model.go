package dashboard

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
)

// FetchDashboardFunc loads organizer dashboard data for a conference and query.
type FetchDashboardFunc func(ctx context.Context, conferenceSlug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error)

var roomTypeCycle = []string{"", "single", "double", "triple"}
var bookingStatusCycle = []string{"", "confirmed", "requested", "cancelled"}

type dashboardLoadedMsg struct {
	result *client.OrganizerDashboardResponse
}

type dashboardLoadErrMsg struct {
	err error
}

// Model is the Bubble Tea model for organizer dashboard exploration.
type Model struct {
	ctx            context.Context
	conferenceSlug string
	fetchDashboard FetchDashboardFunc
	styles         common.Styles

	loading bool
	err     error
	result  *client.OrganizerDashboardResponse

	query         client.DashboardQuery
	editingSearch bool
	searchInput   string
}

// NewModel creates a dashboard model with loading state enabled.
func NewModel(
	ctx context.Context,
	conferenceSlug string,
	fetchDashboard FetchDashboardFunc,
	initialQuery client.DashboardQuery,
	styles common.Styles,
) *Model {
	if ctx == nil {
		ctx = context.Background()
	}

	initialQuery.RoomType = normalizeQueryOption(initialQuery.RoomType)
	initialQuery.BookingStatus = normalizeQueryOption(initialQuery.BookingStatus)
	initialQuery.Search = strings.TrimSpace(initialQuery.Search)

	return &Model{
		ctx:            ctx,
		conferenceSlug: conferenceSlug,
		fetchDashboard: fetchDashboard,
		styles:         styles,
		loading:        true,
		query:          initialQuery,
		searchInput:    initialQuery.Search,
	}
}

// Init starts asynchronous dashboard loading.
func (m *Model) Init() tea.Cmd {
	return m.loadDashboardCmd()
}

func (m *Model) loadDashboardCmd() tea.Cmd {
	return func() tea.Msg {
		if m.fetchDashboard == nil {
			return dashboardLoadErrMsg{err: fmt.Errorf("dashboard loader is not configured")}
		}

		result, err := m.fetchDashboard(m.ctx, m.conferenceSlug, m.query)
		if err != nil {
			return dashboardLoadErrMsg{err: err}
		}

		return dashboardLoadedMsg{result: result}
	}
}

// Update handles async load events and keyboard interactions.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardLoadedMsg:
		m.loading = false
		m.err = nil
		m.result = msg.result
		return m, nil
	case dashboardLoadErrMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	case tea.KeyMsg:
		if m.editingSearch {
			return m.updateSearchInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.query.RoomType = nextCycledValue(m.query.RoomType, roomTypeCycle)
			m.loading = true
			return m, m.loadDashboardCmd()
		case "b":
			m.query.BookingStatus = nextCycledValue(m.query.BookingStatus, bookingStatusCycle)
			m.loading = true
			return m, m.loadDashboardCmd()
		case "s":
			m.query.HasSpecialRequests = !m.query.HasSpecialRequests
			m.loading = true
			return m, m.loadDashboardCmd()
		case "/":
			m.editingSearch = true
			m.searchInput = m.query.Search
		}
	}

	return m, nil
}

func (m *Model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.editingSearch = false
		m.searchInput = m.query.Search
		return m, nil
	case "enter":
		m.query.Search = strings.TrimSpace(m.searchInput)
		m.editingSearch = false
		m.loading = true
		return m, m.loadDashboardCmd()
	case "backspace":
		if len(m.searchInput) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.searchInput)
			m.searchInput = m.searchInput[:len(m.searchInput)-size]
		}
	default:
		if len(msg.Runes) > 0 {
			m.searchInput += string(msg.Runes)
		}
	}

	return m, nil
}

// View renders the organizer dashboard.
func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(common.RenderHeader(m.styles, m.conferenceSlug))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.styles.Subtle.Render("Loading organizer dashboard..."))
		b.WriteString("\n")
		b.WriteString(common.RenderFooter(m.styles, m.footer()))
		b.WriteString("\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Failed to load organizer dashboard: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, m.footer()))
		b.WriteString("\n")
		return b.String()
	}

	result := m.result
	if result == nil {
		b.WriteString(m.styles.Subtle.Render("No dashboard data available."))
		b.WriteString("\n\n")
		b.WriteString(common.RenderFooter(m.styles, m.footer()))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(m.styles.Header.Render("Organizer dashboard"))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render(
		fmt.Sprintf("Registrations: %d | Capacity usage: %.1f%% (%d capacity)",
			result.TotalRegistrations, result.CapacityUsagePct, result.Capacity),
	))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtle.Render(fmt.Sprintf(
		"Filters -> room:%s status:%s special:%t search:%q",
		emptyAsAll(m.query.RoomType),
		emptyAsAll(m.query.BookingStatus),
		m.query.HasSpecialRequests,
		m.query.Search,
	)))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Normal.Render("Room fill rates:"))
	b.WriteString("\n")
	for i := range result.RoomFillRates {
		rate := result.RoomFillRates[i]
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf(
			"- %s: %d/%d (%.1f%%)",
			emptyAsUnknown(rate.RoomType), rate.Registrations, rate.Capacity, rate.FillRatePct,
		)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Normal.Render("Attendees:"))
	b.WriteString("\n")
	if len(result.Attendees) == 0 {
		b.WriteString(m.styles.Subtle.Render("No attendees match current filters/search."))
		b.WriteString("\n")
	} else {
		for i := range result.Attendees {
			row := result.Attendees[i]
			roomLabel := strings.TrimSpace(row.RoomNumber)
			if roomLabel == "" {
				roomLabel = "not booked yet"
			}
			if strings.TrimSpace(row.RoomType) != "" {
				roomLabel = fmt.Sprintf("%s (%s)", roomLabel, row.RoomType)
			}
			notes := strings.TrimSpace(row.DietaryAccessibilityNote)
			if notes == "" {
				notes = "-"
			}
			b.WriteString(m.styles.Subtle.Render(fmt.Sprintf(
				"%s | %s | %s | %s | note: %s | privacy: %s",
				row.Name, row.Email, roomLabel, row.BookingStatus, notes, row.PrivacySetting,
			)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.editingSearch {
		b.WriteString(m.styles.Selected.Render(fmt.Sprintf("Search name: %s_", m.searchInput)))
		b.WriteString("\n\n")
	}
	b.WriteString(common.RenderFooter(m.styles, m.footer()))
	b.WriteString("\n")

	return b.String()
}

func (m *Model) footer() []string {
	if m.editingSearch {
		return []string{"type: edit search", "enter: apply search", "esc: cancel", "q: quit"}
	}
	return []string{"r: cycle room type", "b: cycle booking status", "s: toggle special requests", "/: search name", "q: quit"}
}

func nextCycledValue(current string, options []string) string {
	cur := normalizeQueryOption(current)
	for i := range options {
		if options[i] == cur {
			return options[(i+1)%len(options)]
		}
	}
	return options[0]
}

func normalizeQueryOption(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "all" {
		return ""
	}
	return trimmed
}

func emptyAsAll(value string) string {
	if strings.TrimSpace(value) == "" {
		return "all"
	}
	return value
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

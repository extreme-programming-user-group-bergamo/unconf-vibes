package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockListClient struct {
	listConferencesFn func(ctx context.Context) ([]client.ConferenceResponse, error)
}

func (m *mockListClient) ListConferences(ctx context.Context) ([]client.ConferenceResponse, error) {
	if m.listConferencesFn != nil {
		return m.listConferencesFn(ctx)
	}
	return nil, nil
}

func sampleConferences() []client.ConferenceResponse {
	return []client.ConferenceResponse{
		{
			ID:            1,
			Slug:          "gophercon-2026",
			Name:          "GopherCon 2026",
			Description:   "Go conference",
			Location:      "Denver, CO",
			StartDate:     "2026-06-15",
			EndDate:       "2026-06-18",
			Capacity:      500,
			AttendeeCount: 120,
			Status:        "upcoming",
		},
		{
			ID:            2,
			Slug:          "rustconf-2026",
			Name:          "RustConf 2026",
			Description:   "Rust conference",
			Location:      "Portland, OR",
			StartDate:     "2026-08-01",
			EndDate:       "2026-08-03",
			Capacity:      300,
			AttendeeCount: 50,
			Status:        "active",
		},
		{
			ID:            3,
			Slug:          "pycon-2025",
			Name:          "PyCon 2025",
			Description:   "Python conference",
			Location:      "Pittsburgh, PA",
			StartDate:     "2025-04-10",
			EndDate:       "2025-04-13",
			Capacity:      400,
			AttendeeCount: 350,
			Status:        "past",
		},
	}
}

func TestListCmd_HappyPath(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return sampleConferences(), nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "DATES")
	assert.Contains(t, output, "LOCATION")
	assert.Contains(t, output, "ATTENDEES")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "GopherCon 2026")
	assert.Contains(t, output, "RustConf 2026")
	assert.NotContains(t, output, "PyCon 2025") // past filtered out by default
}

func TestListCmd_AllFlagIncludesPast(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return sampleConferences(), nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "GopherCon 2026")
	assert.Contains(t, output, "RustConf 2026")
	assert.Contains(t, output, "PyCon 2025")
}

func TestListCmd_DefaultFilterHidesPast(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{ID: 1, Name: "Past Event", StartDate: "2024-01-01", EndDate: "2024-01-03", Status: "past", Location: "NYC"},
				{ID: 2, Name: "Upcoming Event", StartDate: "2026-09-01", EndDate: "2026-09-03", Status: "upcoming", Location: "LA"},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.NotContains(t, output, "Past Event")
	assert.Contains(t, output, "Upcoming Event")
}

func TestListCmd_SortByDateAscending(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{ID: 1, Name: "Later", StartDate: "2026-12-01", EndDate: "2026-12-03", Status: "upcoming", Location: "NYC"},
				{ID: 2, Name: "Earlier", StartDate: "2026-03-01", EndDate: "2026-03-03", Status: "upcoming", Location: "LA"},
				{ID: 3, Name: "Middle", StartDate: "2026-07-01", EndDate: "2026-07-03", Status: "upcoming", Location: "SF"},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	earlierIdx := bytes.Index([]byte(output), []byte("Earlier"))
	middleIdx := bytes.Index([]byte(output), []byte("Middle"))
	laterIdx := bytes.Index([]byte(output), []byte("Later"))
	assert.Less(t, earlierIdx, middleIdx)
	assert.Less(t, middleIdx, laterIdx)
}

func TestListCmd_EmptyWithAll(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "No conferences available.")
}

func TestListCmd_EmptyWithoutAll(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "No upcoming conferences. Use --all to show past conferences.")
}

func TestListCmd_EmptyAfterFilterWithoutAll(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{ID: 1, Name: "Past Only", StartDate: "2024-01-01", EndDate: "2024-01-03", Status: "past", Location: "NYC"},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "No upcoming conferences. Use --all to show past conferences.")
}

func TestListCmd_ClientError(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	cmd := newListCmd(mock)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list conferences")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestListCmd_DateFormatting(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{
					ID:            1,
					Name:          "TestConf",
					StartDate:     "2026-06-15",
					EndDate:       "2026-06-18",
					Status:        "upcoming",
					Location:      "Denver",
					AttendeeCount: 10,
				},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Jun 15 - Jun 18, 2026")
}

func TestListCmd_SingleConference(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{
					ID:            1,
					Name:          "Solo Conf",
					StartDate:     "2026-01-10",
					EndDate:       "2026-01-12",
					Status:        "upcoming",
					Location:      "Austin, TX",
					AttendeeCount: 42,
				},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Solo Conf")
	assert.Contains(t, output, "Jan 10 - Jan 12, 2026")
	assert.Contains(t, output, "Austin, TX")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "upcoming")
}

func TestListCmd_CrossYearDateFormatting(t *testing.T) {
	mock := &mockListClient{
		listConferencesFn: func(_ context.Context) ([]client.ConferenceResponse, error) {
			return []client.ConferenceResponse{
				{
					ID:            1,
					Name:          "New Year Conf",
					StartDate:     "2026-12-30",
					EndDate:       "2027-01-02",
					Status:        "upcoming",
					Location:      "Berlin",
					AttendeeCount: 75,
				},
			}, nil
		},
	}

	cmd := newListCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Dec 30, 2026 - Jan 02, 2027")
}

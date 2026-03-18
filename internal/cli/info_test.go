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

type mockInfoClient struct {
	getConferenceFn func(ctx context.Context, slug string) (*client.ConferenceResponse, error)
}

func (m *mockInfoClient) GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error) {
	if m.getConferenceFn != nil {
		return m.getConferenceFn(ctx, slug)
	}
	return nil, nil
}

func sampleConference() *client.ConferenceResponse {
	return &client.ConferenceResponse{
		ID:            1,
		Slug:          "socrates-26",
		Name:          "SoCraTes 2026",
		Description:   "Software Craftsmanship and Testing Conference",
		Location:      "Saarbrücken, Germany",
		StartDate:     "2026-10-07",
		EndDate:       "2026-10-10",
		Capacity:      200,
		AttendeeCount: 42,
		Status:        "upcoming",
	}
}

func TestInfoCmd_HappyPath(t *testing.T) {
	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, slug string) (*client.ConferenceResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			return sampleConference(), nil
		},
	}

	cmd := newInfoCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "SoCraTes 2026")
	assert.Contains(t, output, "Software Craftsmanship and Testing Conference")
	assert.Contains(t, output, "Saarbrücken, Germany")
	assert.Contains(t, output, "Oct 07 - Oct 10, 2026")
	assert.Contains(t, output, "upcoming")
	assert.Contains(t, output, "200")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "Rooms")
	assert.Contains(t, output, "No room information available yet.")
}

func TestInfoCmd_NotFound(t *testing.T) {
	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return nil, client.ErrConferenceNotFound
		},
	}

	cmd := newInfoCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conference info")
	assert.Contains(t, stderr.String(), "not found")
	assert.Contains(t, stderr.String(), "unconf list")
}

func TestInfoCmd_ClientError(t *testing.T) {
	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	cmd := newInfoCmd(mock)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conference info")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestInfoCmd_DateFormatting(t *testing.T) {
	conf := sampleConference()
	conf.StartDate = "2026-06-15"
	conf.EndDate = "2026-06-18"

	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return conf, nil
		},
	}

	cmd := newInfoCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Jun 15 - Jun 18, 2026")
}

func TestInfoCmd_OutputStructure(t *testing.T) {
	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return sampleConference(), nil
		},
	}

	cmd := newInfoCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	// Verify section headers
	assert.Contains(t, output, "═")
	assert.Contains(t, output, "Rooms")
	assert.Contains(t, output, "─")
	// Verify field labels
	assert.Contains(t, output, "Location:")
	assert.Contains(t, output, "Dates:")
	assert.Contains(t, output, "Status:")
	assert.Contains(t, output, "Capacity:")
	assert.Contains(t, output, "Attendees:")
}

func TestInfoCmd_MissingArgument(t *testing.T) {
	mock := &mockInfoClient{}

	cmd := newInfoCmd(mock)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestInfoCmd_AllFieldsDisplayed(t *testing.T) {
	conf := &client.ConferenceResponse{
		ID:            99,
		Slug:          "test-conf",
		Name:          "Test Conference",
		Description:   "A test conference",
		Location:      "Test City",
		StartDate:     "2026-03-01",
		EndDate:       "2026-03-03",
		Capacity:      500,
		AttendeeCount: 123,
		Status:        "active",
	}

	mock := &mockInfoClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return conf, nil
		},
	}

	cmd := newInfoCmd(mock)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"test-conf"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "500")
	assert.Contains(t, output, "123")
	assert.Contains(t, output, "active")
	assert.Contains(t, output, "Test City")
	assert.Contains(t, output, "Test Conference")
}

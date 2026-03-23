package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerStartup_InitializesDatabaseAndServesHealth(t *testing.T) {
	moduleRoot := mustFindModuleRoot(t)
	port := mustAllocatePort(t)
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "integration.db")
	apiEndpoint := fmt.Sprintf("http://127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	serverBinary := filepath.Join(tempDir, "unconf-server")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", serverBinary, "./cmd/server")
	buildCmd.Dir = moduleRoot
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build server binary: %s", string(buildOutput))

	serverCmd := exec.CommandContext(ctx, serverBinary)
	serverCmd.Dir = moduleRoot
	serverCmd.Env = append(os.Environ(),
		"UNCONF_API_ENDPOINT="+apiEndpoint,
		"UNCONF_DB_PATH="+dbPath,
		"UNCONF_LOG_LEVEL=debug",
		"UNCONF_GITHUB_CLIENT_ID=test-client-id",
		"UNCONF_GITHUB_CLIENT_SECRET=test-client-secret",
		"UNCONF_PASETO_SYMMETRIC_KEY=0123456789abcdef0123456789abcdef",
	)

	stopped := false
	stopServer := func() {
		if stopped {
			return
		}
		stopped = true
		cancel()
		waitForProcessExit(serverCmd, 3*time.Second)
	}
	t.Cleanup(stopServer)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	serverCmd.Stdout = &stdoutBuffer
	serverCmd.Stderr = &stderrBuffer

	require.NoError(t, serverCmd.Start())

	readyURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	require.NoError(t, waitForHealthy(ctx, readyURL))

	body, statusCode, err := readHealthResponse(readyURL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Contains(t, body, `"status":"ok"`)

	_, err = os.Stat(dbPath)
	require.NoError(t, err)

	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'users'`).Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "users", tableName)

	stopServer()

	t.Logf("server stdout: %s", stdoutBuffer.String())
	t.Logf("server stderr: %s", stderrBuffer.String())
}

func waitForProcessExit(cmd *exec.Cmd, timeout time.Duration) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
}

func waitForHealthy(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 1 * time.Second}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for health endpoint: %w", ctx.Err())
		case <-ticker.C:
			resp, err := client.Get(url)
			if err != nil {
				continue
			}

			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func readHealthResponse(url string) (string, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("failed to call health endpoint: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read health response: %w", err)
	}

	return string(body), resp.StatusCode, nil
}

func mustAllocatePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return addr.Port
}

func mustFindModuleRoot(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	t.Fatalf("failed to locate module root from %q", cwd)
	return ""
}

const testSymmetricKey = "0123456789abcdef0123456789abcdef"

func TestUpdateUserProfile_PersistsChanges(t *testing.T) {
	moduleRoot := mustFindModuleRoot(t)
	port := mustAllocatePort(t)
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "integration.db")
	apiEndpoint := fmt.Sprintf("http://127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	serverBinary := filepath.Join(tempDir, "unconf-server")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", serverBinary, "./cmd/server")
	buildCmd.Dir = moduleRoot
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build server binary: %s", string(buildOutput))

	serverCmd := exec.CommandContext(ctx, serverBinary)
	serverCmd.Dir = moduleRoot
	serverCmd.Env = append(os.Environ(),
		"UNCONF_API_ENDPOINT="+apiEndpoint,
		"UNCONF_DB_PATH="+dbPath,
		"UNCONF_LOG_LEVEL=debug",
		"UNCONF_GITHUB_CLIENT_ID=test-client-id",
		"UNCONF_GITHUB_CLIENT_SECRET=test-client-secret",
		"UNCONF_PASETO_SYMMETRIC_KEY="+testSymmetricKey,
	)

	stopped := false
	stopServer := func() {
		if stopped {
			return
		}
		stopped = true
		cancel()
		waitForProcessExit(serverCmd, 3*time.Second)
	}
	t.Cleanup(stopServer)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	serverCmd.Stdout = &stdoutBuffer
	serverCmd.Stderr = &stderrBuffer

	require.NoError(t, serverCmd.Start())

	readyURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	require.NoError(t, waitForHealthy(ctx, readyURL))

	// Insert test user directly into DB
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`INSERT INTO users (id, github_id, email, display_name, privacy_setting) VALUES (?, ?, ?, ?, ?)`,
		1, "gh-99999", "test@example.com", "Original Name", "public")
	require.NoError(t, err)

	// Insert a refresh session so the token's session ID is valid
	_, err = db.Exec(`INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at, last_access_jti) VALUES (?, ?, ?, ?, ?)`,
		1, 1, "testhash", time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339), "test-jti")
	require.NoError(t, err)

	// Generate a valid PASETO access token
	tokenSvc, err := auth.NewTokenService(testSymmetricKey)
	require.NoError(t, err)

	accessToken, err := tokenSvc.IssueAccessToken(ctx, auth.AccessTokenInput{
		UserID:     1,
		SessionID:  1,
		Issuer:     "unconf-api",
		Audience:   "unconf-cli",
		JTI:        "test-jti-1",
		ExpiryTime: time.Now().Add(1 * time.Hour),
	})
	require.NoError(t, err)

	// Step 1: PUT /users/me with updated fields
	updateBody, _ := json.Marshal(map[string]string{
		"display_name":    "Updated Name",
		"privacy_setting": "private",
	})

	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, apiEndpoint+"/users/me", bytes.NewReader(updateBody))
	require.NoError(t, err)
	putReq.Header.Set("Authorization", "Bearer "+accessToken)
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := http.DefaultClient.Do(putReq)
	require.NoError(t, err)
	defer func() { _ = putResp.Body.Close() }()

	putBody, err := io.ReadAll(putResp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, putResp.StatusCode, "PUT /users/me response: %s", string(putBody))

	var putResult map[string]any
	require.NoError(t, json.Unmarshal(putBody, &putResult))
	assert.Equal(t, "Updated Name", putResult["display_name"])
	assert.Equal(t, "private", putResult["privacy_setting"])

	// Step 2: GET /users/me to verify persistence
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"/users/me", nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+accessToken)

	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	defer func() { _ = getResp.Body.Close() }()

	getBody, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "GET /users/me response: %s", string(getBody))

	var getResult map[string]any
	require.NoError(t, json.Unmarshal(getBody, &getResult))
	assert.Equal(t, "Updated Name", getResult["display_name"])
	assert.Equal(t, "private", getResult["privacy_setting"])

	stopServer()

	t.Logf("server stdout: %s", stdoutBuffer.String())
	t.Logf("server stderr: %s", stderrBuffer.String())
}

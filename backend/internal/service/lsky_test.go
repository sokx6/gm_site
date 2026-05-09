package service

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFile implements multipart.File for testing.
type fakeFile struct {
	*bytes.Reader
}

func (f *fakeFile) Close() error { return nil }

// ──────────────────────────────────────────────
// Test 1: NewClient initializes fields correctly
// ──────────────────────────────────────────────

func TestLskyClient_NewClient(t *testing.T) {
	client := NewLskyClient("https://lsky.example.com", "user@example.com", "pa$$word")

	assert.Equal(t, "https://lsky.example.com", client.baseURL)
	assert.Equal(t, "user@example.com", client.email)
	assert.Equal(t, "pa$$word", client.password)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	assert.Empty(t, client.token)
}

// ──────────────────────────────────────
// Test 2: GetToken fetches and stores token
// ──────────────────────────────────────

func TestLskyClient_GetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/tokens", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req map[string]string
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "user@example.com", req["email"])
		assert.Equal(t, "pa$$word", req["password"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			Status: true,
			Data: struct {
				Token string `json:"token"`
			}{
				Token: "1|test-token-abc",
			},
		})
	}))
	defer server.Close()

	client := NewLskyClient(server.URL, "user@example.com", "pa$$word")
	err := client.ensureToken()
	require.NoError(t, err)
	assert.Equal(t, "1|test-token-abc", client.token)
}

// ──────────────────────────────────────
// Test 3: Upload sends file and returns URL
// ──────────────────────────────────────

func TestLskyClient_Upload(t *testing.T) {
	var (
		tokenCalled  atomic.Bool
		uploadCalled atomic.Bool
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/tokens":
			tokenCalled.Store(true)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse{
				Status: true,
				Data: struct {
					Token string `json:"token"`
				}{
					Token: "1|upload-test-token",
				},
			})

		case "/api/v1/upload":
			uploadCalled.Store(true)
			assert.Equal(t, "Bearer 1|upload-test-token", r.Header.Get("Authorization"))

			err := r.ParseMultipartForm(10 << 20)
			require.NoError(t, err)

			f, fh, err := r.FormFile("file")
			require.NoError(t, err)
			defer f.Close()

			assert.Equal(t, "test.png", fh.Filename)

			content, err := io.ReadAll(f)
			require.NoError(t, err)
			assert.Equal(t, "fake-image-bytes", string(content))

			w.Header().Set("Content-Type", "application/json")
			var resp uploadResponse
			resp.Status = true
			resp.Data.Links.URL = "https://images.example.com/uploads/test.png"
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewLskyClient(server.URL, "user@example.com", "pa$$word")
	file := &fakeFile{bytes.NewReader([]byte("fake-image-bytes"))}
	header := &multipart.FileHeader{Filename: "test.png"}

	url, err := client.UploadImage(file, header)
	require.NoError(t, err)
	assert.True(t, tokenCalled.Load())
	assert.True(t, uploadCalled.Load())
	assert.Equal(t, "https://images.example.com/uploads/test.png", url)
}

// ───────────────────────────────────────────────────
// Test 4: Upload with 401 triggers token refresh retry
// ───────────────────────────────────────────────────

func TestLskyClient_Upload_TokenRefresh(t *testing.T) {
	var (
		tokenReqs  atomic.Int32
		uploadReqs atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/tokens":
			tokenReqs.Add(1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse{
				Status: true,
				Data: struct {
					Token string `json:"token"`
				}{
					Token: "1|refreshed-token",
				},
			})

		case "/api/v1/upload":
			current := uploadReqs.Add(1)
			if current == 1 {
				// First upload: 401 triggers token refresh
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// Second upload: succeeds with refreshed token
			assert.Equal(t, "Bearer 1|refreshed-token", r.Header.Get("Authorization"))

			var resp uploadResponse
			resp.Status = true
			resp.Data.Links.URL = "https://images.example.com/uploads/retry.png"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewLskyClient(server.URL, "user@example.com", "pa$$word")
	file := &fakeFile{bytes.NewReader([]byte("retry-image-content"))}
	header := &multipart.FileHeader{Filename: "retry-test.png"}

	url, err := client.UploadImage(file, header)
	require.NoError(t, err)
	assert.Equal(t, "https://images.example.com/uploads/retry.png", url)
	assert.Equal(t, int32(2), tokenReqs.Load(), "token should be fetched twice")
	assert.Equal(t, int32(2), uploadReqs.Load(), "upload should be attempted twice")
}

// ─────────────────────────────────────────────
// Test 5: Upload returns errors correctly
// ─────────────────────────────────────────────

func TestLskyClient_Upload_Error(t *testing.T) {
	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/tokens":
				var resp tokenResponse
				resp.Status = true
				resp.Data.Token = "1|error-test-token"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			case "/api/v1/upload":
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			}
		}))
		defer server.Close()

		client := NewLskyClient(server.URL, "user@example.com", "pa$$word")
		file := &fakeFile{bytes.NewReader([]byte("test"))}
		header := &multipart.FileHeader{Filename: "test.png"}

		url, err := client.UploadImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
		assert.Empty(t, url)
	})

	t.Run("response status false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/tokens":
				var resp tokenResponse
				resp.Status = true
				resp.Data.Token = "1|error-test-token"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			case "/api/v1/upload":
				var resp uploadResponse
				resp.Status = false
				resp.Message = "file type not allowed"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}
		}))
		defer server.Close()

		client := NewLskyClient(server.URL, "user@example.com", "pa$$word")
		file := &fakeFile{bytes.NewReader([]byte("test"))}
		header := &multipart.FileHeader{Filename: "test.png"}

		url, err := client.UploadImage(file, header)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file type not allowed")
		assert.Empty(t, url)
	})
}

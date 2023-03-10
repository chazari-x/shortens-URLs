package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"main/internal/app/handlers"
	"main/internal/pkg/shortens"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader([]byte(body)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	respHeader := resp.Request.URL

	defer resp.Body.Close()

	switch method {
	case "GET":
		return resp.StatusCode, respHeader.String()
	default:
		return resp.StatusCode, string(respBody)
	}
}

func TestServer(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/{id}", handlers.Get)
	r.Post("/", handlers.Post)
	ts := httptest.NewServer(r)
	defer ts.Close()

	var urls = []string{"https://pkg.go.dev/net/http@go1.17.2",
		"https://pkg.go.dev/net/http@go1.17.2",
		"https://github.com/chazari-x/shortens-URLs/pull/2",
		"https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
	}

	var n = 0
	for i := 0; i < 25; i++ {
		statusCode, actual := testRequest(t, ts, "POST", "/", urls[n])
		assert.Equal(t, http.StatusCreated, statusCode)
		assert.Equal(t, "http://localhost:8080/"+shortens.Short(i), actual)

		statusCode, actual = testRequest(t, ts, "GET", "/"+shortens.Short(i), "")
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, urls[n], actual)

		if n == 3 {
			n = 0
		} else {
			n++
		}
	}

}

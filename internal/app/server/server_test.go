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
)

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader([]byte(body)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	respHeader := resp.Request.URL

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		require.NoError(t, err)
	}(resp.Body)

	switch method {
	case "GET":
		return resp.StatusCode, respHeader.String()
	default:
		return resp.StatusCode, string(respBody)
	}
}

func TestServer(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/*", handlers.Get)
	r.Post("/", handlers.Post)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, actual := testRequest(t, ts, "POST", "/", "https://stackoverflow.com/questions/13896592/how-to-convert-url-url-to-string-in-go-google-app-engine")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "http://localhost:8080/0", actual)

	statusCode, actual = testRequest(t, ts, "POST", "/", "https://github.com/chazari-x?tab=overview&from=2023-03-01&to=2023-03-07")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "http://localhost:8080/1", actual)

	statusCode, actual = testRequest(t, ts, "POST", "/", "https://github.com/chazari-x/shortens-URLs/pull/2")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "http://localhost:8080/2", actual)

	statusCode, actual = testRequest(t, ts, "POST", "/", "https://github.com/golang-standards/project-layout/blob/master/README_ru.md")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "http://localhost:8080/3", actual)

	statusCode, actual = testRequest(t, ts, "GET", "/0", "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "https://stackoverflow.com/questions/13896592/how-to-convert-url-url-to-string-in-go-google-app-engine", actual)

	statusCode, actual = testRequest(t, ts, "GET", "/1", "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "https://github.com/chazari-x?tab=overview&from=2023-03-01&to=2023-03-07", actual)

	statusCode, actual = testRequest(t, ts, "GET", "/2", "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "https://github.com/chazari-x/shortens-URLs/pull/2", actual)

	statusCode, actual = testRequest(t, ts, "GET", "/3", "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "https://github.com/golang-standards/project-layout/blob/master/README_ru.md", actual)
}

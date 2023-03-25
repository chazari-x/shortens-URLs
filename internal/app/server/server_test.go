package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"main/internal/app/config"
	"main/internal/app/handlers"
)

type (
	short struct {
		Result string `json:"result"`
	}

	some struct {
		URL string `json:"url"`
	}
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
	c := config.GetConfig()

	r := chi.NewRouter()

	if c.BaseURL != "" {
		r.Get("/"+c.BaseURL+"/{id}", handlers.Get)
	} else {
		r.Get("/{id}", handlers.Get)
	}
	r.Post("/", handlers.Post)
	r.Post("/api/shorten", handlers.Shorten)
	ts := httptest.NewServer(r)
	defer ts.Close()

	var urls = []string{"https://pkg.go.dev/net/http@go1.17.2",
		"https://pkg.go.dev/net/http@go1.17.2",
		"https://github.com/chazari-x/shortens-URLs/pull/2",
		"https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
	}

	var n = 0
	for i := 0; i < 25; i += 2 {
		var expectedOne string
		var expectedTwo string
		var pathOne string
		var pathTwo string
		if c.BaseURL != "" {
			expectedOne = "http://" + c.ServerAddress + "/" + c.BaseURL + "/" + strconv.FormatInt(int64(i), 36)
			marshal, err := json.Marshal(short{Result: "http://" + c.ServerAddress + "/" + c.BaseURL + "/" + strconv.FormatInt(int64(i+1), 36)})
			expectedTwo = string(marshal)
			if err != nil {
				log.Fatal(err)
			}
			pathOne = "/" + c.BaseURL + "/" + strconv.FormatInt(int64(i), 36)
			pathTwo = "/" + c.BaseURL + "/" + strconv.FormatInt(int64(i+1), 36)
		} else {
			expectedOne = "http://" + c.ServerAddress + "/" + strconv.FormatInt(int64(i), 36)
			marshal, err := json.Marshal(short{Result: "http://" + c.ServerAddress + "/" + strconv.FormatInt(int64(i+1), 36)})
			expectedTwo = string(marshal)
			if err != nil {
				log.Fatal(err)
			}
			pathOne = "/" + strconv.FormatInt(int64(i), 36)
			pathTwo = "/" + strconv.FormatInt(int64(i+1), 36)
		}

		statusCode, actual := testRequest(t, ts, "POST", "/", urls[n])
		assert.Equal(t, http.StatusCreated, statusCode)
		assert.Equal(t, expectedOne, actual)

		url, err := json.Marshal(some{URL: urls[n]})
		if err != nil {
			log.Fatal(err)
		}
		statusCode, actual = testRequest(t, ts, "POST", "/api/shorten", string(url))
		assert.Equal(t, http.StatusCreated, statusCode)
		assert.Equal(t, expectedTwo, actual)

		statusCode, actual = testRequest(t, ts, "GET", pathOne, "")
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, urls[n], actual)

		statusCode, actual = testRequest(t, ts, "GET", pathTwo, "")
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, urls[n], actual)

		if n == 3 {
			n = 0
		} else {
			n++
		}
	}
}

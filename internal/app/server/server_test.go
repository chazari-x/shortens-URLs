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
	"main/internal/app/storage"
)

type (
	short struct {
		Result string `json:"result"`
	}

	original struct {
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

	defer func() {
		_ = resp.Body.Close()
	}()

	switch method {
	case "GET":
		return resp.StatusCode, respHeader.String()
	default:
		return resp.StatusCode, string(respBody)
	}
}

func TestServer(t *testing.T) {
	conf, err := config.ParseConfig()
	if err != nil {
		log.Print("parse config err: ", err)
	}

	sModel, err := storage.StartStorage(conf)
	if err != nil {
		log.Print("start storage file path err: ", err)
	}

	c := handlers.NewController(sModel, conf, sModel.DB)

	r := chi.NewRouter()
	r.Get("/"+conf.BaseURL+"{id}", c.Get)
	r.Get("/api/user/urls", c.UserURLs)
	r.Post("/", c.Post)
	r.Post("/api/shorten", c.Shorten)
	ts := httptest.NewServer(handlers.MiddlewaresConveyor(r))
	defer ts.Close()

	var urls = []string{"https://m.vk.com/login?slogin_h=9c4b5dff2b9d2ec030.187f50f7956785726a&role=fast&to=ZmVlZA--",
		"https://ok.ru/dk?st.cmd=anonymMain",
		"https://www.google.ru/",
		"https://github.com/chazari-x/shortens-URLs/actions/runs/4631562598/jobs/8194566021?pr=9",
	}

	var n = 0
	for i := 0; i < 25; i += 2 {
		expectedOne := "http://" + conf.ServerAddress + conf.BaseURL + strconv.FormatInt(int64(i), 36)
		marshal, err := json.Marshal(short{Result: "http://" + conf.ServerAddress + conf.BaseURL + strconv.FormatInt(int64(i+1), 36)})
		expectedTwo := string(marshal)
		if err != nil {
			log.Fatal(err)
		}
		pathOne := "/" + conf.BaseURL + strconv.FormatInt(int64(i), 36)
		pathTwo := "/" + conf.BaseURL + strconv.FormatInt(int64(i+1), 36)

		statusCode, actual := testRequest(t, ts, "POST", "/", urls[n])
		assert.Equal(t, http.StatusCreated, statusCode)
		assert.Equal(t, expectedOne, actual)

		url, err := json.Marshal(original{URL: urls[n]})
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

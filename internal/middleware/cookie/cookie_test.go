package cookie

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
)

func init() {
	config.Config.SecureCookieHashKey = "very-secret"
	config.Config.SecureCookieBlockKey = "alotsecretalotsecretalotsecretgr"
	config.Config.SecureCookieExpire = 3600
}

func TestMiddleware_NewUserCookie(t *testing.T) {
	middleware := New()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(string(constants.XUserID))

		assert.NotEmpty(t, userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	cookies := resp.Cookies()
	require.NotEmpty(t, cookies)

	var userCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "user" {
			userCookie = c
			break
		}
	}
	require.NotNil(t, userCookie)
	assert.NotEmpty(t, userCookie.Value)
}

func TestMiddleware_ExistingUserCookie(t *testing.T) {
	middleware := New()
	s := securecookie.New([]byte(config.Config.SecureCookieHashKey), []byte(config.Config.SecureCookieBlockKey))

	userID := uuid.New().String()
	encoded, err := s.Encode("user", user{UserID: userID})
	require.NoError(t, err)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, userID, r.Header.Get(string(constants.XUserID)))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "user", Value: encoded})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMiddleware_InvalidUserCookie(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "user", Value: "invalid-cookie-value"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestMiddleware_RequestToUserUrlsWithoutCookie(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, userUrls, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error(fmt.Sprintf("failed to close body: %v", err))
		}
	}()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

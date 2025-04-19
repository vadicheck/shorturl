// Package cookie provides a middleware to handle secure cookies for user authentication.
// It checks for the presence of a "user" cookie, creates a new one if absent, and decodes
// the cookie to retrieve the user information (UserID) for the request context.
package cookie

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

// user represents the structure of the user information stored in the cookie.
type user struct {
	UserID string `json:"user_id"`
}

const userUrls = "/api/user/urls"

// New returns a middleware function for secure cookie authentication.
//
// The middleware checks if the "user" cookie exists in the incoming request. If the cookie
// is missing, a new "user" cookie is created with a new unique UserID. If the cookie exists,
// it is decoded, and the UserID is extracted and added to the request's header.
//
// The UserID is set in the `X-User-ID` header for the downstream handlers to use. If the cookie
// is absent, the middleware sets a new cookie in the response and continues processing the request.
//
// If there is an error while decoding the cookie or creating a new one, an error response is sent
// to the client with an appropriate status code.
//
// Parameters:
//   - None (this is a middleware factory function)
//
// Returns:
//   - A middleware function that can be used with `http.Handle` or other HTTP routers.
func New() func(next http.Handler) http.Handler {
	slog.Info("cookie middleware enabled")

	var (
		s = securecookie.New(
			[]byte(config.Config.SecureCookieHashKey),
			[]byte(config.Config.SecureCookieBlockKey),
		)
	)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userCookie, err := r.Cookie("user")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				slog.Error("\"user\" cookie not found")
				httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
				return
			}

			if userCookie == nil || errors.Is(err, http.ErrNoCookie) {
				user := &user{
					UserID: uuid.New().String(),
				}

				encoded, errCookieEncode := s.Encode("user", user)
				if errCookieEncode != nil {
					slog.Error("can't build secure cookie", sl.Err(errCookieEncode))
					httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
					return
				}

				cookie := &http.Cookie{
					Name:   "user",
					Value:  encoded,
					Path:   "/",
					MaxAge: int(config.Config.SecureCookieExpire.Seconds()),
				}
				http.SetCookie(w, cookie)

				r.Header.Set(string(constants.XUserID), user.UserID)

				if userUrls == r.URL.String() {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNoContent)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			u := &user{}
			if err = s.Decode("user", userCookie.Value, u); err != nil {
				slog.Error("can't decode user cookie", sl.Err(err))
				httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
				return
			}

			if u.UserID == "" {
				slog.Error("user_id is absent in cookie")
				httpError.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			r.Header.Set(string(constants.XUserID), u.UserID)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

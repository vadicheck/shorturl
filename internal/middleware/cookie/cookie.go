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

type user struct {
	UserID string `json:"user_id"`
}

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

				encoded, err := s.Encode("user", user)
				if err != nil {
					slog.Error("can't build secure cookie", sl.Err(err))
					httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
					return
				}

				cookie := &http.Cookie{
					Name:     "user",
					Value:    encoded,
					Path:     "/",
					Secure:   true,
					HttpOnly: true,
					MaxAge:   int(config.Config.SecureCookieExpire.Seconds()),
				}
				http.SetCookie(w, cookie)

				r.Header.Set(string(constants.XUserID), user.UserID)
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

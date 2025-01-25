package jwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

const userUrls = "/api/user/urls"

func New() func(next http.Handler) http.Handler {
	slog.Info("jwt middleware enabled")

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			jwtCookie, err := r.Cookie("jwt")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				slog.Error("jwt cookie not found")
				httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
				return
			}

			if userUrls == r.URL.String() && (jwtCookie == nil || errors.Is(err, http.ErrNoCookie)) {
				httpError.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if jwtCookie == nil || errors.Is(err, http.ErrNoCookie) {
				token, userID, err := buildJWTString(config.Config.JwtSecret, config.Config.JwtTokenExpire)
				if err != nil {
					slog.Error("can't build jwt token", sl.Err(err))
					httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
					return
				}

				cookie := &http.Cookie{
					Name:     "jwt",
					Value:    token,
					Path:     "/",
					Domain:   "",
					Secure:   false,
					HttpOnly: true,
					MaxAge:   int(config.Config.JwtTokenExpire.Seconds()),
				}

				http.SetCookie(w, cookie)

				newCtx := context.WithValue(r.Context(), constants.ContextUserID, userID)
				next.ServeHTTP(w, r.WithContext(newCtx))
				return
			}

			jwtToken := jwtCookie.Value

			decodedJwtToken, err := decodeJwtToken(jwtToken, config.Config.JwtSecret)
			if err != nil {
				slog.Error("can't decode jwt token", sl.Err(err))
				httpError.RespondWithError(w, http.StatusInternalServerError, "Auth error")
				return
			}

			if decodedJwtToken.UserID == "" {
				slog.Error("user_id is absent in jwt")
				httpError.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			newCtx := context.WithValue(r.Context(), constants.ContextUserID, decodedJwtToken.UserID)
			next.ServeHTTP(w, r.WithContext(newCtx))
		}

		return http.HandlerFunc(fn)
	}
}

func buildJWTString(jwtSecret string, jwtTokenExpire time.Duration) (tokenString, userID string, err error) {
	userID = uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTokenExpire)),
		},
		UserID: userID,
	})

	tokenString, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("can't sign token: %w", err)
	}

	return tokenString, userID, nil
}

func decodeJwtToken(jwtToken, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return token.Claims.(*Claims), nil
}

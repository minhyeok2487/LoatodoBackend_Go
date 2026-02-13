package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const UsernameKey contextKey = "username"

// permitAllPaths are paths that don't require authentication.
// Matches WebSecurityConfig.PERMIT_ALL_LINK from Java.
var permitAllPaths = []string{
	"/auth/",
	"/oauth2/",
	"/v3/mail/",
	"/v3/notification/",
	"/v3/auth/signup",
	"/v3/auth/character",
	"/v3/auth/login",
	"/v2/boards/",
	"/v3/boards/",
	"/v2/comments",
	"/v3/home/test",
	"/auth/authorize",
	"/login/oauth2/",
	"/api/v1/auth/signup",
	"/api/v1/member/character",
	"/api/v1/auth/login",
	"/api/v1/mail/",
	"/manage/health",
	"/manage/info",
	"/manage/prometheus",
	"/api/v1/games/",
	"/api/v1/events/",
	"/api/v1/suggestions/",
}

// permitGetPaths are paths that allow unauthenticated GET requests.
var permitGetPaths = []string{
	"/v3/comments",
	"/api/v1/recruting-board/",
	"/api/v1/community/",
}

// MemberGetter looks up a member's role by username.
type MemberGetter interface {
	GetRole(ctx context.Context, username string) (string, error)
}

// JwtMiddleware creates a Chi middleware for JWT authentication.
func JwtMiddleware(tp *TokenProvider, memberGetter MemberGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := parseBearerToken(r)

			if token != "" && token != "null" {
				// Token present - validate it
				username, err := tp.ValidateToken(token)
				if err != nil {
					if isPermitAllPath(r) {
						// Invalid token on permit-all path - continue without auth
						ctx := context.WithValue(r.Context(), UsernameKey, "")
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
					sendErrorResponse(w, http.StatusForbidden, "유효하지 않은 토큰입니다.")
					return
				}

				// Check admin role for /admin paths
				if strings.HasPrefix(r.URL.Path, "/admin") {
					role, err := memberGetter.GetRole(r.Context(), username)
					if err != nil || role != "ADMIN" {
						sendErrorResponse(w, http.StatusForbidden, "관리자 권한이 필요합니다.")
						return
					}
				}

				ctx := context.WithValue(r.Context(), UsernameKey, username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No token
			if !isPermitAllPath(r) {
				sendErrorResponse(w, http.StatusForbidden, "인증이 필요한 서비스입니다.")
				return
			}

			ctx := context.WithValue(r.Context(), UsernameKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUsername extracts the authenticated username from the request context.
func GetUsername(r *http.Request) string {
	if v, ok := r.Context().Value(UsernameKey).(string); ok {
		return v
	}
	return ""
}

func parseBearerToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.HasPrefix(bearer, "Bearer ") {
		return bearer[7:]
	}
	return ""
}

func isPermitAllPath(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	for _, p := range permitAllPaths {
		if strings.HasSuffix(p, "/") {
			if strings.HasPrefix(path, p) {
				return true
			}
		} else if path == p {
			return true
		}
	}

	if method == http.MethodGet {
		for _, p := range permitGetPaths {
			if strings.HasSuffix(p, "/") {
				if strings.HasPrefix(path, p) {
					return true
				}
			} else if path == p {
				return true
			}
		}
	}

	return false
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

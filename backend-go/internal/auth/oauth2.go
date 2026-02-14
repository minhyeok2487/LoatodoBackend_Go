package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo represents the user info returned by Google's userinfo endpoint.
type GoogleUserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// OAuth2Handler handles Google OAuth2 login flow.
type OAuth2Handler struct {
	oauthConfig   *oauth2.Config
	tokenProvider *TokenProvider
	db            *sql.DB
}

// NewOAuth2Handler creates a new OAuth2Handler.
func NewOAuth2Handler(clientID, clientSecret, redirectURI string, tp *TokenProvider, db *sql.DB) *OAuth2Handler {
	return &OAuth2Handler{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		},
		tokenProvider: tp,
		db:            db,
	}
}

// HandleAuthorize saves redirect_url cookie and redirects to Google OAuth2 authorization URL.
// GET /auth/authorize?redirect_url=...
func (h *OAuth2Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "http://localhost:3000"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "redirect_url",
		Value:    redirectURL,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.oauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback exchanges the authorization code for a token, gets user info,
// finds or creates a member, generates a JWT, and redirects to the frontend.
// GET /login/oauth2/code/google?code=...
func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("oauth2 code exchange failed", "error", err)
		http.Error(w, "failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := h.getUserInfo(token)
	if err != nil {
		slog.Error("failed to get user info", "error", err)
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}

	// Find or create member
	email := userInfo.Email
	memberID, err := h.findOrCreateMember(r.Context(), email)
	if err != nil {
		slog.Error("failed to find or create member", "error", err, "email", email)
		http.Error(w, "failed to process user", http.StatusInternalServerError)
		return
	}

	_ = memberID

	// Generate JWT
	jwt, err := h.tokenProvider.CreateToken(email)
	if err != nil {
		slog.Error("failed to create JWT", "error", err)
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	// Get redirect_url from cookie
	redirectURL := "http://localhost:3000"
	if cookie, err := r.Cookie("redirect_url"); err == nil && cookie.Value != "" {
		redirectURL = cookie.Value
	}

	// Clear the redirect_url cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "redirect_url",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Redirect to frontend with token
	targetURL := fmt.Sprintf("%s/sociallogin?token=%s&username=%s", redirectURL, jwt, email)
	http.Redirect(w, r, targetURL, http.StatusTemporaryRedirect)
}

// getUserInfo fetches user info from Google using the OAuth2 token.
func (h *OAuth2Handler) getUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := h.oauthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("requesting user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request returned status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decoding user info: %w", err)
	}

	return &userInfo, nil
}

// findOrCreateMember looks up a member by username (email). If not found, creates a new one.
func (h *OAuth2Handler) findOrCreateMember(ctx context.Context, email string) (int64, error) {
	var memberID int64
	err := h.db.QueryRowContext(ctx,
		`SELECT member_id FROM member WHERE username = ?`, email,
	).Scan(&memberID)

	if err == nil {
		// Member exists, update provider if needed
		_, _ = h.db.ExecContext(ctx,
			`UPDATE member SET auth_provider = 'google', last_modified_date = ? WHERE member_id = ?`,
			time.Now(), memberID,
		)
		return memberID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("querying member: %w", err)
	}

	// Create new member
	now := time.Now()
	result, err := h.db.ExecContext(ctx,
		`INSERT INTO member (username, role, auth_provider, created_date, last_modified_date)
		 VALUES (?, 'USER', 'google', ?, ?)`,
		email, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting member: %w", err)
	}

	return result.LastInsertId()
}

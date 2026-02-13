package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"lostark-todo-backend/internal/auth"
)

// AuthService handles authentication business logic.
type AuthService struct {
	db            *sql.DB
	tokenProvider *auth.TokenProvider
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *sql.DB, tp *auth.TokenProvider) *AuthService {
	return &AuthService{db: db, tokenProvider: tp}
}

// SignUpRequest is the request body for signup.
type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginMemberRequest is the request body for login.
type LoginMemberRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse is the response for signup and login.
type AuthResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// SignUp creates a new member with bcrypt password.
func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) (*AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("아이디와 비밀번호를 입력해주세요.")
	}

	// Check if username already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM member WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("checking username: %w", err)
	}
	if exists {
		return nil, errors.New("이미 사용중인 아이디 입니다.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO member (username, password, role, auth_provider, created_date, last_modified_date)
		 VALUES (?, ?, 'USER', 'none', ?, ?)`,
		req.Username, string(hashedPassword), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting member: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	token, err := s.tokenProvider.CreateToken(req.Username)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}

	return &AuthResponse{
		ID:       id,
		Username: req.Username,
		Token:    token,
	}, nil
}

// Login authenticates a member and returns a JWT.
func (s *AuthService) Login(ctx context.Context, req LoginMemberRequest) (*AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("아이디와 비밀번호를 입력해주세요.")
	}

	var id int64
	var hashedPassword string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, password FROM member WHERE username = ?", req.Username,
	).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("등록되지 않은 아이디 입니다.")
		}
		return nil, fmt.Errorf("querying member: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, errors.New("비밀번호가 일치하지 않습니다.")
	}

	token, err := s.tokenProvider.CreateToken(req.Username)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}

	return &AuthResponse{
		ID:       id,
		Username: req.Username,
		Token:    token,
	}, nil
}

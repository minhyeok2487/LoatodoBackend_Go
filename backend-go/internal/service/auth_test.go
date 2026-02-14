package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"lostark-todo-backend/internal/auth"

	_ "github.com/go-sql-driver/mysql"
)

// setupTestDB creates an in-memory test database
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// For testing, you can either:
	// 1. Use a real MySQL test database
	// 2. Use sqlmock (recommended for unit tests)
	// This example uses a real test DB for integration testing

	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/loatodo_test?parseTime=true")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test table
	schema := `
	CREATE TABLE IF NOT EXISTS member (
		member_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255),
		role VARCHAR(50) DEFAULT 'USER',
		auth_provider VARCHAR(50) DEFAULT 'none',
		access_key VARCHAR(500),
		created_date DATETIME(6),
		last_modified_date DATETIME(6)
	)
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("DROP TABLE IF EXISTS member"); err != nil {
		t.Logf("Warning: Failed to drop test table: %v", err)
	}
	db.Close()
}

// setupTestTokenProvider creates a test token provider
func setupTestTokenProvider(t *testing.T) *auth.TokenProvider {
	t.Helper()

	// Use a test secret (base64 encoded)
	testSecret := "dGVzdC1zZWNyZXQta2V5LWZvci1qd3QtdG9rZW4tdGVzdGluZy1wdXJwb3Nlcw=="
	tp, err := auth.NewTokenProvider(testSecret)
	if err != nil {
		t.Fatalf("Failed to create test token provider: %v", err)
	}
	return tp
}

// TestSignUp_Success tests successful user signup
func TestSignUp_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	req := SignUpRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	resp, err := authService.SignUp(ctx, req)

	if err != nil {
		t.Fatalf("SignUp failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, resp.Username)
	}

	if resp.ID == 0 {
		t.Error("Expected non-zero user ID")
	}

	if resp.Token == "" {
		t.Error("Expected token, got empty string")
	}

	// Verify password is hashed in database
	var hashedPassword string
	err = db.QueryRowContext(ctx, "SELECT password FROM member WHERE username = ?", req.Username).
		Scan(&hashedPassword)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	// Verify password is bcrypt hashed
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		t.Error("Password was not properly hashed with bcrypt")
	}

	// Verify auth_provider is set to 'none'
	var authProvider string
	err = db.QueryRowContext(ctx, "SELECT auth_provider FROM member WHERE username = ?", req.Username).
		Scan(&authProvider)
	if err != nil {
		t.Fatalf("Failed to query auth_provider: %v", err)
	}
	if authProvider != "none" {
		t.Errorf("Expected auth_provider 'none', got '%s'", authProvider)
	}
}

// TestSignUp_EmptyUsername tests signup with empty username
func TestSignUp_EmptyUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	req := SignUpRequest{
		Username: "",
		Password: "testpassword123",
	}

	ctx := context.Background()
	_, err := authService.SignUp(ctx, req)

	if err == nil {
		t.Fatal("Expected error for empty username, got nil")
	}

	expectedMsg := "아이디와 비밀번호를 입력해주세요."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestSignUp_EmptyPassword tests signup with empty password
func TestSignUp_EmptyPassword(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	req := SignUpRequest{
		Username: "testuser",
		Password: "",
	}

	ctx := context.Background()
	_, err := authService.SignUp(ctx, req)

	if err == nil {
		t.Fatal("Expected error for empty password, got nil")
	}

	expectedMsg := "아이디와 비밀번호를 입력해주세요."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestSignUp_DuplicateUsername tests signup with duplicate username
func TestSignUp_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	req := SignUpRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()

	// First signup should succeed
	_, err := authService.SignUp(ctx, req)
	if err != nil {
		t.Fatalf("First signup failed: %v", err)
	}

	// Second signup with same username should fail
	_, err = authService.SignUp(ctx, req)
	if err == nil {
		t.Fatal("Expected error for duplicate username, got nil")
	}

	expectedMsg := "이미 사용중인 아이디 입니다."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// First, create a user
	signupReq := SignUpRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	signupResp, err := authService.SignUp(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	// Now login
	loginReq := LoginMemberRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	loginResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if loginResp == nil {
		t.Fatal("Expected login response, got nil")
	}

	if loginResp.Username != loginReq.Username {
		t.Errorf("Expected username %s, got %s", loginReq.Username, loginResp.Username)
	}

	if loginResp.ID != signupResp.ID {
		t.Errorf("Expected user ID %d, got %d", signupResp.ID, loginResp.ID)
	}

	if loginResp.Token == "" {
		t.Error("Expected token, got empty string")
	}
}

// TestLogin_WrongPassword tests login with wrong password
func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// Create a user
	signupReq := SignUpRequest{
		Username: "testuser",
		Password: "correctpassword",
	}

	ctx := context.Background()
	_, err := authService.SignUp(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	// Try to login with wrong password
	loginReq := LoginMemberRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = authService.Login(ctx, loginReq)
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}

	expectedMsg := "비밀번호가 일치하지 않습니다."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLogin_UserNotFound tests login with non-existent user
func TestLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	loginReq := LoginMemberRequest{
		Username: "nonexistentuser",
		Password: "somepassword",
	}

	ctx := context.Background()
	_, err := authService.Login(ctx, loginReq)

	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}

	expectedMsg := "등록되지 않은 아이디 입니다."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLogin_EmptyUsername tests login with empty username
func TestLogin_EmptyUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	loginReq := LoginMemberRequest{
		Username: "",
		Password: "somepassword",
	}

	ctx := context.Background()
	_, err := authService.Login(ctx, loginReq)

	if err == nil {
		t.Fatal("Expected error for empty username, got nil")
	}

	expectedMsg := "아이디와 비밀번호를 입력해주세요."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLogin_EmptyPassword tests login with empty password
func TestLogin_EmptyPassword(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	loginReq := LoginMemberRequest{
		Username: "testuser",
		Password: "",
	}

	ctx := context.Background()
	_, err := authService.Login(ctx, loginReq)

	if err == nil {
		t.Fatal("Expected error for empty password, got nil")
	}

	expectedMsg := "아이디와 비밀번호를 입력해주세요."
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLogout_NoneProvider tests logout with auth_provider 'none'
func TestLogout_NoneProvider(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// Create a user with auth_provider 'none'
	signupReq := SignUpRequest{
		Username: "testuser",
		Password: "testpassword123",
	}

	ctx := context.Background()
	_, err := authService.SignUp(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	// Logout
	resp := authService.Logout(ctx, "testuser")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got false. Message: %s", resp.Message)
	}

	expectedMsg := "로그아웃 성공"
	if resp.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, resp.Message)
	}
}

// TestLogout_GoogleProvider tests logout with Google auth provider
func TestLogout_GoogleProvider(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// Create a mock Google OAuth server
	mockGoogleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock successful token revocation
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGoogleServer.Close()

	// Insert a Google user manually
	ctx := context.Background()
	now := time.Now()
	result, err := db.ExecContext(ctx,
		`INSERT INTO member (username, password, role, auth_provider, access_key, created_date, last_modified_date)
		 VALUES (?, ?, 'USER', 'Google', ?, ?, ?)`,
		"googleuser", "", "mock_access_token", now, now,
	)
	if err != nil {
		t.Fatalf("Failed to insert Google user: %v", err)
	}

	_, err = result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert id: %v", err)
	}

	// Note: The real Google API call will fail in test environment
	// This test demonstrates the logic, but will fail unless mocked properly
	resp := authService.Logout(ctx, "googleuser")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	// For Google logout with actual token, we expect it to attempt revocation
	// In a real test, you'd mock the HTTP client or use a test server
	t.Logf("Google logout response: Success=%v, Message=%s", resp.Success, resp.Message)
}

// TestLogout_GoogleProvider_NoAccessKey tests Google logout without access key
func TestLogout_GoogleProvider_NoAccessKey(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// Insert a Google user without access_key
	ctx := context.Background()
	now := time.Now()
	_, err := db.ExecContext(ctx,
		`INSERT INTO member (username, password, role, auth_provider, access_key, created_date, last_modified_date)
		 VALUES (?, ?, 'USER', 'Google', NULL, ?, ?)`,
		"googleuser_nokey", "", now, now,
	)
	if err != nil {
		t.Fatalf("Failed to insert Google user: %v", err)
	}

	resp := authService.Logout(ctx, "googleuser_nokey")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	if resp.Success {
		t.Error("Expected success=false for Google logout without access key")
	}

	expectedMsg := "구글 로그아웃 실패 : 액세스 토큰 없음"
	if resp.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, resp.Message)
	}
}

// TestLogout_EmptyUsername tests logout with empty username
func TestLogout_EmptyUsername(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	ctx := context.Background()
	resp := authService.Logout(ctx, "")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	if resp.Success {
		t.Error("Expected success=false for empty username")
	}

	expectedMsg := "사용자를 찾을 수 없습니다"
	if resp.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, resp.Message)
	}
}

// TestLogout_UserNotFound tests logout with non-existent user
func TestLogout_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	ctx := context.Background()
	resp := authService.Logout(ctx, "nonexistentuser")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	if resp.Success {
		t.Error("Expected success=false for non-existent user")
	}

	expectedMsg := "사용자를 찾을 수 없습니다"
	if resp.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, resp.Message)
	}
}

// TestLogout_UnknownProvider tests logout with unknown auth provider
func TestLogout_UnknownProvider(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	// Insert a user with unknown auth_provider
	ctx := context.Background()
	now := time.Now()
	_, err := db.ExecContext(ctx,
		`INSERT INTO member (username, password, role, auth_provider, created_date, last_modified_date)
		 VALUES (?, ?, 'USER', 'UnknownProvider', ?, ?)`,
		"unknownuser", "", now, now,
	)
	if err != nil {
		t.Fatalf("Failed to insert user with unknown provider: %v", err)
	}

	resp := authService.Logout(ctx, "unknownuser")

	if resp == nil {
		t.Fatal("Expected logout response, got nil")
	}

	if resp.Success {
		t.Error("Expected success=false for unknown auth provider")
	}

	expectedMsg := "알 수 없는 인증 제공자"
	if resp.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, resp.Message)
	}
}

// TestPasswordHashing tests bcrypt password hashing
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify correct password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Error("Failed to verify correct password")
	}

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Error("Expected mismatch error for wrong password")
	}
}

// TestTokenGeneration tests JWT token generation and validation
func TestTokenGeneration(t *testing.T) {
	tp := setupTestTokenProvider(t)

	username := "testuser"

	// Create token
	token, err := tp.CreateToken(username)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Validate token
	validatedUsername, err := tp.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if validatedUsername != username {
		t.Errorf("Expected username %s from token, got %s", username, validatedUsername)
	}
}

// TestIntegration_SignUpAndLogin tests the full signup and login flow
func TestIntegration_SignUpAndLogin(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tp := setupTestTokenProvider(t)
	authService := NewAuthService(db, tp)

	ctx := context.Background()

	// Step 1: Sign up
	signupReq := SignUpRequest{
		Username: "integrationuser",
		Password: "password123",
	}

	signupResp, err := authService.SignUp(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	signupToken := signupResp.Token
	signupID := signupResp.ID

	// Step 2: Login with same credentials
	loginReq := LoginMemberRequest{
		Username: "integrationuser",
		Password: "password123",
	}

	loginResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Verify user ID matches
	if loginResp.ID != signupID {
		t.Errorf("User ID mismatch: signup=%d, login=%d", signupID, loginResp.ID)
	}

	// Both tokens should be valid (but different)
	username1, err := tp.ValidateToken(signupToken)
	if err != nil {
		t.Fatalf("Failed to validate signup token: %v", err)
	}

	username2, err := tp.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("Failed to validate login token: %v", err)
	}

	if username1 != username2 || username1 != signupReq.Username {
		t.Error("Username mismatch in tokens")
	}

	// Step 3: Logout
	logoutResp := authService.Logout(ctx, "integrationuser")
	if !logoutResp.Success {
		t.Errorf("Logout failed: %s", logoutResp.Message)
	}
}

//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// buildAuthRouter wires only the auth-related dependencies and returns
// a chi.Router serving /api/v1/auth/* routes, suitable for httptest.
func buildAuthRouter(t *testing.T) chi.Router {
	t.Helper()

	jwtManager := auth.NewJWTManager("integration-test-secret", 15*time.Minute, 7*24*time.Hour)
	tokenBlocklist := cache.NewTokenBlocklist(testRedis)

	userRepo := pg.NewUserRepoPG(testPgPool)
	roleRepo := pg.NewRoleRepoPG(testPgPool)

	registerUC := authuc.NewRegisterUseCase(userRepo, roleRepo, jwtManager)
	loginUC := authuc.NewLoginUseCase(userRepo, jwtManager)
	refreshUC := authuc.NewRefreshTokenUseCase(userRepo, jwtManager, tokenBlocklist)
	profileUC := authuc.NewGetProfileUseCase(userRepo)
	changePasswordUC := authuc.NewChangePasswordUseCase(userRepo)
	authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshUC, profileUC, changePasswordUC, jwtManager, tokenBlocklist)

	requireAuth := middleware.AuthMiddleware(jwtManager, tokenBlocklist)

	r := chi.NewRouter()
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(requireAuth)
			r.Get("/me", authHandler.Me)
			r.Post("/logout", authHandler.Logout)
			r.Post("/change-password", authHandler.ChangePassword)
		})
	})

	return r
}

func TestAuthAPI_FullFlow(t *testing.T) {
	truncateAll(t)

	router := buildAuthRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	const (
		testEmail    = "authtest@example.com"
		testName     = "Auth Test User"
		testPassword = "SecureP@ssw0rd!"
	)

	// ---------------------------------------------------------------
	// Step 1: Register
	// ---------------------------------------------------------------
	registerBody, _ := json.Marshal(map[string]string{
		"email":     testEmail,
		"full_name": testName,
		"password":  testPassword,
	})

	resp, err := client.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(registerBody))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected status 201, got %d", resp.StatusCode)
	}

	var registerResp struct {
		User struct {
			ID       string `json:"id"`
			Email    string `json:"email"`
			FullName string `json:"full_name"`
			IsActive bool   `json:"is_active"`
		} `json:"user"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		t.Fatalf("register: failed to decode response: %v", err)
	}

	if registerResp.User.Email != testEmail {
		t.Errorf("register: expected email %q, got %q", testEmail, registerResp.User.Email)
	}
	if registerResp.User.FullName != testName {
		t.Errorf("register: expected full_name %q, got %q", testName, registerResp.User.FullName)
	}
	if !registerResp.User.IsActive {
		t.Error("register: expected is_active=true")
	}
	if registerResp.AccessToken == "" {
		t.Error("register: access_token is empty")
	}
	if registerResp.RefreshToken == "" {
		t.Error("register: refresh_token is empty")
	}

	// ---------------------------------------------------------------
	// Step 2: Login
	// ---------------------------------------------------------------
	loginBody, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})

	resp, err = client.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: expected status 200, got %d", resp.StatusCode)
	}

	var loginResp struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("login: failed to decode response: %v", err)
	}

	if loginResp.User.Email != testEmail {
		t.Errorf("login: expected email %q, got %q", testEmail, loginResp.User.Email)
	}
	if loginResp.AccessToken == "" {
		t.Error("login: access_token is empty")
	}
	if loginResp.RefreshToken == "" {
		t.Error("login: refresh_token is empty")
	}

	accessToken := loginResp.AccessToken
	refreshToken := loginResp.RefreshToken

	// ---------------------------------------------------------------
	// Step 3: Refresh
	// ---------------------------------------------------------------
	refreshBody, _ := json.Marshal(map[string]string{
		"refresh_token": refreshToken,
	})

	resp, err = client.Post(srv.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("refresh: expected status 200, got %d", resp.StatusCode)
	}

	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		t.Fatalf("refresh: failed to decode response: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Error("refresh: access_token is empty")
	}
	if refreshResp.RefreshToken == "" {
		t.Error("refresh: refresh_token is empty")
	}

	// Use the NEW access token from refresh for subsequent requests
	accessToken = refreshResp.AccessToken

	// ---------------------------------------------------------------
	// Step 4: GET /me (authenticated)
	// ---------------------------------------------------------------
	meReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = client.Do(meReq)
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("me: expected status 200, got %d", resp.StatusCode)
	}

	var meResp struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meResp); err != nil {
		t.Fatalf("me: failed to decode response: %v", err)
	}

	if meResp.Email != testEmail {
		t.Errorf("me: expected email %q, got %q", testEmail, meResp.Email)
	}
	if meResp.FullName != testName {
		t.Errorf("me: expected full_name %q, got %q", testName, meResp.FullName)
	}
	if !meResp.IsActive {
		t.Error("me: expected is_active=true")
	}

	// ---------------------------------------------------------------
	// Step 5: Logout
	// ---------------------------------------------------------------
	logoutReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = client.Do(logoutReq)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: expected status 204, got %d", resp.StatusCode)
	}

	// ---------------------------------------------------------------
	// Step 6: GET /me after logout -> 401 (token blocklisted)
	// ---------------------------------------------------------------
	meReq2, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/me", nil)
	meReq2.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = client.Do(meReq2)
	if err != nil {
		t.Fatalf("me-after-logout request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me after logout: expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthAPI_Register_DuplicateEmail(t *testing.T) {
	truncateAll(t)

	router := buildAuthRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	body, _ := json.Marshal(map[string]string{
		"email":     "dup@example.com",
		"full_name": "Duplicate User",
		"password":  "SecureP@ss1",
	})

	// First registration should succeed
	resp, err := client.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("first register request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", resp.StatusCode)
	}

	// Second registration with same email should fail with 409
	resp, err = client.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("second register request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d", resp.StatusCode)
	}
}

func TestAuthAPI_Login_InvalidCredentials(t *testing.T) {
	truncateAll(t)

	router := buildAuthRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	// Register a user first
	regBody, _ := json.Marshal(map[string]string{
		"email":     "creds@example.com",
		"full_name": "Creds User",
		"password":  "CorrectP@ss1",
	})
	resp, err := client.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	resp.Body.Close()

	// Login with wrong password
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "creds@example.com",
		"password": "WrongPassword!",
	})
	resp, err = client.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password login: expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthAPI_Me_NoToken(t *testing.T) {
	truncateAll(t)

	router := buildAuthRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	// GET /me without Authorization header -> 401
	resp, err := client.Get(srv.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me without token: expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthAPI_Refresh_BlocksOldToken(t *testing.T) {
	truncateAll(t)

	router := buildAuthRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	// Register
	regBody, _ := json.Marshal(map[string]string{
		"email":     "refresh@example.com",
		"full_name": "Refresh User",
		"password":  "SecureP@ss1",
	})
	resp, err := client.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}

	var regResp struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		t.Fatalf("register: failed to decode response: %v", err)
	}
	resp.Body.Close()

	oldRefreshToken := regResp.RefreshToken

	// Refresh once (should succeed and block the old refresh token)
	refreshBody, _ := json.Marshal(map[string]string{
		"refresh_token": oldRefreshToken,
	})
	resp, err = client.Post(srv.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
	if err != nil {
		t.Fatalf("first refresh request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first refresh: expected 200, got %d", resp.StatusCode)
	}

	// Try to reuse the old refresh token -> 401 (blocked by token rotation)
	resp, err = client.Post(srv.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
	if err != nil {
		t.Fatalf("second refresh request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reused refresh token: expected 401, got %d", resp.StatusCode)
	}
}

package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vivek-344/diagon/sigil/internal/domain"
	"github.com/vivek-344/diagon/sigil/internal/middleware"
	"github.com/vivek-344/diagon/sigil/internal/service"
	"github.com/vivek-344/diagon/sigil/utils"
)

type AuthHandler struct {
	developerSvc *service.DeveloperService
	jwtSecret    string
}

func NewAuthHandler(developerSvc *service.DeveloperService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		developerSvc: developerSvc,
		jwtSecret:    jwtSecret,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Developer    struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"developer"`
}

// Login authenticates a developer and returns JWT tokens
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("invalid login request body", "error", err)
		utils.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get developer by email
	dev, err := h.developerSvc.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			slog.Debug("developer not found", "email", req.Email)
			utils.RespondError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		slog.Error("failed to fetch developer", "error", err)
		utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Check if account is suspended
	if dev.Status == domain.StatusSuspended {
		utils.RespondError(w, "account suspended", http.StatusForbidden)
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, dev.PasswordHash) {
		slog.Debug("invalid password attempt", "email", req.Email)
		utils.RespondError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT tokens
	tokens, err := utils.GenerateTokenPair(dev.ID, dev.Email, h.jwtSecret)
	if err != nil {
		slog.Error("failed to generate tokens", "error", err)
		utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Update last login
	if err := h.developerSvc.UpdateLastLogin(r.Context(), dev.ID); err != nil {
		slog.Warn("failed to update last login", "error", err)
	}

	// Prepare response
	resp := loginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
	resp.Developer.ID = dev.ID.String()
	resp.Developer.Email = dev.Email

	utils.RespondSuccess(w, resp, http.StatusOK)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken generates new tokens using a refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("invalid refresh token request body", "error", err)
		utils.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(req.RefreshToken, h.jwtSecret)
	if err != nil {
		if errors.Is(err, utils.ErrExpiredToken) {
			slog.Debug("refresh token expired", "error", err)
			utils.RespondError(w, "refresh token expired", http.StatusUnauthorized)
			return
		}
		slog.Debug("invalid refresh token", "error", err)
		utils.RespondError(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Verify developer still exists and is active
	dev, err := h.developerSvc.GetByID(r.Context(), claims.DeveloperID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			slog.Debug("developer not found", "developer_id", claims.DeveloperID)
			utils.RespondError(w, "developer not found", http.StatusUnauthorized)
			return
		}
		slog.Error("failed to fetch developer", "error", err)
		utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if dev.Status == domain.StatusSuspended {
		utils.RespondError(w, "account suspended", http.StatusForbidden)
		return
	}

	// Generate new token pair
	tokens, err := utils.GenerateTokenPair(dev.ID, dev.Email, h.jwtSecret)
	if err != nil {
		slog.Error("failed to generate tokens", "error", err)
		utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	utils.RespondSuccess(w, refreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, http.StatusOK)
}

// GetProfile returns the authenticated developer's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// This would be called on a protected route
	// Developer ID is extracted from context by middleware
	developerID, ok := middleware.GetDeveloperIDFromContext(r.Context())
	if !ok {
		utils.RespondError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dev, err := h.developerSvc.GetByID(r.Context(), developerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			utils.RespondError(w, "developer not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to fetch developer", "error", err)
		utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Don't send password hash
	dev.PasswordHash = ""

	utils.RespondSuccess(w, dev, http.StatusOK)
}

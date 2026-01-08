package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vivek-344/diagon/sigil/internal/domain"
	"github.com/vivek-344/diagon/sigil/internal/service"
	"github.com/vivek-344/diagon/sigil/utils"
)

type DeveloperHandler struct {
	svc *service.DeveloperService
}

func NewDeveloperHandler(svc *service.DeveloperService) *DeveloperHandler {
	return &DeveloperHandler{svc: svc}
}

type createRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FullName    *string `json:"full_name,omitempty"`
	CompanyName *string `json:"company_name,omitempty"`
}

type createResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *DeveloperHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dev, err := h.svc.Create(r.Context(), domain.CreateDeveloperInput{
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		CompanyName: req.CompanyName,
	}, "")
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailExists):
			slog.Debug("email already registered", "email", req.Email)
			utils.RespondError(w, "email already registered", http.StatusConflict)
		case errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrWeakPassword) || errors.Is(err, domain.ErrShortPassword):
			slog.Debug("validation error", "error", err)
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
		default:
			slog.Error("failed to create developer", "error", err)
			utils.RespondError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	utils.RespondSuccess(w, createResponse{
		ID:    dev.ID.String(),
		Email: dev.Email,
	}, http.StatusCreated)
}

func (h *DeveloperHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) Update(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) UpdateLastLogin(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) AddMetadata(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) Delete(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

func (h *DeveloperHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	utils.RespondError(w, "not implemented", http.StatusNotImplemented)
}

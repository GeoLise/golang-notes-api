package handlers

import (
	"encoding/json"
	"net/http"
	"notes-api/internal/middlewares"
	"notes-api/internal/models"
	"notes-api/internal/respond"
	"notes-api/internal/storage"
	"strconv"
)

type Notes struct {
	repo           *storage.Notes
	authMiddleware *middlewares.AuthMiddleware
}

func NewNotes(repo *storage.Notes, authMiddleware *middlewares.AuthMiddleware) *Notes {
	return &Notes{repo: repo, authMiddleware: authMiddleware}
}

func (h *Notes) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /notes", h.Get)
	mux.HandleFunc("POST /notes", h.Create)
	mux.HandleFunc("PUT /notes/{id}", h.Update)
}

func (h *Notes) Get(w http.ResponseWriter, r *http.Request) {

	user, err := h.authMiddleware.GetSession(r)

	if err != nil {
		respond.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	notes, err := h.repo.Get(r.Context(), user.Id)

	if err != nil {
		respond.ServerError(w, "get motes by user", err)
		return
	}

	respond.JSON(w, http.StatusOK, notes)
}

func (h *Notes) Create(w http.ResponseWriter, r *http.Request) {

	var n models.Note

	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		respond.Error(w, http.StatusBadRequest, "Wrong json")
		return
	}

	if err := n.Validate(); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authMiddleware.GetSession(r)

	n.UserId = user.Id

	if err != nil {
		respond.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), n)

	if err != nil {
		respond.ServerError(w, "creating note", err)
		return
	}

	respond.JSON(w, http.StatusCreated, created)

}

func (h *Notes) Update(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		respond.Error(w, http.StatusMethodNotAllowed, err.Error())
	}

	var n models.Note

	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		respond.Error(w, http.StatusBadRequest, "Wrong json")
		return
	}

	if err := n.Validate(); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.authMiddleware.GetSession(r); err != nil {
		respond.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, n)

	if err != nil {
		respond.ServerError(w, "updating note", err)
		return
	}

	respond.JSON(w, http.StatusCreated, updated)

}

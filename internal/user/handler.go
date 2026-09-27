package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"notes-api/internal/respond"
	"strconv"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.create)
	mux.HandleFunc("GET /users", h.getAll)
	mux.HandleFunc("GET /users/{id}", h.getById)
	mux.HandleFunc("PUT /users/{id}", h.update)
	mux.HandleFunc("DELETE /users/{id}", h.delete)

}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respond.Error(w, http.StatusBadRequest, "JSON говно прислали")
		return
	}

	if err := u.Validate(); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), u)

	if err != nil {
		respond.ServerError(w, "create user", err)
		return
	}

	respond.JSON(w, http.StatusCreated, created)
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAll(r.Context())
	if err != nil {
		respond.ServerError(w, "get all users", err)
	}

	respond.JSON(w, http.StatusOK, users)
}

func (h *Handler) getById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Wrong id")
		return
	}

	user, err := h.repo.GetById(r.Context(), id)

	if errors.Is(err, ErrorNotFound) {
		respond.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		respond.ServerError(w, "get user by id", err)
		return
	}

	respond.JSON(w, http.StatusOK, user)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Wrong id")
		return
	}

	var u User

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respond.Error(w, http.StatusBadRequest, "JSON говно прислали")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, u)

	if errors.Is(err, ErrorNotFound) {
		respond.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		respond.ServerError(w, "update user", err)
		return
	}

	respond.JSON(w, http.StatusOK, updated)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Wrong id")
		return
	}

	deleted, err := h.repo.Delete(r.Context(), id)

	if errors.Is(err, ErrorNotFound) {
		respond.Error(w, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		respond.ServerError(w, "delete user", err)
		return
	}

	respond.JSON(w, http.StatusOK, deleted)
}

package handlers

import (
	"fmt"
	"net/http"
	"notes-api/internal/auth"
	"notes-api/internal/models"
	"notes-api/internal/respond"
	"notes-api/internal/storage"
	"os"
	"time"
)

type Auth struct {
	sessionsRepo   *storage.Sessions
	logincodesRepo *storage.LoginCodes
}

func NewAuth(sessionsRepo *storage.Sessions, logincodesRepo *storage.LoginCodes) *Auth {
	return &Auth{sessionsRepo: sessionsRepo, logincodesRepo: logincodesRepo}
}

func (h *Auth) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/get-auth-link", h.GetAuthLink)
	mux.HandleFunc("GET /auth/confirm", h.Confirm)
	mux.HandleFunc("GET /auth/me", h.Me)

}

func (h *Auth) GetAuthLink(w http.ResponseWriter, r *http.Request) {
	code, err := h.logincodesRepo.Generate(r.Context())
	if err != nil {
		respond.ServerError(w, "Неизвестная ошибка", err)
		return
	}

	link := fmt.Sprintf("https://t.me/%s?start=%s", os.Getenv("TELEGRAM_BOT_NAME"), code)

	respond.JSON(w, http.StatusOK, map[string]string{"link": link})
}

func (h *Auth) Confirm(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		respond.Error(w, http.StatusBadRequest, "Не передан код авторизации")
		return
	}

	val, err := h.logincodesRepo.IsPending(r.Context(), code)

	if err != nil || val == true {
		respond.Error(w, http.StatusConflict, "Код еще не активирован")
		return
	}

	userId, err := h.logincodesRepo.GetUserIdByCode(r.Context(), code)

	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	sessionToken := auth.NewToken()

	_, err = h.sessionsRepo.Create(r.Context(), models.Session{
		UserId:    userId,
		TokenHash: auth.HashToken(sessionToken),
		ExpiresAt: time.Now().Add(time.Hour * 24),
	})

	if err != nil {
		respond.ServerError(w, "create session", err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"token": sessionToken})

}

func (h *Auth) Me(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	if token == "" {
		respond.Error(w, http.StatusUnauthorized, "No token")
		return
	}

	user, err := h.sessionsRepo.GetByToken(r.Context(), token)

	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	respond.JSON(w, http.StatusOK, user)
}

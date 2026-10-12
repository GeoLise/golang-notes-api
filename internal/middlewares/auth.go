package middlewares

import (
	"net/http"
	"notes-api/internal/models"
	"notes-api/internal/storage"
)

type AuthMiddleware struct {
	sessions *storage.Sessions
}

func NewAuthMiddleWare(sessionsStorage *storage.Sessions) *AuthMiddleware {
	return &AuthMiddleware{sessions: sessionsStorage}
}

func (m *AuthMiddleware) GetSession(r *http.Request) (models.User, error) {

	token := r.Header.Get("Authorization")

	user, err := m.sessions.GetByToken(r.Context(), token)

	if err != nil {
		return models.User{}, err
	}

	return user, nil

}

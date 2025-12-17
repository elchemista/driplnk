package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	handler "github.com/elchemista/driplnk/internal/adapters/http"
	"github.com/elchemista/driplnk/internal/domain"
	"github.com/elchemista/driplnk/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_UpdateProfile(t *testing.T) {
	mockUsers := mocks.NewMockUserRepository()
	mockSessions := mocks.NewMockSessionManager()

	h := handler.NewUserHandler(mockUsers, mockSessions, nil)

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{
			ID:     "user-1",
			Handle: "original",
			Email:  "test@example.com",
		}
		mockUsers.AddUser(user)
		mockSessions.SetCurrentUser("user-1")

		form := url.Values{}
		form.Add("handle", "newhandle")
		form.Add("title", "New Name")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.UpdateProfile(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		// Verify update in repo
		updated, _ := mockUsers.GetByID(context.Background(), "user-1")
		assert.Equal(t, "newhandle", updated.Handle)
		assert.Equal(t, "New Name", updated.Title)
	})

	t.Run("HandleTaken", func(t *testing.T) {
		user := &domain.User{ID: "user-1", Handle: "original", Email: "u1@example.com"}
		otherUser := &domain.User{ID: "user-2", Handle: "taken", Email: "u2@example.com"}
		mockUsers.AddUser(user)
		mockUsers.AddUser(otherUser)
		mockSessions.SetCurrentUser("user-1")

		form := url.Values{}
		form.Add("handle", "taken")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.UpdateProfile(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestUserHandler_UpdateTheme(t *testing.T) {
	mockUsers := mocks.NewMockUserRepository()
	mockSessions := mocks.NewMockSessionManager()

	h := handler.NewUserHandler(mockUsers, mockSessions, nil)

	t.Run("Success", func(t *testing.T) {
		user := &domain.User{
			ID:     "user-1",
			Handle: "original",
			Email:  "test@example.com",
		}
		mockUsers.AddUser(user)
		mockSessions.SetCurrentUser("user-1")

		form := url.Values{}
		form.Add("primary_color", "#000000")
		form.Add("layout", "grid")

		req := httptest.NewRequest(http.MethodPost, "/dashboard/theme", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.UpdateTheme(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)

		updated, _ := mockUsers.GetByID(context.Background(), "user-1")
		assert.Equal(t, "#000000", updated.Theme.PrimaryColor)
		assert.Equal(t, "grid", updated.Theme.LayoutStyle)
	})
}

package handlers

import (
	"errors"
	"fmt"
	"golang-vigilate-project/internal/helpers"
	"golang-vigilate-project/internal/models"
	"log"
	"net/http"
	"strings"
	"time"
)

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	OK                   bool               `json:"ok"`
	AccessToken          string             `json:"access_token"`
	AccessTokenExpiresAt time.Time          `json:"access_token_expires_at"`
	RefreshToken         string             `json:"refresh_token"`
	RefreshTokenExpires  time.Time          `json:"refresh_token_expires"`
	User                 createUserResponse `json:"user"`
	SessionID            string             `json:"session_id"`
}

type createUserResponse struct {
	Username         string    `json:"username"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	PasswordChangeAt time.Time `json:"password_change_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// Login attempts to log the user in
func (repo *DBRepo) Login(w http.ResponseWriter, r *http.Request) {
	var req loginUserRequest
	var resp loginUserResponse
	var errResp jsonResp

	err := helpers.ReadJSONBody(r, &req)
	if err != nil {
		errResp.Message = "Invalid JSON"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
	}

	id, _, err := repo.DB.Authenticate(req.Email, req.Password)
	if errors.Is(err, models.ErrInvalidCredentials) {
		errResp.Message = "Invalid username or password"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	} else if errors.Is(err, models.ErrInactiveAccount) {
		errResp.Message = "Inactive account"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	} else if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	}

	user, err := repo.DB.GetUserById(id)
	if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	}
	duration := 45 * time.Minute
	accessToken, accTokenPayload, err := repo.TokenMaker.CreateToken(user.Email, "ADMIN", duration)
	if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	}

	refreshToken, refTokenPayload, err := repo.TokenMaker.CreateToken(user.Email, "ADMIN", duration*24)
	if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	}

	createSessions := models.CreateSessionsParams{
		Email:        user.Email,
		ID:           accTokenPayload.ID,
		RefreshToken: refreshToken,
		UserAgent:    r.UserAgent(),
		ClientIp:     r.RemoteAddr,
		IsBlocked:    false,
		ExpiresAt:    refTokenPayload.ExpiredAt,
	}

	sessions, err := repo.DB.CreateSession(createSessions)
	if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, r, errResp)
		return
	}

	resp.AccessToken = accessToken
	resp.AccessTokenExpiresAt = accTokenPayload.ExpiredAt
	resp.RefreshToken = refreshToken
	resp.RefreshTokenExpires = refTokenPayload.ExpiredAt
	resp.User = newUserResponse(user)
	resp.SessionID = sessions.ID.String()
	resp.OK = true

	helpers.RenderJSON(w, r, resp)

}

func newUserResponse(user models.User) createUserResponse {
	return createUserResponse{
		FullName:  user.FirstName + " " + user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

// Logout logs the user out
func (repo *DBRepo) Logout(w http.ResponseWriter, r *http.Request) {

	// delete the remember me token, if any
	cookie, err := r.Cookie(fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]))
	if err != nil {
	} else {
		key := cookie.Value
		// have a remember token, so get the token
		if len(key) > 0 {
			// key length > 0, so it might be a valid token
			split := strings.Split(key, "|")
			hash := split[1]
			err = repo.DB.DeleteToken(hash)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// delete the remember me cookie, if any
	delCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_gowatcher_remember", app.PreferenceMap["identifier"]),
		Value:    "",
		Domain:   app.Domain,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
	}
	http.SetCookie(w, &delCookie)

	_ = app.Session.RenewToken(r.Context())
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	repo.App.Session.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

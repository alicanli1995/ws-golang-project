package handlers

import (
	"errors"
	"golang-observer-project/internal/helpers"
	"golang-observer-project/internal/models"
	"net/http"
	"time"
)

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	OK                   bool                    `json:"ok"`
	AccessToken          string                  `json:"access_token"`
	AccessTokenExpiresAt time.Time               `json:"access_token_expires_at"`
	RefreshToken         string                  `json:"refresh_token"`
	RefreshTokenExpires  time.Time               `json:"refresh_token_expires"`
	User                 createUserLoginResponse `json:"user"`
	SessionID            string                  `json:"session_id"`
}

type createUserLoginResponse struct {
	Username         string    `json:"username"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	PasswordChangeAt time.Time `json:"password_change_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type createUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func newUserLoginResponse(user models.User) createUserLoginResponse {
	return createUserLoginResponse{
		FullName:  user.FirstName + " " + user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
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
		helpers.RenderJSON(w, errResp)
	}

	id, _, err := repo.DB.Authenticate(req.Email, req.Password)
	if errors.Is(err, models.ErrInvalidCredentials) {
		errResp.Message = "Invalid username or password"
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	} else if errors.Is(err, models.ErrInactiveAccount) {
		errResp.Message = "Inactive account"
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	} else if err != nil {
		errResp.Message = "Something went wrong. Please try again later"
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	}

	user, err := repo.DB.GetUserById(id)
	if err != nil {
		errResp.Message = "DB Getting user by id error"
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	}
	duration := 45 * time.Minute
	accessToken, accTokenPayload, err := repo.TokenMaker.CreateToken(user.Email, "ADMIN", duration)
	if err != nil {
		errResp.Message = "While creating access token something went wrong."
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	}

	refreshToken, refTokenPayload, err := repo.TokenMaker.CreateToken(user.Email, "ADMIN", duration*24)
	if err != nil {
		errResp.Message = "While creating refresh token something went wrong."
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
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
		errResp.Message = "While creating session something went wrong."
		errResp.OK = false
		helpers.RenderJSON(w, errResp)
		return
	}

	resp.AccessToken = accessToken
	resp.AccessTokenExpiresAt = accTokenPayload.ExpiredAt
	resp.RefreshToken = refreshToken
	resp.RefreshTokenExpires = refTokenPayload.ExpiredAt
	resp.User = newUserLoginResponse(user)
	resp.SessionID = sessions.ID.String()
	resp.OK = true

	helpers.RenderJSON(w, resp)

}

// Register attempts to register a new user
func (repo *DBRepo) Register(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	var resp jsonResp

	err := helpers.ReadJSONBody(r, &req)
	if err != nil {
		resp.Message = "Invalid JSON"
		resp.OK = false
		helpers.RenderJSON(w, resp)
		return
	}

	user := models.User{
		FirstName:   req.Name,
		LastName:    req.Surname,
		Email:       req.Email,
		Password:    []byte(req.Password),
		AccessLevel: 3,
		UserActive:  1,
	}

	// check validation for first name and last name
	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Password == nil {
		resp.Message = "All fields are required"
		resp.OK = false
		helpers.RenderJSON(w, resp)
		return
	}

	// validate email is already in use
	user, err = repo.DB.GetUserByEmail(user.Email)
	if err != nil {
		resp.Message = "Something went wrong. Please try again later"
		resp.OK = false
		helpers.RenderJSON(w, resp)
		return
	}

	if user.ID > 0 {
		resp.Message = "Email address is already in use"
		resp.OK = false
		helpers.RenderJSON(w, resp)
		return
	}

	_, err = repo.DB.InsertUser(user)
	if err != nil {
		resp.Message = "While inserting user something went wrong."
		resp.OK = false
		helpers.RenderJSON(w, resp)
		return
	}

	resp.OK = true
	helpers.RenderJSON(w, resp)
}

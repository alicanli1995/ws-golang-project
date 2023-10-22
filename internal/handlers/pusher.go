package handlers

import (
	"github.com/pusher/pusher-http-go"
	"io"
	"log"
	"net/http"
)

func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {

	// get the user from the session
	u, _ := repo.DB.GetUserById(repo.App.Session.GetInt(r.Context(), "userID"))

	params, _ := io.ReadAll(r.Body)

	presenceData := pusher.MemberData{
		UserID: string(rune(u.ID)),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   string(rune(u.ID)),
		},
	}

	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response)
}

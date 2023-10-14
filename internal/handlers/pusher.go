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

func (repo *DBRepo) PusherTest(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "Hello from the server!",
	}

	err := repo.App.WsClient.Trigger("public-channel", "test-event", data)
	if err != nil {
		log.Println(err)
		return
	}

	_, _ = w.Write([]byte("Triggered event"))
}

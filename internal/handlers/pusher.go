package handlers

import (
	"net/http"
)

// SendPrivateMessage sends a private message to a user
func (repo *DBRepo) SendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	userID := r.URL.Query().Get("id")

	data := map[string]string{
		"message": msg,
	}

	_ = app.WsClient.Trigger("private-channel-"+userID, "private-message", data)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"message": "ok"}`))

}

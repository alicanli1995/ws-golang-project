package sms

import (
	"encoding/json"
	"errors"
	"golang-vigilate-project/internal/config"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func SendTextTwilio(to, msg string, app *config.AppConfig) error {
	secret := app.PreferenceMap["twilio_auth_token"]
	accountSid := app.PreferenceMap["twilio_sid"]

	urlString := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", app.PreferenceMap["twilio_phone_number"])
	msgData.Set("Body", msg)
	msgDataReader := *strings.NewReader(msgData.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlString, &msgDataReader)
	req.SetBasicAuth(accountSid, secret)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, _ := client.Do(req)
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(res.Body)
		err := decoder.Decode(&data)
		if err == nil {
			log.Printf("Twilio response: %s", data["sid"])
			return err
		}
	} else {
		log.Println("Twilio error: ", res.Status)
		return errors.New("failed to send sms")
	}

	return nil

}

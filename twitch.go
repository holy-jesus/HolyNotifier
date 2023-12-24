package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	// "sync"
	// "github.com/deta/deta-go/deta"
	// "github.com/deta/deta-go/service/base"
	// "github.com/go-telegram/bot"
	// "github.com/go-telegram/bot/models"
)

var VERSIONS = map[string]interface{}{"channel.update": "2", "stream.online": "1", "stream.offline": "1"}

func verifyHMAC(r *http.Request, body []byte) bool {
	var bytes []byte
	bytes = append(bytes, r.Header.Get("Twitch-Eventsub-Message-Id")...)
	bytes = append(bytes, r.Header.Get("Twitch-Eventsub-Message-Timestamp")...)
	bytes = append(bytes, body...)
	mac := hmac.New(sha256.New, []byte(KEY))
	mac.Write(bytes)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.Replace(r.Header.Get("Twitch-Eventsub-Message-Signature"), "sha256=", "", 1)), []byte(expectedMAC))
}

func verifyTime(r *http.Request) bool {
	now := time.Now()
	formattedDate, err := time.Parse(time.RFC3339, r.Header.Get("Twitch-Eventsub-Message-Timestamp"))
	if err != nil {
		return false
	}
	dif := now.Sub(formattedDate)
	return dif.Minutes() <= 10
}

func verifyEvent(r *http.Request, body []byte) bool {
	wrong_request := 0
	if !verifyHMAC(r, body) {
		wrong_request += 1
	}
	if !verifyTime(r) {
		wrong_request += 1
	}
	if  VERSIONS[r.Header.Get("Twitch-Eventsub-Subscription-Type")] != r.Header.Get("Twitch-Eventsub-Subscription-Version") {
		wrong_request += 1
	}
	if wrong_request == 0 {
		return true
	} else {
		return false
	}
}

func twitchWebHook(w http.ResponseWriter, r *http.Request) {
	// var wg sync.WaitGroup
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if mediaType != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	body, _ := io.ReadAll(r.Body)
	messageType := r.Header.Get("Twitch-Eventsub-Message-Type")
	if messageType == "notification" {
		return
	} else if messageType == "webhook_callback_verification" {
		m := map[string]interface{}{"challenge": "", "subscription": map[string]interface{}{"id": "", "status": "", "type": "", "version": "", "cost": 0, "condition": map[string]interface{}{"broadcaster_user_id": ""}, "transport": map[string]interface{}{"method": "", "callback": ""}, "created_at": ""}}
		errJson := json.Unmarshal(body, &m)
		if errJson != nil || !verifyEvent(r, body) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, m["challenge"])
		return
	} else if messageType == "revocation" {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

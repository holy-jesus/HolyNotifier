package main

import (
	"fmt"
	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	// "github.com/go-telegram/bot"
	// "github.com/go-telegram/bot/models"
)

var Deta, err_deta = deta.New(deta.WithProjectKey(os.Getenv("DETA_PROJECT_KEY")))
var DB, err_base = base.New(Deta, "config")

var KEY = ""
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!+-#$%&,.")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	var wg sync.WaitGroup
	if err_deta != nil {
		panic(err_deta)
	}
	if err_base != nil {
		panic(err_base)
	}
	wg.Add(1)
	go func() {
		item := map[string]interface{}{"key": "secret", "value": ""}
		err := DB.Get("secret", &item)
		if err != nil {
			KEY = randSeq(99)
			item["value"] = KEY
			DB.Put(item)
		} else {
			if str, ok := item["value"].(string); ok {
				KEY = str
			} else {
				fmt.Printf("secret should be string, not %T", item["value"])
			}
		}
		wg.Done()
	}()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/twitchwebhook", twitchWebHook)

	log.Printf("App listening on port %s!", port)
	wg.Wait()
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Здесь скоро будет ТАКОЕ...")
}

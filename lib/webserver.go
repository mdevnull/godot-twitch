package lib

import (
	"fmt"
	"net/http"

	"github.com/nicklaw5/helix/v2"
)

func WebServer(twitchClient *helix.Client) <-chan bool {
	clientAuthChan := make(chan bool)

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got a request in")

		resp, err := twitchClient.RequestUserAccessToken(r.URL.Query().Get("code"))
		if err != nil {
			fmt.Printf("error: unable to get access token: %s\n", err.Error())
			return
		}

		twitchClient.SetUserAccessToken(resp.Data.AccessToken)
		clientAuthChan <- true
	})

	server := &http.Server{Addr: ":8189", Handler: mux}
	go server.ListenAndServe()
	fmt.Println("webserver should be up")

	return clientAuthChan
}

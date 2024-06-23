package lib

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	twitchSessionPayload struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		ReconnectURL string `json:"reconnect_url"`
	}
	twitchSubscriptionPayload struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Type   string `json:"type"`
	}
	twitchPayload struct {
		Session      *twitchSessionPayload      `json:"session"`
		Subscription *twitchSubscriptionPayload `json:"subscription"`
		Event        map[string]interface{}     `json:"event"`
	}
	twitchMetaData struct {
		ID   string `json:"message_id"`
		Type string `json:"message_type"`
	}
	TwitchMessage struct {
		Metadata twitchMetaData `json:"metadata"`
		Payload  twitchPayload  `json:"payload"`
	}
)

func Websocket(useDebug bool) (<-chan TwitchMessage, <-chan string) {
	twitchEventChan := make(chan TwitchMessage, 1)
	sessionIDChan := make(chan string, 1)
	go func() {
		wsConenctURL := "wss://eventsub.wss.twitch.tv/ws?keepalive_timeout_seconds=30"
		if useDebug {
			wsConenctURL = "ws://localhost:8190/ws"
		}

		for {
			newReconnectURL, err := makeConnAndRead(wsConenctURL, twitchEventChan, sessionIDChan)
			if err != nil {
				panic(err)
			}
			if newReconnectURL != "" {
				wsConenctURL = newReconnectURL
			}
		}
	}()

	return twitchEventChan, sessionIDChan
}

func makeConnAndRead(
	url string,
	msgOutChan chan<- TwitchMessage, sessionIDChan chan<- string,
) (string, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		fmt.Printf("error: unable to establish websocket connection\n")
	}

	defer conn.Close()
	for {
		msg := TwitchMessage{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			return "", fmt.Errorf("error: parsing event JSON: %w", err)
		}

		switch msg.Metadata.Type {
		case "session_welcome":
			sessionIDChan <- msg.Payload.Session.ID
			if msg.Payload.Session.ReconnectURL != "" {
				return msg.Payload.Session.ReconnectURL, nil
			}
		case "session_reconnect":
			return msg.Payload.Session.ReconnectURL, nil
		case "session_keepalive":
			fmt.Println("ws: keep alive")
		default:
			msgOutChan <- msg
		}
	}
}

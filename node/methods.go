package node

import (
	"fmt"
	"main/lib"
	"os/exec"
	"strings"
	"sync"

	"github.com/nicklaw5/helix/v2"
	"graphics.gd/classdb/FileAccess"
	"graphics.gd/classdb/OS"
	"graphics.gd/variant/Float"
)

func (h *GodotTwitch) Ready() {
	if h.ClientID == "" || h.ClientSecret == "" {
		lib.LogErr("missing client id or client secret")
		return
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:     h.ClientID,
		ClientSecret: h.ClientSecret,
		RedirectURI:  "http://localhost:8189/",
	})
	if err != nil {
		lib.LogErr(fmt.Sprintf("unable to create client: %s\n", err.Error()))
		return
	}

	authURLString := client.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		Scopes: []string{
			"bits:read",
			"channel:read:charity", "channel:read:redemptions", "channel:read:ads", "channel:read:subscriptions",
			"channel:read:polls", "channel:read:predictions", "channel:read:goals",
			"moderator:read:followers", "moderator:read:shoutouts",
		},
	})
	h.AuthURL = authURLString
	lib.LogInfo(authURLString)

	h.LatestFollower = ""
	h.LatestSubscriber = ""

	h.apiInfoResponseLock = sync.Mutex{}
	h.apiInfoResponseQueue = make([]interface{}, 0)
	h.eventProcessLock = sync.Mutex{}
	h.eventProcessQueue = make([]lib.TwitchMessage, 0)
	h.twitchClient = client

	h.IsAuthenticated = false
	// check if we have a access and refresh token to load
	if bool(h.StoreToken) {
		h.IsAuthenticated = h.readTokens()
	}

	go func() {
		// either read from file failed or its the first start so run through normal auth
		if !h.IsAuthenticated {
			clientAuthedMsgChan := lib.WebServer(client)
			// wait for auth callback
			<-clientAuthedMsgChan

			h.hasNewToken = true
		}

		broadcasterUserResp, err := client.GetUsers(&helix.UsersParams{})
		if err != nil {
			fmt.Printf("error: unable to get current user: %s\n", err.Error())
			return
		}
		broadcasterUserID := broadcasterUserResp.Data.Users[0].ID

		followerResp, err := client.GetChannelFollows(&helix.GetChannelFollowsParams{
			BroadcasterID: broadcasterUserID,
			First:         1,
		})
		if err != nil {
			fmt.Printf("error: unable to get latest channel follower: %s\n", err.Error())
		} else {
			if len(followerResp.Data.Channels) > 0 {
				follower := followerResp.Data.Channels[0]

				h.apiInfoResponseLock.Lock()
				h.apiInfoResponseQueue = append(h.apiInfoResponseQueue, LatestFollowerUpdate{follower.Username})
				h.apiInfoResponseLock.Unlock()
			}
		}

		subscribersResp, err := client.GetSubscriptions(&helix.SubscriptionsParams{
			BroadcasterID: broadcasterUserID,
			First:         1,
		})
		if err != nil {
			fmt.Printf("error: unable to get latest subscriber: %s\n", err.Error())
		} else {
			if len(subscribersResp.Data.Subscriptions) > 0 {
				subscriber := subscribersResp.Data.Subscriptions[0]

				h.apiInfoResponseLock.Lock()
				h.apiInfoResponseQueue = append(h.apiInfoResponseQueue, LatestSubscriberUpdate{subscriber.UserName})
				h.apiInfoResponseLock.Unlock()
			}
		}

		msgChan, sessChan := lib.Websocket(bool(h.UseDebugWS))
		for {
			select {
			case wsSessionID := <-sessChan:
				if bool(h.UseDebugWS) {
					fmt.Println("Debug WS so we ignore session. But consider it ACK :patDev:")
					continue
				}

				// every time we (re)connect we have to subscribwe events again
				lib.EventSetup(client, wsSessionID, broadcasterUserID)

			case msg := <-msgChan:
				h.eventProcessLock.Lock()
				h.eventProcessQueue = append(h.eventProcessQueue, msg)
				h.eventProcessLock.Unlock()
			}
		}
	}()
}

func (h *GodotTwitch) Process(delta Float.X) {
	if h.hasNewToken {
		h.hasNewToken = false
		h.IsAuthenticated = true

		if bool(h.StoreToken) {
			var accessWriteFa FileAccess.Instance = FileAccess.Open("user://twitch_access_token.txt", FileAccess.Write)
			accessWriteFa.StoreString(h.twitchClient.GetUserAccessToken())

			var refreshWriteFa FileAccess.Instance = FileAccess.Open("user://twitch_refresh_token.txt", FileAccess.Write)
			refreshWriteFa.StoreString(h.twitchClient.GetRefreshToken())
		}
	}

	h.handleApiUpdateTick()
	h.handleEventTick()
}

func (h *GodotTwitch) OpenAuthInBrowser() {
	var openCmd *exec.Cmd
	switch strings.ToLower(OS.GetName()) {
	case "windows":
		winQuotedURL := strings.ReplaceAll(h.AuthURL, "&", "^&")
		openCmd = exec.Command("cmd", "/c", "start", winQuotedURL)
	case "macos":
		openCmd = exec.Command("open", h.AuthURL)
	case "linux":
		openCmd = exec.Command("xdg-open", h.AuthURL)
	default:
		lib.LogWarn("unable to open browser on current platform")
		return
	}

	if err := openCmd.Run(); err != nil {
		lib.LogErr(fmt.Sprintf(" error opening browser for auth: %s", err.Error()))
	}
}

func (h *GodotTwitch) readTokens() bool {
	if !FileAccess.FileExists("user://twitch_access_token.txt") {
		return false
	}
	if !FileAccess.FileExists("user://twitch_refresh_token.txt") {
		return false
	}

	accessReadFa := FileAccess.Open("user://twitch_access_token.txt", FileAccess.Read)
	gdAccess := FileAccess.Instance(accessReadFa).GetAsText()

	refreshReadFa := FileAccess.Open("user://twitch_refresh_token.txt", FileAccess.Read)
	gdRefresh := FileAccess.Instance(refreshReadFa).GetAsText()

	if gdAccess == "" || gdRefresh == "" {
		return false
	}

	h.twitchClient.SetUserAccessToken(gdAccess)
	h.twitchClient.SetRefreshToken(gdRefresh)
	return true
}

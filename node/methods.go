package node

import (
	"fmt"
	"main/lib"
	"os/exec"
	"strings"
	"sync"

	"github.com/nicklaw5/helix/v2"
	"grow.graphics/gd"
)

func (h *GodotTwitch) Ready(godoCtx gd.Context) {
	if h.ClientID.String() == "" || h.ClientSecret.String() == "" {
		lib.LogErr(godoCtx, "missing client id or client secret")
		return
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:     h.ClientID.String(),
		ClientSecret: h.ClientSecret.String(),
		RedirectURI:  "http://localhost:8189/",
	})
	if err != nil {
		lib.LogErr(godoCtx, fmt.Sprintf("unable to create client: %s\n", err.Error()))
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
	h.AuthURL = h.Pin().String(authURLString)
	lib.LogInfo(godoCtx, authURLString)

	h.LatestFollower = h.Pin().String("")
	h.LatestSubscriber = h.Pin().String("")

	h.apiInfoResponseLock = sync.Mutex{}
	h.apiInfoResponseQueue = make([]interface{}, 0)
	h.eventProcessLock = sync.Mutex{}
	h.eventProcessQueue = make([]lib.TwitchMessage, 0)
	h.twitchClient = client

	h.IsAuthenticated = gd.Bool(false)
	// check if we have a access and refresh token to load
	if bool(h.StoreToken) {
		h.IsAuthenticated = gd.Bool(h.readTokens(godoCtx))
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
				lib.EventSetup(godoCtx, client, wsSessionID, broadcasterUserID)

			case msg := <-msgChan:
				h.eventProcessLock.Lock()
				h.eventProcessQueue = append(h.eventProcessQueue, msg)
				h.eventProcessLock.Unlock()
			}
		}
	}()
}

func (h *GodotTwitch) Process(godoCtx gd.Context, delta gd.Float) {
	if h.hasNewToken {
		h.hasNewToken = false
		h.IsAuthenticated = true

		if bool(h.StoreToken) {
			var fa gd.FileAccess
			accessWriteFa := fa.Open(godoCtx, godoCtx.String("user://twitch_access_token.txt"), gd.FileAccessModeFlags(2))
			accessWriteFa.StoreString(godoCtx.String(h.twitchClient.GetUserAccessToken()))

			refreshWriteFa := fa.Open(godoCtx, godoCtx.String("user://twitch_refresh_token.txt"), gd.FileAccessModeFlags(2))
			refreshWriteFa.StoreString(godoCtx.String(h.twitchClient.GetRefreshToken()))
		}
	}

	h.handleApiUpdateTick(godoCtx)
	h.handleEventTick(godoCtx)
}

func (h *GodotTwitch) OpenAuthInBrowser(godoCtx gd.Context) {
	godotOS := gd.OS(godoCtx)

	var openCmd *exec.Cmd
	switch strings.ToLower(godotOS.GetName(godoCtx).String()) {
	case "windows":
		winQuotedURL := strings.ReplaceAll(h.AuthURL.String(), "&", "^&")
		openCmd = exec.Command("cmd", "/c", "start", winQuotedURL)
	case "macos":
		openCmd = exec.Command("open", h.AuthURL.String())
	case "linux":
		openCmd = exec.Command("xdg-open", h.AuthURL.String())
	default:
		lib.LogWarn(godoCtx, "unable to open browser on current platform")
		return
	}

	if err := openCmd.Run(); err != nil {
		lib.LogErr(godoCtx, fmt.Sprintf(" error opening browser for auth: %s", err.Error()))
	}
}

func (h *GodotTwitch) readTokens(godoCtx gd.Context) bool {
	var fa gd.FileAccess
	if !fa.FileExists(godoCtx, godoCtx.String("user://twitch_access_token.txt")) {
		return false
	}
	if !fa.FileExists(godoCtx, godoCtx.String("user://twitch_refresh_token.txt")) {
		return false
	}

	accessReadFa := fa.Open(godoCtx, godoCtx.String("user://twitch_access_token.txt"), gd.FileAccessModeFlags(1))
	gdAccess := accessReadFa.GetAsText(godoCtx, true)

	refreshReadFa := fa.Open(godoCtx, godoCtx.String("user://twitch_refresh_token.txt"), gd.FileAccessModeFlags(1))
	gdRefresh := refreshReadFa.GetAsText(godoCtx, true)

	if gdAccess.String() == "" || gdRefresh.String() == "" {
		return false
	}

	h.twitchClient.SetUserAccessToken(gdAccess.String())
	h.twitchClient.SetRefreshToken(gdRefresh.String())
	return true
}

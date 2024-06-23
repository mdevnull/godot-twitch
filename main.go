package main

import (
	"fmt"
	"main/lib"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
	"grow.graphics/gd"
	"grow.graphics/gd/gdextension"
)

type GodotTwitch struct {
	gd.Class[GodotTwitch, gd.Node]

	ClientID     gd.String `gd:"twitch_client_id"`
	ClientSecret gd.String `gd:"twitch_client_secret"`

	AuthURL    gd.String `gd:"auth_url"`
	UseDebugWS gd.Bool   `gd:"use_debug_ws_server"`

	OnFollow       gd.SignalAs[func(gd.String)] `gd:"on_follow"`
	LatestFollower gd.String                    `gd:"latest_follower"`

	OnSubscibtion    gd.SignalAs[func(gd.String, gd.Int, gd.Int)]         `gd:"on_subscribtion"`
	OnGiftSubs       gd.SignalAs[func(gd.String, gd.Int, gd.Int, gd.Int)] `gd:"on_sub_gift"`
	LatestSubscriber gd.String                                            `gd:"latest_subscriber"`

	OnIncomingRaid       gd.SignalAs[func(gd.String, gd.String, gd.Int)]                                  `gd:"on_raid"`
	OnRewardRedemtionAdd gd.SignalAs[func(gd.String, gd.String, gd.String, gd.String, gd.String, gd.Int)] `gd:"on_redeem"`
	OnShoutoutCreate     gd.SignalAs[func(gd.String, gd.String, gd.String, gd.String)]                    `gd:"on_shoutout_create"`
	OnDonation           gd.SignalAs[func(gd.String, gd.Float, gd.String)]                                `gd:"on_donation"`
	OnPollBegin          gd.SignalAs[func(gd.String, gd.Int)]                                             `gd:"on_poll_begin"`
	OnPollEnd            gd.SignalAs[func(gd.String)]                                                     `gd:"on_poll_end"`
	OnPredictionBegin    gd.SignalAs[func(gd.String, gd.Int)]                                             `gd:"on_prediction_begin"`
	OnPredictionLock     gd.SignalAs[func(gd.String)]                                                     `gd:"on_prediction_lock"`
	OnPredictionEnd      gd.SignalAs[func(gd.String)]                                                     `gd:"on_prediction_end"`

	twitchClient *helix.Client

	eventProcessLock  sync.Mutex
	eventProcessQueue []lib.TwitchMessage
}

func (h *GodotTwitch) Ready(godoCtx gd.Context) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     h.ClientID.String(),
		ClientSecret: h.ClientSecret.String(),
		RedirectURI:  "http://localhost:8189/",
	})
	if err != nil {
		fmt.Printf("error: unable to create client: %s\n", err.Error())
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

	h.eventProcessLock = sync.Mutex{}
	h.twitchClient = client
	go func() {
		clientAuthedMsgChan := lib.WebServer(client)
		// wait for auth callback
		<-clientAuthedMsgChan

		broadcasterUserResp, err := client.GetUsers(&helix.UsersParams{})
		if err != nil {
			fmt.Printf("error: unable to get current user: %s\n", err.Error())
			return
		}
		broadcasterUserID := broadcasterUserResp.Data.Users[0].ID

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
	h.eventProcessLock.Lock()
	defer h.eventProcessLock.Unlock()

	if len(h.eventProcessQueue) <= 0 {
		return
	}

	for _, msg := range h.eventProcessQueue {
		h.handleEvent(godoCtx, msg)
	}
	h.eventProcessQueue = make([]lib.TwitchMessage, 0)
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

func (h *GodotTwitch) handleEvent(godoCtx gd.Context, eventMsg lib.TwitchMessage) {
	if eventMsg.Payload.Subscription == nil {
		fmt.Printf("%+v\n", eventMsg)
		lib.LogWarn(godoCtx, fmt.Sprintf("received non subscribtion event: %s", eventMsg.Metadata.Type))
		return
	}

	lib.LogInfo(godoCtx, fmt.Sprintf("received event for: %s", eventMsg.Payload.Subscription.Type))

	switch eventMsg.Payload.Subscription.Type {
	case helix.EventSubTypeChannelFollow:
		h.LatestFollower = h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
		h.OnFollow.Emit(h.LatestFollower)
	case helix.EventSubTypeChannelSubscription:
		h.LatestSubscriber = h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
		// only emit signal for non gift subs as we emit from the gift event for gifts and we
		// we do not want double events
		if !eventMsg.Payload.Event["is_gift"].(bool) {
			tier := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "tier")
			h.OnSubscibtion.Emit(h.LatestSubscriber, gd.Int(1), tier)
		}
	case helix.EventSubTypeChannelSubscriptionMessage:
		h.LatestSubscriber = h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
		tier := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "tier")
		if tier >= 1000 {
			tier = tier % 1000
		}
		months := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "cumulative_months")
		h.OnSubscibtion.Emit(h.LatestSubscriber, months, tier)
	case helix.EventSubTypeChannelSubscriptionGift:
		isAnonymous := h.readBoolFromEvent(godoCtx, eventMsg.Payload.Event, "is_anonymous")
		var gifterName gd.String
		var totalAmountForUser gd.Int
		if !isAnonymous {
			gifterName = h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
			totalAmountForUser = h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "cumulative_total")
		} else {
			gifterName = h.Pin().String("")
		}
		tier := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "tier")
		giftAmount := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "total")

		h.OnGiftSubs.Emit(gifterName, giftAmount, tier, totalAmountForUser)
	case helix.EventSubTypeChannelRaid:
		fromUserID := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "from_broadcaster_user_id")
		fromUser := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "from_broadcaster_user_name")
		viewerCount := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "viewers")

		userResp, err := h.twitchClient.GetUsers(&helix.UsersParams{
			IDs: []string{fromUserID.String()},
		})
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch user %s: %s", fromUserID.String(), err.Error()))
			return
		}
		if len(userResp.Data.Users) <= 0 {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch user %s: empty result", fromUserID.String()))
			return
		}

		userObj := userResp.Data.Users[0]
		profilePicUrl := h.Pin().String(userObj.ProfileImageURL)

		h.OnIncomingRaid.Emit(fromUser, profilePicUrl, viewerCount)
	case helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd:
		rewardInterface, ok := eventMsg.Payload.Event["reward"]
		if !ok {
			lib.LogErr(godoCtx, "missing event data: reward")
			return
		}

		reward := rewardInterface.(map[string]interface{})
		fromUser := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
		userInput := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_input")
		rewardID := h.readStringFromEvent(godoCtx, reward, "id")
		rewardTitle := h.readStringFromEvent(godoCtx, reward, "title")
		rewardPrompt := h.readStringFromEvent(godoCtx, reward, "prompt")
		rewardCost := h.readIntFromEvent(godoCtx, reward, "cost")

		h.OnRewardRedemtionAdd.Emit(
			fromUser, userInput,
			rewardID, rewardTitle, rewardPrompt, rewardCost,
		)
	case helix.EventSubShoutoutCreate:
		broadcasterID := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "to_broadcaster_user_id")
		broadcasterName := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "to_broadcaster_user_name")

		userResp, err := h.twitchClient.GetUsers(&helix.UsersParams{
			IDs: []string{broadcasterID.String()},
		})
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch user %s: %s", broadcasterID.String(), err.Error()))
			return
		}
		if len(userResp.Data.Users) <= 0 {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch user %s: empty result", broadcasterID.String()))
			return
		}

		userObj := userResp.Data.Users[0]
		profilePicUrl := h.Pin().String(userObj.ProfileImageURL)

		channelInfo, err := h.twitchClient.GetChannelInformation(&helix.GetChannelInformationParams{
			BroadcasterIDs: []string{broadcasterID.String()},
		})
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch channel info for %s: %s", broadcasterID.String(), err.Error()))
			return
		}
		if len(channelInfo.Data.Channels) <= 0 {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to fetch channel %s: empty result", broadcasterID.String()))
			return
		}

		channelObj := channelInfo.Data.Channels[0]
		lastGameName := h.Pin().String(channelObj.GameName)
		lastStreamTitle := h.Pin().String(channelObj.Title)

		h.OnShoutoutCreate.Emit(broadcasterName, profilePicUrl, lastGameName, lastStreamTitle)
	case helix.EventSubTypeCharityDonation:
		amountInterface, ok := eventMsg.Payload.Event["amount"]
		if !ok {
			lib.LogErr(godoCtx, "missing event data: amount")
			return
		}

		amount := amountInterface.(map[string]interface{})

		donatorName := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")
		value := h.readIntFromEvent(godoCtx, amount, "value")
		decimalPlaces := h.readIntFromEvent(godoCtx, amount, "decimal_places")
		currency := h.readStringFromEvent(godoCtx, amount, "currency")

		valueFloat := float64(value)
		if decimalPlaces > 0 {
			valueFloat = float64(value) / math.Pow(10, float64(decimalPlaces))
		}

		h.OnDonation.Emit(donatorName, gd.Float(valueFloat), currency)
	case helix.EventSubTypeChannelPollBegin:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")
		endsAtStr := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "ends_at")
		endsAt, err := time.Parse(time.RFC3339, endsAtStr.String())
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("error converting timestamp: %s", err.Error()))
			return
		}

		//TODO: arrays and stuff is kinda undocumented in grow.graphics/gd rn so I have to trial and error a bunch
		// 			later to transmit poll options

		h.OnPollBegin.Emit(title, gd.Int(endsAt.Unix()))
	case helix.EventSubTypeChannelPollProgress:
		//TODO: until the above todo isnt fixed this doesnt make a lot of sense
	case helix.EventSubTypeChannelPollEnd:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		//TODO: also see todo in EventSubTypeChannelPollBegin handling

		h.OnPollEnd.Emit(title)
	case helix.EventSubTypeChannelPredictionBegin:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")
		locksAtStr := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "locks_at")
		locksAt, err := time.Parse(time.RFC3339, locksAtStr.String())
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("error converting timestamp: %s", err.Error()))
			return
		}

		//TODO: also see todo in EventSubTypeChannelPollBegin handling

		h.OnPredictionBegin.Emit(title, gd.Int(locksAt.Unix()))
	case helix.EventSubTypeChannelPredictionProgress:
		//TODO: also see todo in EventSubTypeChannelPollBegin handling
	case helix.EventSubTypeChannelPredictionLock:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		//TODO: also see todo in EventSubTypeChannelPollBegin handling

		h.OnPredictionLock.Emit(title)
	case helix.EventSubTypeChannelPredictionEnd:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		//TODO: also see todo in EventSubTypeChannelPollBegin handling

		h.OnPredictionEnd.Emit(title)
	}
}

func (h *GodotTwitch) readStringFromEvent(godoCtx gd.Context, eventPayload map[string]interface{}, key string) gd.String {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(godoCtx, fmt.Sprintf("missing event data: %s", key))
		return h.Pin().String("")
	}

	switch v := val.(type) {
	case string:
		return h.Pin().String(v)
	case int:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case int32:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case int64:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case uint:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case uint32:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case uint64:
		return h.Pin().String(fmt.Sprintf("%d", v))
	case float32:
		return h.Pin().String(fmt.Sprintf("%.2f", v))
	case float64:
		return h.Pin().String(fmt.Sprintf("%.2f", v))
	case nil:
		return h.Pin().String("")
	default:
		lib.LogWarn(godoCtx, fmt.Sprintf("unknown event data type %T", val))
		return h.Pin().String("")
	}
}

func (h *GodotTwitch) readIntFromEvent(godoCtx gd.Context, eventPayload map[string]interface{}, key string) gd.Int {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(godoCtx, fmt.Sprintf("missing event data: %s", key))
		return gd.Int(0)
	}

	switch v := val.(type) {
	case string:
		asint, err := strconv.Atoi(v)
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("unable to read %s as int: %s", key, err.Error()))
			return gd.Int(0)
		}
		return gd.Int(asint)
	case int:
		return gd.Int(v)
	case int32:
		return gd.Int(v)
	case int64:
		return gd.Int(v)
	case uint:
		return gd.Int(v)
	case uint32:
		return gd.Int(v)
	case uint64:
		return gd.Int(v)
	case float32:
		return gd.Int(v)
	case float64:
		return gd.Int(v)
	case nil:
		return gd.Int(0)
	default:
		lib.LogWarn(godoCtx, fmt.Sprintf("unknown event data type %T", val))
		return gd.Int(0)
	}
}

func (h *GodotTwitch) readBoolFromEvent(godoCtx gd.Context, eventPayload map[string]interface{}, key string) gd.Bool {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(godoCtx, fmt.Sprintf("missing event data: %s", key))
		return gd.Bool(false)
	}

	valAsBoo, isBool := val.(bool)
	if !isBool {
		lib.LogWarn(godoCtx, fmt.Sprintf("cannot read %T as bool", val))
		return gd.Bool(false)
	}

	return gd.Bool(valAsBoo)
}

func main() {
	godot, ok := gdextension.Link()
	if !ok {
		panic("Unable to link to godot")
	}
	gd.Register[GodotTwitch](godot)
}

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
	OnPollBegin          gd.SignalAs[func(gd.String, gd.Int, gd.ArrayOf[gd.Dictionary])]                  `gd:"on_poll_begin"`
	OnPollProgress       gd.SignalAs[func(gd.String, gd.ArrayOf[gd.Dictionary])]                          `gd:"on_poll_progress"`
	OnPollEnd            gd.SignalAs[func(gd.String, gd.ArrayOf[gd.Dictionary])]                          `gd:"on_poll_end"`
	OnPredictionBegin    gd.SignalAs[func(gd.String, gd.Int, gd.ArrayOf[gd.Dictionary])]                  `gd:"on_prediction_begin"`
	OnPredictionProgress gd.SignalAs[func(gd.String, gd.ArrayOf[gd.Dictionary])]                          `gd:"on_prediction_progress"`
	OnPredictionLock     gd.SignalAs[func(gd.String, gd.ArrayOf[gd.Dictionary])]                          `gd:"on_prediction_lock"`
	OnPredictionEnd      gd.SignalAs[func(gd.String, gd.ArrayOf[gd.Dictionary])]                          `gd:"on_prediction_end"`

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

		choicesArray := h.readPollChoices(godoCtx, eventMsg, true)
		if choicesArray == nil {
			return
		}

		h.OnPollBegin.Emit(title, gd.Int(endsAt.Unix()), choicesArray)
	case helix.EventSubTypeChannelPollProgress:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		choicesArray := h.readPollChoices(godoCtx, eventMsg, false)
		if choicesArray == nil {
			return
		}

		h.OnPollProgress.Emit(title, choicesArray)
	case helix.EventSubTypeChannelPollEnd:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		choicesArray := h.readPollChoices(godoCtx, eventMsg, false)
		if choicesArray == nil {
			return
		}

		h.OnPollEnd.Emit(title, choicesArray)
	case helix.EventSubTypeChannelPredictionBegin:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")
		locksAtStr := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "locks_at")
		locksAt, err := time.Parse(time.RFC3339, locksAtStr.String())
		if err != nil {
			lib.LogErr(godoCtx, fmt.Sprintf("error converting timestamp: %s", err.Error()))
			return
		}

		outcomesArray := h.readPredictionOutComes(godoCtx, eventMsg, true, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionBegin.Emit(title, gd.Int(locksAt.Unix()), outcomesArray)
	case helix.EventSubTypeChannelPredictionProgress:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(godoCtx, eventMsg, false, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionProgress.Emit(title, outcomesArray)
	case helix.EventSubTypeChannelPredictionLock:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(godoCtx, eventMsg, false, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionLock.Emit(title, outcomesArray)
	case helix.EventSubTypeChannelPredictionEnd:
		title := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(godoCtx, eventMsg, false, true)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionEnd.Emit(title, outcomesArray)
	}
}

func (h *GodotTwitch) readStringFromEvent(godoCtx gd.Context, eventPayload map[string]interface{}, key string) gd.String {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(godoCtx, fmt.Sprintf("missing event data: %s", key))
		return godoCtx.String("")
	}

	switch v := val.(type) {
	case string:
		return godoCtx.String(v)
	case int:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case int32:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case int64:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case uint:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case uint32:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case uint64:
		return godoCtx.String(fmt.Sprintf("%d", v))
	case float32:
		return godoCtx.String(fmt.Sprintf("%.2f", v))
	case float64:
		return godoCtx.String(fmt.Sprintf("%.2f", v))
	case nil:
		return godoCtx.String("")
	default:
		lib.LogWarn(godoCtx, fmt.Sprintf("unknown event data type %T", val))
		return godoCtx.String("")
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

func (h *GodotTwitch) readPollChoices(
	godoCtx gd.Context, eventMsg lib.TwitchMessage,
	onlyBeginning bool,
) gd.ArrayOf[gd.Dictionary] {
	choicesArray := gd.NewArrayOf[gd.Dictionary](godoCtx)
	choicesInterfaces, ok := eventMsg.Payload.Event["choices"]
	if !ok {
		lib.LogErr(godoCtx, "poll without any choices")
		return nil
	}
	choices := choicesInterfaces.([]interface{})
	if len(choices) > 0 {
		for _, choiceInterface := range choices {
			choice, ok := choiceInterface.(map[string]interface{})
			if !ok {
				lib.LogErr(godoCtx, fmt.Sprintf("error converting single choice: expected map[string]interface{} but got %T", choiceInterface))
				return nil
			}

			choiceID := h.readStringFromEvent(godoCtx, choice, "id")
			choiceName := h.readStringFromEvent(godoCtx, choice, "title")

			choiceDict := godoCtx.Dictionary()
			choiceDict.SetIndex(godoCtx.Variant(godoCtx.String("id")), godoCtx.Variant(choiceID))
			choiceDict.SetIndex(godoCtx.Variant(godoCtx.String("title")), godoCtx.Variant(choiceName))

			if !onlyBeginning {
				bitsVoted := h.readIntFromEvent(godoCtx, choice, "bits_votes")
				pointsVoted := h.readIntFromEvent(godoCtx, choice, "channel_points_votes")
				totalVoted := h.readIntFromEvent(godoCtx, choice, "votes")

				choiceDict.SetIndex(godoCtx.Variant(godoCtx.String("bits_votes")), godoCtx.Variant(bitsVoted))
				choiceDict.SetIndex(godoCtx.Variant(godoCtx.String("channel_points_votes")), godoCtx.Variant(pointsVoted))
				choiceDict.SetIndex(godoCtx.Variant(godoCtx.String("votes")), godoCtx.Variant(totalVoted))
			}

			choicesArray.Append(choiceDict)
		}
	}

	return choicesArray
}

func (h *GodotTwitch) readPredictionOutComes(
	godoCtx gd.Context, eventMsg lib.TwitchMessage,
	onlyBeginning bool, withWinnings bool,
) gd.ArrayOf[gd.Dictionary] {
	outcomesArray := gd.NewArrayOf[gd.Dictionary](godoCtx)
	outcomesInterfaces, ok := eventMsg.Payload.Event["outcomes"]
	if !ok {
		lib.LogErr(godoCtx, "prediction without any outcomes")
		return nil
	}
	outcomes := outcomesInterfaces.([]interface{})
	if len(outcomes) > 0 {
		for _, outcomeInterface := range outcomes {
			outcome, ok := outcomeInterface.(map[string]interface{})
			if !ok {
				lib.LogErr(godoCtx, fmt.Sprintf("error converting single outcome: expected map[string]interface{} but got %T", outcomeInterface))
				return nil
			}

			outcomeID := h.readStringFromEvent(godoCtx, outcome, "id")
			outcomeName := h.readStringFromEvent(godoCtx, outcome, "title")
			outcomeColor := h.readStringFromEvent(godoCtx, outcome, "color")

			outcomeDict := godoCtx.Dictionary()
			outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("id")), godoCtx.Variant(outcomeID))
			outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("title")), godoCtx.Variant(outcomeName))
			outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("color")), godoCtx.Variant(outcomeColor))

			if !onlyBeginning {
				outcomeUsers := h.readIntFromEvent(godoCtx, outcome, "users")
				outcomePoints := h.readIntFromEvent(godoCtx, outcome, "channel_points")

				topPredictorsArray := gd.NewArrayOf[gd.Dictionary](godoCtx)
				topPredictorsInterface, ok := eventMsg.Payload.Event["top_predictors"]
				if !ok {
					lib.LogErr(godoCtx, "prediction without any top_predictors")
				} else {
					topPredictors := topPredictorsInterface.([]interface{})
					if len(topPredictors) > 0 {
						for _, topPredictorInterface := range topPredictors {
							topPredictor, ok := topPredictorInterface.(map[string]interface{})
							if !ok {
								lib.LogErr(godoCtx, fmt.Sprintf("error converting single top predictor: expected map[string]interface{} but got %T", topPredictorInterface))
								return nil
							}

							tpID := h.readStringFromEvent(godoCtx, topPredictor, "user_id")
							tpName := h.readStringFromEvent(godoCtx, topPredictor, "user_name")
							tpPoints := h.readIntFromEvent(godoCtx, topPredictor, "channel_points_used")

							topPredictorDict := godoCtx.Dictionary()
							topPredictorDict.SetIndex(godoCtx.Variant(godoCtx.String("user_id")), godoCtx.Variant(tpID))
							topPredictorDict.SetIndex(godoCtx.Variant(godoCtx.String("user_name")), godoCtx.Variant(tpName))
							topPredictorDict.SetIndex(godoCtx.Variant(godoCtx.String("channel_points_used")), godoCtx.Variant(tpPoints))
							if withWinnings {
								tpWon := h.readIntFromEvent(godoCtx, topPredictor, "channel_points_won")
								topPredictorDict.SetIndex(godoCtx.Variant(godoCtx.String("channel_points_won")), godoCtx.Variant(tpWon))
							}

							topPredictorsArray.Append(topPredictorDict)
						}
					}
				}

				outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("users")), godoCtx.Variant(outcomeUsers))
				outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("channel_points")), godoCtx.Variant(outcomePoints))
				outcomeDict.SetIndex(godoCtx.Variant(godoCtx.String("top_predictors")), godoCtx.Variant(topPredictorsArray))
			}

			outcomesArray.Append(outcomeDict)
		}
	}

	return outcomesArray
}

func main() {
	godot, ok := gdextension.Link()
	if !ok {
		panic("Unable to link to godot")
	}
	gd.Register[GodotTwitch](godot)
}

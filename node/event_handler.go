package node

import (
	"fmt"
	"main/lib"
	"math"
	"time"

	"github.com/nicklaw5/helix/v2"
	"grow.graphics/gd"
)

func (h *GodotTwitch) handleEventTick(godoCtx gd.Context) {
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

func (h *GodotTwitch) handleEvent(godoCtx gd.Context, eventMsg lib.TwitchMessage) {
	if eventMsg.Payload.Subscription == nil {
		fmt.Printf("%+v\n", eventMsg)
		lib.LogWarn(godoCtx, fmt.Sprintf("received non subscribtion event: %s", eventMsg.Metadata.Type))
		return
	}

	lib.LogInfo(godoCtx, fmt.Sprintf("received event for: %s", eventMsg.Payload.Subscription.Type))

	switch eventMsg.Payload.Subscription.Type {
	case helix.EventSubTypeChannelFollow:
		tmpName := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")

		h.LatestFollower.Free()
		h.LatestFollower = h.Pin().String(tmpName.String())

		h.OnFollow.Emit(tmpName)
	case helix.EventSubTypeChannelSubscription:
		tmpName := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")

		h.LatestSubscriber.Free()
		h.LatestSubscriber = h.Pin().String(tmpName.String())

		// only emit signal for non gift subs as we emit from the gift event for gifts and we
		// we do not want double events
		if !eventMsg.Payload.Event["is_gift"].(bool) {
			tier := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "tier")
			h.OnSubscibtion.Emit(tmpName, gd.Int(1), tier)
		}
	case helix.EventSubTypeChannelSubscriptionMessage:
		tmpName := h.readStringFromEvent(godoCtx, eventMsg.Payload.Event, "user_name")

		h.LatestSubscriber.Free()
		h.LatestSubscriber = h.Pin().String(tmpName.String())

		tier := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "tier")
		if tier >= 1000 {
			tier = tier % 1000
		}
		months := h.readIntFromEvent(godoCtx, eventMsg.Payload.Event, "cumulative_months")
		h.OnSubscibtion.Emit(tmpName, months, tier)
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

func (h *GodotTwitch) handleApiUpdateTick(godoCtx gd.Context) {
	h.apiInfoResponseLock.Lock()
	defer h.apiInfoResponseLock.Unlock()

	if len(h.apiInfoResponseQueue) <= 0 {
		return
	}

	for _, apiInfo := range h.apiInfoResponseQueue {
		switch apiInfo := apiInfo.(type) {

		case LatestFollowerUpdate:
			if h.LatestFollower.String() != "" {
				lib.LogInfo(godoCtx, "non empty follower string on api update. skip api update because event should bee more up to date")
				continue
			}
			h.LatestFollower.Free()
			h.LatestFollower = h.Pin().String(apiInfo.Username)

		case LatestSubscriberUpdate:
			if h.LatestSubscriber.String() != "" {
				lib.LogInfo(godoCtx, "non empty follower string on api update. skip api update because event should bee more up to date")
				continue
			}
			h.LatestSubscriber.Free()
			h.LatestSubscriber = h.Pin().String(apiInfo.Username)
		}
	}
	h.apiInfoResponseQueue = make([]interface{}, 0)
}

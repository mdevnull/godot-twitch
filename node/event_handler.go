package node

import (
	"fmt"
	"main/lib"
	"time"

	"github.com/nicklaw5/helix/v2"
	"graphics.gd/variant/Float"
)

func (h *GodotTwitch) handleEventTick() {
	h.eventProcessLock.Lock()
	defer h.eventProcessLock.Unlock()

	if len(h.eventProcessQueue) <= 0 {
		return
	}

	for _, msg := range h.eventProcessQueue {
		h.handleEvent(msg)
	}
	h.eventProcessQueue = make([]lib.TwitchMessage, 0)
}

func (h *GodotTwitch) handleEvent(eventMsg lib.TwitchMessage) {
	if eventMsg.Payload.Subscription == nil {
		fmt.Printf("%+v\n", eventMsg)
		lib.LogWarn(fmt.Sprintf("received non subscribtion event: %s", eventMsg.Metadata.Type))
		return
	}

	lib.LogInfo(fmt.Sprintf("received event for: %s", eventMsg.Payload.Subscription.Type))

	switch eventMsg.Payload.Subscription.Type {
	case helix.EventSubTypeChannelFollow:
		tmpName := h.readStringFromEvent(eventMsg.Payload.Event, "user_name")

		h.LatestFollower = tmpName

		h.OnFollow.Emit(tmpName)
	case helix.EventSubTypeChannelSubscription:
		tmpName := h.readStringFromEvent(eventMsg.Payload.Event, "user_name")

		h.LatestSubscriber = tmpName

		// only emit signal for non gift subs as we emit from the gift event for gifts and we
		// we do not want double events
		if !eventMsg.Payload.Event["is_gift"].(bool) {
			tier := h.readIntFromEvent(eventMsg.Payload.Event, "tier")
			h.OnSubscibtion.Emit(tmpName, 1, tier)
		}
	case helix.EventSubTypeChannelSubscriptionMessage:
		tmpName := h.readStringFromEvent(eventMsg.Payload.Event, "user_name")

		h.LatestSubscriber = tmpName

		tier := h.readIntFromEvent(eventMsg.Payload.Event, "tier")
		if tier >= 1000 {
			tier = tier % 1000
		}
		months := h.readIntFromEvent(eventMsg.Payload.Event, "cumulative_months")
		h.OnSubscibtion.Emit(tmpName, months, tier)
	case helix.EventSubTypeChannelSubscriptionGift:
		isAnonymous := h.readBoolFromEvent(eventMsg.Payload.Event, "is_anonymous")
		var gifterName string
		var totalAmountForUser int
		if !isAnonymous {
			gifterName = h.readStringFromEvent(eventMsg.Payload.Event, "user_name")
			totalAmountForUser = h.readIntFromEvent(eventMsg.Payload.Event, "cumulative_total")
		} else {
			gifterName = ""
		}
		tier := h.readIntFromEvent(eventMsg.Payload.Event, "tier")
		giftAmount := h.readIntFromEvent(eventMsg.Payload.Event, "total")

		h.OnGiftSubs.Emit(gifterName, giftAmount, tier, totalAmountForUser)
	case helix.EventSubTypeChannelRaid:
		fromUserID := h.readStringFromEvent(eventMsg.Payload.Event, "from_broadcaster_user_id")
		fromUser := h.readStringFromEvent(eventMsg.Payload.Event, "from_broadcaster_user_name")
		viewerCount := h.readIntFromEvent(eventMsg.Payload.Event, "viewers")

		userResp, err := h.twitchClient.GetUsers(&helix.UsersParams{
			IDs: []string{fromUserID},
		})
		if err != nil {
			lib.LogErr(fmt.Sprintf("unable to fetch user %s: %s", fromUserID, err.Error()))
			return
		}
		if len(userResp.Data.Users) <= 0 {
			lib.LogErr(fmt.Sprintf("unable to fetch user %s: empty result", fromUserID))
			return
		}

		userObj := userResp.Data.Users[0]
		profilePicUrl := userObj.ProfileImageURL

		h.OnIncomingRaid.Emit(fromUser, profilePicUrl, viewerCount)
	case helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd:
		rewardInterface, ok := eventMsg.Payload.Event["reward"]
		if !ok {
			lib.LogErr("missing event data: reward")
			return
		}

		reward := rewardInterface.(map[string]interface{})
		fromUser := h.readStringFromEvent(eventMsg.Payload.Event, "user_name")
		userInput := h.readStringFromEvent(eventMsg.Payload.Event, "user_input")
		rewardID := h.readStringFromEvent(reward, "id")
		rewardTitle := h.readStringFromEvent(reward, "title")
		rewardPrompt := h.readStringFromEvent(reward, "prompt")
		rewardCost := h.readIntFromEvent(reward, "cost")

		h.OnRewardRedemtionAdd.Emit(
			fromUser, userInput,
			rewardID, rewardTitle, rewardPrompt, rewardCost,
		)
	case helix.EventSubShoutoutCreate:
		broadcasterID := h.readStringFromEvent(eventMsg.Payload.Event, "to_broadcaster_user_id")
		broadcasterName := h.readStringFromEvent(eventMsg.Payload.Event, "to_broadcaster_user_name")

		userResp, err := h.twitchClient.GetUsers(&helix.UsersParams{
			IDs: []string{broadcasterID},
		})
		if err != nil {
			lib.LogErr(fmt.Sprintf("unable to fetch user %s: %s", broadcasterID, err.Error()))
			return
		}
		if len(userResp.Data.Users) <= 0 {
			lib.LogErr(fmt.Sprintf("unable to fetch user %s: empty result", broadcasterID))
			return
		}

		userObj := userResp.Data.Users[0]
		profilePicUrl := userObj.ProfileImageURL

		channelInfo, err := h.twitchClient.GetChannelInformation(&helix.GetChannelInformationParams{
			BroadcasterIDs: []string{broadcasterID},
		})
		if err != nil {
			lib.LogErr(fmt.Sprintf("unable to fetch channel info for %s: %s", broadcasterID, err.Error()))
			return
		}
		if len(channelInfo.Data.Channels) <= 0 {
			lib.LogErr(fmt.Sprintf("unable to fetch channel %s: empty result", broadcasterID))
			return
		}

		channelObj := channelInfo.Data.Channels[0]
		lastGameName := channelObj.GameName
		lastStreamTitle := channelObj.Title

		h.OnShoutoutCreate.Emit(broadcasterName, profilePicUrl, lastGameName, lastStreamTitle)
	case helix.EventSubTypeCharityDonation:
		amountInterface, ok := eventMsg.Payload.Event["amount"]
		if !ok {
			lib.LogErr("missing event data: amount")
			return
		}

		amount := amountInterface.(map[string]interface{})

		donatorName := h.readStringFromEvent(eventMsg.Payload.Event, "user_name")
		value := h.readIntFromEvent(amount, "value")
		decimalPlaces := h.readIntFromEvent(amount, "decimal_places")
		currency := h.readStringFromEvent(amount, "currency")

		valueFloat := Float.X(value)
		if decimalPlaces > 0 {
			valueFloat = Float.X(value) / Float.Pow(10, Float.X(decimalPlaces))
		}

		h.OnDonation.Emit(donatorName, valueFloat, currency)
	case helix.EventSubTypeChannelPollBegin:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")
		endsAtStr := h.readStringFromEvent(eventMsg.Payload.Event, "ends_at")
		endsAt, err := time.Parse(time.RFC3339, endsAtStr)
		if err != nil {
			lib.LogErr(fmt.Sprintf("error converting timestamp: %s", err.Error()))
			return
		}

		choicesArray := h.readPollChoices(eventMsg, true)
		if choicesArray == nil {
			return
		}

		h.OnPollBegin.Emit(title, int(endsAt.Unix()), choicesArray)
	case helix.EventSubTypeChannelPollProgress:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")

		choicesArray := h.readPollChoices(eventMsg, false)
		if choicesArray == nil {
			return
		}

		h.OnPollProgress.Emit(title, choicesArray)
	case helix.EventSubTypeChannelPollEnd:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")

		choicesArray := h.readPollChoices(eventMsg, false)
		if choicesArray == nil {
			return
		}

		h.OnPollEnd.Emit(title, choicesArray)
	case helix.EventSubTypeChannelPredictionBegin:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")
		locksAtStr := h.readStringFromEvent(eventMsg.Payload.Event, "locks_at")
		locksAt, err := time.Parse(time.RFC3339, locksAtStr)
		if err != nil {
			lib.LogErr(fmt.Sprintf("error converting timestamp: %s", err.Error()))
			return
		}

		outcomesArray := h.readPredictionOutComes(eventMsg, true, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionBegin.Emit(title, int(locksAt.Unix()), outcomesArray)
	case helix.EventSubTypeChannelPredictionProgress:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(eventMsg, false, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionProgress.Emit(title, outcomesArray)
	case helix.EventSubTypeChannelPredictionLock:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(eventMsg, false, false)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionLock.Emit(title, outcomesArray)
	case helix.EventSubTypeChannelPredictionEnd:
		title := h.readStringFromEvent(eventMsg.Payload.Event, "title")

		outcomesArray := h.readPredictionOutComes(eventMsg, false, true)
		if outcomesArray == nil {
			return
		}

		h.OnPredictionEnd.Emit(title, outcomesArray)
	}
}

func (h *GodotTwitch) handleApiUpdateTick() {
	h.apiInfoResponseLock.Lock()
	defer h.apiInfoResponseLock.Unlock()

	if len(h.apiInfoResponseQueue) <= 0 {
		return
	}

	for _, apiInfo := range h.apiInfoResponseQueue {
		switch apiInfo := apiInfo.(type) {

		case LatestFollowerUpdate:
			if h.LatestFollower != "" {
				lib.LogInfo("non empty follower string on api update. skip api update because event should bee more up to date")
				continue
			}

			h.LatestFollower = apiInfo.Username

		case LatestSubscriberUpdate:
			if h.LatestSubscriber != "" {
				lib.LogInfo("non empty follower string on api update. skip api update because event should bee more up to date")
				continue
			}
			h.LatestSubscriber = apiInfo.Username
		}
	}
	h.apiInfoResponseQueue = make([]interface{}, 0)
}

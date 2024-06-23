package lib

import (
	"fmt"

	"github.com/nicklaw5/helix/v2"
	"grow.graphics/gd"
)

func EventSetup(
	godoCtx gd.Context, client *helix.Client,
	webSocketSessionID string,
	broadcasterUserID string,
) {
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelFollow,
		Version: "2",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
			ModeratorUserID:   broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelSubscription,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelSubscriptionMessage,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelSubscriptionGift,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelRaid,
		Version: "1",
		Condition: helix.EventSubCondition{
			ToBroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubShoutoutCreate,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
			ModeratorUserID:   broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeCharityDonation,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPollBegin,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPollProgress,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPollEnd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPredictionBegin,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPredictionProgress,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPredictionLock,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
	subEvent(godoCtx, client, &helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPredictionEnd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterUserID,
		},
		Transport: helix.EventSubTransport{Method: "websocket", SessionID: webSocketSessionID},
	})
}

func subEvent(godoCtx gd.Context, client *helix.Client, eventPayload *helix.EventSubSubscription) {
	subResp, err := client.CreateEventSubSubscription(eventPayload)
	if err != nil {
		LogErr(godoCtx, fmt.Sprintf("subscribtion for event %s failed: %s", eventPayload.Type, err.Error()))
		return
	}
	if subResp.Error != "" {
		LogErr(godoCtx, fmt.Sprintf(
			"event sub %s failed: %s - %s",
			eventPayload.Type,
			subResp.Error,
			subResp.ErrorMessage,
		))
		return
	}
	LogInfo(godoCtx, fmt.Sprintf("subscibed to event %s", eventPayload.Type))
}

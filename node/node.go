package node

import (
	"main/lib"
	"sync"

	"github.com/nicklaw5/helix/v2"
	"grow.graphics/gd"
)

type GodotTwitch struct {
	gd.Class[GodotTwitch, gd.Node]

	ClientID     gd.String `gd:"twitch_client_id"`
	ClientSecret gd.String `gd:"twitch_client_secret"`

	AuthURL         gd.String `gd:"auth_url"`
	UseDebugWS      gd.Bool   `gd:"use_debug_ws_server"`
	StoreToken      gd.Bool   `gd:"store_token"`
	IsAuthenticated gd.Bool   `gd:"is_authed"`

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

	apiInfoResponseLock  sync.Mutex
	apiInfoResponseQueue []interface{}

	hasNewToken bool
}

type (
	LatestFollowerUpdate struct {
		Username string
	}
	LatestSubscriberUpdate struct {
		Username string
	}
)

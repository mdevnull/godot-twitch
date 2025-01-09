package node

import (
	"main/lib"
	"sync"

	"github.com/nicklaw5/helix/v2"
	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Signal"
)

type GodotTwitch struct {
	classdb.Extension[GodotTwitch, Node.Instance] `gd:"GodotTwitch"
		allows you to react to Twitch Events inside Godot.`

	ClientID     string `gd:"twitch_client_id"`
	ClientSecret string `gd:"twitch_client_secret"`

	AuthURL string `gd:"auth_url"
		URI to open to authenticate with twitch`
	UseDebugWS bool `gd:"use_debug_ws_server"
		If true tries to load tokens from disk and stores new tokens to disk`
	StoreToken      bool `gd:"store_token"`
	IsAuthenticated bool `gd:"is_authed"
		True if client has been authenticated. This can be true when _ready if store_token is true and valid tokens are stored on disk`

	OnFollow Signal.Solo[string] `gd:"on_follow(username)"
		channel.follow`
	LatestFollower string `gd:"latest_follower"
		Username of latest follower`

	OnSubscibtion Signal.Trio[string, int, int] `gd:"on_subscribtion(username,months,tier)"
		Twitch Event: channel.subscribe ( only for non gifts ) and channel.subscription.message`
	OnGiftSubs Signal.Quad[string, int, int, int] `gd:"on_sub_gift(username,qty,tier,total)"
		Twitch Event: channel.subscription.gift, if Gifter is anonymous username may be empty and total may be zero`
	LatestSubscriber string `gd:"latest_subscriber"
		Username of latest subscriber`

	OnIncomingRaid Signal.Trio[string, string, int] `gd:"on_raid(username,profile_picture_url,viewer_count)"
		Twitch Event: channel.raid ( incoming raids )`
	OnRewardRedemtionAdd Signal.Hexa[string, string, string, string, string, int] `gd:"on_redeem(username,user_input,reward_id,reward_title,reward_prompt,reward_cost)"
		Twitch Event: channel.channel_points_custom_reward_redemption.add`
	OnShoutoutCreate Signal.Quad[string, string, string, string] `gd:"on_shoutout_create(username,profile_picture_url,last_stream_game,last_stream_title)"
		Twitch Event: channel.shoutout.create`
	OnDonation Signal.Trio[string, Float.X, string] `gd:"on_donation(username,amount,currency)"
		Twitch Event: channel.charity_campaign.donate`
	OnPollBegin Signal.Trio[string, int, []Choice] `gd:"on_poll_begin(title,unix_time,choices)"
		Twitch Event: channel.poll.begin`
	OnPollProgress Signal.Pair[string, []Choice] `gd:"on_poll_progress(title,choices)"
		Twitch Event: channel.poll.progress, includes bits_votes, channel_points_votes and votes`
	OnPollEnd Signal.Pair[string, []Choice] `gd:"on_poll_end(title,choices)"
		Twitch Event: channel.poll.end, includes bits_votes, channel_points_votes and votes`
	OnPredictionBegin Signal.Trio[string, int, []PredictionOutcome] `gd:"on_prediction_begin(title,unix_lock_time,outcomes)"
		Twitch Event: channel.prediction.begin`
	OnPredictionProgress Signal.Pair[string, []PredictionOutcome] `gd:"on_prediction_progress(title,outcomes)"
		Twitch Event: channel.prediction.begin, includes users, channel_points and top_predictors`
	OnPredictionLock Signal.Pair[string, []PredictionOutcome] `gd:"on_prediction_lock(title,outcomes)"
		Twitch Event: channel.prediction.lock, includes users, channel_points and top_predictors`
	OnPredictionEnd Signal.Pair[string, []PredictionOutcome] `gd:"on_prediction_end(title,outcomes)"
		Twitch Event: channel.prediction.end, includes users, channel_points and top_predictors`

	twitchClient *helix.Client

	eventProcessLock  sync.Mutex
	eventProcessQueue []lib.TwitchMessage

	apiInfoResponseLock  sync.Mutex
	apiInfoResponseQueue []interface{}

	hasNewToken bool
}

type Choice struct {
	ID                 string `gd:"id"`
	Title              string `gd:"title"`
	BitsVoted          int    `gd:"bits_votes"`
	ChannelPointsVoted int    `gd:"channel_points_votes"`
	Votes              int    `gd:"votes"`
}

type PredictionOutcome struct {
	ID            string         `gd:"id"`
	Title         string         `gd:"title"`
	Color         string         `gd:"color"`
	Users         int            `gd:"users"`
	ChannelPoints int            `gd:"channel_points"`
	TopPredictors []TopPredictor `gd:"top_predictors"`
}

type TopPredictor struct {
	UserID            string `gd:"user_id"`
	UserName          string `gd:"user_name"`
	ChannelPointsUsed int    `gd:"channel_points"`
	ChannelPointsWon  int    `gd:"channel_points_won"`
}

type (
	LatestFollowerUpdate struct {
		Username string
	}
	LatestSubscriberUpdate struct {
		Username string
	}
)

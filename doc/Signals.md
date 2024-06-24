# All Node signals with their parameters

|Signal name|Twitch event|Param 1|Param 2|Param 3|Param 4|Param 5|Param 6|
|---|---|---|---|---|---|---|---|
|on_follow|channel.follow|Username||||||
|on_subscribtion|channel.subscribe ( only for non gifts ) and channel.subscription.message|Username|Months|Tier (1,2,3)||||
|on_sub_gift|channel.subscription.gift|Gifter username ( may be empty if anonymous )|Number of gifted subs|Tier|Total amount of subs ( might be zero if anonymous )|||
|on_raid|channel.raid ( incoming raids )|Broadcaster Username|Profile picture URL|viewer count||||
|on_redeem|channel.channel_points_custom_reward_redemption.add|Username|User Input if any|Reward ID|Reward title|Reward promt|Reward cost|
|on_shoutout_create|channel.shoutout.create|Username|User Profile picture|Last stream category/game|Last stream title|||
|on_donation|channel.charity_campaign.donate|Username|Donation amount|Currency||||
|on_poll_begin|channel.poll.begin|Title|Unix Timestamp of end time|Array of Dictionaries of choices|||
|on_poll_progress|channel.poll.progress|Title|Array of Dictionaries of choices|||||
|on_poll_end|channel.poll.end|Title|Array of Dictionaries of choices|||||
|on_prediction_begin|channel.prediction.begin|Title|Unix timestamp of lock time|Array of Dictionaries of outcomes||||
|on_prediction_progress|channel.prediction.progress|Title|Array of Dictionaries of outcomes|||||
|on_prediction_lock|channel.prediction.lock|Title|Array of Dictionaries of outcomes|||||
|on_prediction_end|channel.prediction.end|Title|Array of Dictionaries of outcomes|||||

# Choices dictionary keys

* id
* title
* bits_votes ( only on progress and end )
* channel_points_votes ( only on progress and end )
* votes ( only on progress and end )

# Outcomes dictionary keys

* id
* title
* color
* users ( not on begin event; number of users )
* channel_points ( not on begin event )
* top_predictors ( not on begin event; array of dictionaries again )

# Top predictors dictionary keys

* user_id
* user_name
* channel_points_used
* channel_points_won ( only on end )
package lib

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/huh"
	"github.com/nicklaw5/helix/v2"
)

type EventItem struct {
	title       string
	description string
	twitchEvent string
	MakeForm    func() *huh.Form
	MakePayload func(*huh.Form) string
}

func (i EventItem) Title() string       { return i.title }
func (i EventItem) Description() string { return i.description }
func (i EventItem) FilterValue() string { return i.title }
func (i EventItem) TwitchIdent() string { return i.twitchEvent }

var pollChoiseList = []string{"Yeah!", "Yes!", "Yep!", "Heck Yeah!", "Uh Yeah!", "ofc", "fr fr", "based", "slay", "no"}
var predictionColors = []string{"blue", "pink", "blue", "pink", "blue", "pink", "blue", "pink", "blue", "pink"}

var EventItems = []list.Item{
	EventItem{
		title:       "Follow",
		twitchEvent: helix.EventSubTypeChannelFollow,
		description: "Test new follower event",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("username").Title("Username").Prompt("?"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			return strings.ReplaceAll(fmt.Sprintf(`{
	"metadata": {
		"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
		"message_type": "notification",
		"message_timestamp": "2022-11-16T10:11:12.464757833Z",
		"subscription_type": "channel.follow",
		"subscription_version": "2"
	},
	"payload": {
		"subscription": {
				"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
				"type": "channel.follow",
				"version": "2",
				"status": "enabled",
				"cost": 0,
				"condition": {
					"broadcaster_user_id": "1337",
					"moderator_user_id": "1337"
				},
				"transport": {
						"method": "webhook",
						"callback": "https://example.com/webhooks/callback"
				},
				"created_at": "2019-11-16T10:11:12.634234626Z"
		},
		"event": {
				"user_id": "1234",
				"user_login": "%s",
				"user_name": "%s",
				"broadcaster_user_id": "1337",
				"broadcaster_user_login": "cooler_user",
				"broadcaster_user_name": "Cooler_User",
				"followed_at": "2020-07-15T18:16:11.17106713Z"
		}
	}
}`, strings.ToLower(f.GetString("username")), f.GetString("username")), "\n", "")
		},
	},
	EventItem{
		title:       "Sub",
		twitchEvent: helix.EventSubTypeChannelSubscription,
		description: "Test a new subsciber",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("username").Title("Username").Prompt("?"),
					huh.NewSelect[int]().Key("tier").Title("Tier").Options(
						huh.NewOption("T1", 1000),
						huh.NewOption("T2", 2000),
						huh.NewOption("T3", 3000),
					),
					huh.NewConfirm().Key("is_gift").Title("Is gift Sub"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			return strings.ReplaceAll(
				fmt.Sprintf(
					`{
						"metadata": {
							"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
							"message_type": "notification",
							"message_timestamp": "2022-11-16T10:11:12.464757833Z",
							"type": "channel.subscribe",
							"subscription_version": "1"
						},
						"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.subscribe",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
										"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"user_id": "1234",
									"user_login": "%s",
									"user_name": "%s",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cooler_user",
									"broadcaster_user_name": "Cooler_User",
									"tier": "%d",
									"is_gift": %t
							}
						}
					}`,
					strings.ToLower(f.GetString("username")),
					f.GetString("username"),
					f.GetInt("tier"),
					f.GetBool("is_gift"),
				), "\n", "")
		},
	},
	EventItem{
		title:       "Ongoing sub",
		twitchEvent: helix.EventSubTypeChannelSubscriptionMessage,
		description: "Test sub with message",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("username").Title("Username").Prompt("?"),
					huh.NewSelect[int]().Key("tier").Title("Tier").Options(
						huh.NewOption("T1", 1000),
						huh.NewOption("T2", 2000),
						huh.NewOption("T3", 3000),
					),
				),
				huh.NewGroup(
					huh.NewInput().Key("cumulative_months").Title("Cumulative"),
					huh.NewInput().Key("streak_months").Title("Streak"),
					huh.NewInput().Key("duration_months").Title("Duration"),
				).Title("Duration info"),
			)
		},
		MakePayload: func(f *huh.Form) string {
			cumuMonths, _ := strconv.Atoi(f.GetString("cumulative_months"))
			streakMonths, _ := strconv.Atoi(f.GetString("streak_months"))
			durationMonths, _ := strconv.Atoi(f.GetString("duration_months"))
			return strings.ReplaceAll(
				fmt.Sprintf(`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
						"subscription": {
								"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
								"type": "channel.subscription.message",
								"version": "1",
								"status": "enabled",
								"cost": 0,
								"condition": {
									"broadcaster_user_id": "1337"
								},
								"transport": {
										"method": "webhook",
										"callback": "https://example.com/webhooks/callback"
								},
								"created_at": "2019-11-16T10:11:12.634234626Z"
						},
						"event": {
								"user_id": "1234",
								"user_login": "%s",
								"user_name": "%s",
								"broadcaster_user_id": "1337",
								"broadcaster_user_login": "cooler_user",
								"broadcaster_user_name": "Cooler_User",
								"tier": "%d",
								"message": {
										"text": "Love the stream! FevziGG",
										"emotes": [
												{
														"begin": 23,
														"end": 30,
														"id": "302976485"
												}
										]
								},
								"cumulative_months": %d,
								"streak_months": %d,
								"duration_months": %d
						}
					}
				}`,
					strings.ToLower(f.GetString("username")),
					f.GetString("username"),
					f.GetInt("tier"),
					cumuMonths,
					streakMonths,
					durationMonths,
				), "\n", "")
		},
	},
	EventItem{
		title:       "Gift subs",
		twitchEvent: helix.EventSubTypeChannelSubscriptionGift,
		description: "Test gift subs",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("username").Title("Username").Prompt("?"),
					huh.NewSelect[int]().Key("tier").Title("Tier").Options(
						huh.NewOption("T1", 1000),
						huh.NewOption("T2", 2000),
						huh.NewOption("T3", 3000),
					),
					huh.NewInput().Key("total").Title("Gift amount"),
					huh.NewInput().Key("cumulative_total").Title("Total gifts from user"),
					huh.NewConfirm().Key("is_anonymous").Title("Is anonymous"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			amount, _ := strconv.Atoi(f.GetString("total"))
			cumuAmounmt, _ := strconv.Atoi(f.GetString("cumulative_total"))
			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.subscription.gift",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
										"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"user_id": "1234",
									"user_login": "%s",
									"user_name": "%s",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cooler_user",
									"broadcaster_user_name": "Cooler_User",
									"total": %d,
									"tier": "%d",
									"cumulative_total": %d,
									"is_anonymous": %t
							}
					}
				}`,
				strings.ToLower(f.GetString("username")),
				f.GetString("username"),
				amount,
				f.GetInt("tier"),
				cumuAmounmt,
				f.GetBool("is_anonymous"),
			)
		},
	},
	EventItem{
		title:       "Incoming raid",
		twitchEvent: helix.EventSubTypeChannelRaid,
		description: "Test an incoming raid",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Key("streamer_id").
						Title("Streamer user ID").
						Description("This must be real so we can fetch some more user info"),
					huh.NewInput().Key("streamer").Title("Streamer username").Prompt("?"),
					huh.NewInput().Key("viewer_count").Title("Viewer count"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			viewers, _ := strconv.Atoi(f.GetString("viewer_count"))
			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.raid",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"to_broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"from_broadcaster_user_id": "%s",
									"from_broadcaster_user_login": "%s",
									"from_broadcaster_user_name": "%s",
									"to_broadcaster_user_id": "1337",
									"to_broadcaster_user_login": "cooler_user",
									"to_broadcaster_user_name": "Cooler_User",
									"viewers": %d
							}
					}
				}`,
				f.GetString("streamer_id"),
				strings.ToLower(f.GetString("streamer")),
				f.GetString("streamer"),
				viewers,
			)
		},
	},
	EventItem{
		title:       "Redeem redeemed",
		twitchEvent: helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		description: "Test a custom reward redemption add",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("username").Title("Username").Prompt("?"),
					huh.NewInput().Key("user_input").Title("User input"),
					huh.NewInput().Key("reward_id").Title("Reward ID"),
					huh.NewInput().Key("reward_title").Title("Reward title"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.channel_points_custom_reward_redemption.add",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"id": "17fa2df1-ad76-4804-bfa5-a40ef63efe63",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cool_user",
									"broadcaster_user_name": "Cool_User",
									"user_id": "9001",
									"user_login": "%s",
									"user_name": "%s",
									"user_input": "%s",
									"status": "unfulfilled",
									"reward": {
											"id": "%s",
											"title": "%s",
											"cost": 100,
											"prompt": "reward prompt"
									},
									"redeemed_at": "2020-07-15T17:16:03.17106713Z"
							}
					}
				}`,
				f.GetString("username"),
				strings.ToLower(f.GetString("username")),
				f.GetString("user_input"),
				f.GetString("reward_id"),
				f.GetString("reward_title"),
			)
		},
	},
	EventItem{
		title:       "Poll start",
		twitchEvent: helix.EventSubTypeChannelPollBegin,
		description: "Test a poll start",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("title").Title("Poll name").Prompt("?"),
					huh.NewInput().Key("duration").Title("Poll duration"),
					huh.NewInput().
						Key("num_options").
						Title("Number of choices").
						Description("Choices will be just a bunch of nonsense for testing. Max is 10."),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			numOpts, _ := strconv.Atoi(f.GetString("num_options"))
			if numOpts <= 0 || numOpts > 10 {
				panic("Too few or too many choises")
			}

			optsStrSlice := make([]string, numOpts)
			for i := 0; i < numOpts; i++ {
				optsStrSlice[i] = fmt.Sprintf(`{"id": "%d", "title": "%s"}`, i, pollChoiseList[i])
			}

			duration, _ := strconv.Atoi(f.GetString("duration"))

			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.poll.begin",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"id": "1243456",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cool_user",
									"broadcaster_user_name": "Cool_User",
									"title": "%s",
									"choices": [ %s ],
									"bits_voting": {
											"is_enabled": true,
											"amount_per_vote": 10
									},
									"channel_points_voting": {
											"is_enabled": true,
											"amount_per_vote": 10
									},
									"started_at": "%s",
									"ends_at": "%s"
							}
					}
				}`,
				f.GetString("title"),
				strings.Join(optsStrSlice, ","),
				time.Now().Format(time.RFC3339),
				time.Now().Add(time.Minute*time.Duration(duration)).Format(time.RFC3339),
			)
		},
	},
	EventItem{
		title:       "Poll end",
		twitchEvent: helix.EventSubTypeChannelPollEnd,
		description: "Test a poll end",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("title").Title("Poll name").Prompt("?"),
					huh.NewInput().
						Key("num_options").
						Title("Number of choices").
						Description("Choices will be just a bunch of nonsense for testing. Max is 10."),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			numOpts, _ := strconv.Atoi(f.GetString("num_options"))
			if numOpts <= 0 || numOpts > 10 {
				panic("Too few or too many choises")
			}

			optsStrSlice := make([]string, numOpts)
			for i := 0; i < numOpts; i++ {
				bitVotes := rand.Intn(100)
				pointVotes := rand.Intn(100)
				optsStrSlice[i] = fmt.Sprintf(
					`{"id": "%d", "title": "%s", "bits_votes": %d, "channel_points_votes": %d, "votes": %d}`,
					i,
					pollChoiseList[i],
					bitVotes,
					pointVotes,
					bitVotes+pointVotes,
				)
			}

			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.poll.end",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"id": "1243456",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cool_user",
									"broadcaster_user_name": "Cool_User",
									"title": "%s",
									"choices": [%s],
									"bits_voting": {
											"is_enabled": true,
											"amount_per_vote": 10
									},
									"channel_points_voting": {
											"is_enabled": true,
											"amount_per_vote": 10
									},
									"status": "completed",
									"started_at": "%s",
									"ended_at": "%s"
							}
					}
				}`,
				f.GetString("title"),
				strings.Join(optsStrSlice, ","),
				time.Now().Add(-time.Minute).Format(time.RFC3339),
				time.Now().Format(time.RFC3339),
			)
		},
	},
	EventItem{
		title:       "Prediction start",
		twitchEvent: helix.EventSubTypeChannelPredictionBegin,
		description: "Test a prediction start",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("title").Title("Prediction question").Prompt("?"),
					huh.NewInput().Key("duration").Title("Prediction time until locked"),
					huh.NewInput().
						Key("num_options").
						Title("Number of outcomes").
						Description("Outcomes will be just a bunch of nonsense for testing. Max is 10."),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			numOpts, _ := strconv.Atoi(f.GetString("num_options"))
			if numOpts <= 0 || numOpts > 10 {
				panic("Too few or too many choises")
			}

			optsStrSlice := make([]string, numOpts)
			for i := 0; i < numOpts; i++ {
				optsStrSlice[i] = fmt.Sprintf(`{"id": "%d", "title": "%s", "color": "%s"}`, i, pollChoiseList[i], predictionColors[i])
			}

			duration, _ := strconv.Atoi(f.GetString("duration"))

			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.prediction.begin",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"id": "1243456",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cool_user",
									"broadcaster_user_name": "Cool_User",
									"title": "%s",
									"outcomes": [ %s ],
									"started_at": "%s",
									"locks_at": "%s"
							}
					}
				}`,
				f.GetString("title"),
				strings.Join(optsStrSlice, ","),
				time.Now().Format(time.RFC3339),
				time.Now().Add(time.Minute*time.Duration(duration)).Format(time.RFC3339),
			)
		},
	},
	EventItem{
		title:       "Prediction end",
		twitchEvent: helix.EventSubTypeChannelPredictionEnd,
		description: "Test a prediction end ( missing a bunch of options but I got tired )",
		MakeForm: func() *huh.Form {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Key("title").Title("Prediction question").Prompt("?"),
				),
			)
		},
		MakePayload: func(f *huh.Form) string {
			return fmt.Sprintf(
				`{
					"metadata": {
						"message_id": "befa7b53-d79d-478f-86b9-120f112b044e",
						"message_type": "notification",
						"message_timestamp": "2022-11-16T10:11:12.464757833Z",
						"type": "channel.subscribe",
						"subscription_version": "1"
					},
					"payload": {
							"subscription": {
									"id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
									"type": "channel.prediction.end",
									"version": "1",
									"status": "enabled",
									"cost": 0,
									"condition": {
											"broadcaster_user_id": "1337"
									},
									"transport": {
											"method": "webhook",
											"callback": "https://example.com/webhooks/callback"
									},
									"created_at": "2019-11-16T10:11:12.634234626Z"
							},
							"event": {
									"id": "1243456",
									"broadcaster_user_id": "1337",
									"broadcaster_user_login": "cool_user",
									"broadcaster_user_name": "Cool_User",
									"title": "%s",
									"winning_outcome_id": "12345",
									"outcomes": [
											{
													"id": "12345",
													"title": "Yeah!",
													"color": "blue",
													"users": 2,
													"channel_points": 15000,
													"top_predictors": [
															{
																	"user_name": "Cool_User",
																	"user_login": "cool_user",
																	"user_id": "1234",
																	"channel_points_won": 10000,
																	"channel_points_used": 500
															},
															{
																	"user_name": "Coolest_User",
																	"user_login": "coolest_user",
																	"user_id": "1236",
																	"channel_points_won": 5000,
																	"channel_points_used": 100
															},
													]
											},
											{
													"id": "22435",
													"title": "No!",
													"users": 2,
													"channel_points": 200,
													"color": "pink",
													"top_predictors": [
															{
																	"user_name": "Cooler_User",
																	"user_login": "cooler_user",
																	"user_id": 12345,
																	"channel_points_won": null,
																	"channel_points_used": 100
															},
															{
																	"user_name": "Elite_User",
																	"user_login": "elite_user",
																	"user_id": 1337,
																	"channel_points_won": null,
																	"channel_points_used": 100
															}
													]
											}
									],
									"status": "resolved",
									"started_at": "%s",
									"ended_at": "%s"
							}
					}
				}`,
				f.GetString("title"),
				time.Now().Add(-time.Minute).Format(time.RFC3339),
				time.Now().Format(time.RFC3339),
			)
		},
	},
}

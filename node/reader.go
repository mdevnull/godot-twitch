package node

import (
	"fmt"
	"main/lib"
	"strconv"
)

func (h *GodotTwitch) readStringFromEvent(eventPayload map[string]interface{}, key string) string {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(fmt.Sprintf("missing event data: %s", key))
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%.2f", v)
	case float64:
		return fmt.Sprintf("%.2f", v)
	case nil:
		return ""
	default:
		lib.LogWarn(fmt.Sprintf("unknown event data type %T", val))
		return ""
	}
}

func (h *GodotTwitch) readIntFromEvent(eventPayload map[string]interface{}, key string) int {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(fmt.Sprintf("missing event data: %s", key))
		return 0
	}

	switch v := val.(type) {
	case string:
		asint, err := strconv.Atoi(v)
		if err != nil {
			lib.LogErr(fmt.Sprintf("unable to read %s as int: %s", key, err.Error()))
			return 0
		}
		return asint
	case int:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case nil:
		return int(0)
	default:
		lib.LogWarn(fmt.Sprintf("unknown event data type %T", val))
		return int(0)
	}
}

func (h *GodotTwitch) readBoolFromEvent(eventPayload map[string]interface{}, key string) bool {
	val, ok := eventPayload[key]
	if !ok {
		lib.LogErr(fmt.Sprintf("missing event data: %s", key))
		return false
	}

	valAsBoo, isBool := val.(bool)
	if !isBool {
		lib.LogWarn(fmt.Sprintf("cannot read %T as bool", val))
		return false
	}

	return valAsBoo
}

func (h *GodotTwitch) readPollChoices(
	eventMsg lib.TwitchMessage,
	onlyBeginning bool,
) []Choice {
	var choicesArray []Choice
	choicesInterfaces, ok := eventMsg.Payload.Event["choices"]
	if !ok {
		lib.LogErr("poll without any choices")
		return nil
	}
	choices := choicesInterfaces.([]interface{})
	if len(choices) > 0 {
		for _, choiceInterface := range choices {
			choice, ok := choiceInterface.(map[string]interface{})
			if !ok {
				lib.LogErr(fmt.Sprintf("error converting single choice: expected map[string]interface{} but got %T", choiceInterface))
				return nil
			}

			choiceID := h.readStringFromEvent(choice, "id")
			choiceName := h.readStringFromEvent(choice, "title")

			var choiceDict Choice
			choiceDict.ID = choiceID
			choiceDict.Title = choiceName

			if !onlyBeginning {
				bitsVoted := h.readIntFromEvent(choice, "bits_votes")
				pointsVoted := h.readIntFromEvent(choice, "channel_points_votes")
				totalVoted := h.readIntFromEvent(choice, "votes")

				choiceDict.BitsVoted = bitsVoted
				choiceDict.ChannelPointsVoted = pointsVoted
				choiceDict.Votes = totalVoted
			}

			choicesArray = append(choicesArray, choiceDict)
		}
	}

	return choicesArray
}

func (h *GodotTwitch) readPredictionOutComes(
	eventMsg lib.TwitchMessage,
	onlyBeginning bool, withWinnings bool,
) []PredictionOutcome {
	var outcomesArray []PredictionOutcome
	outcomesInterfaces, ok := eventMsg.Payload.Event["outcomes"]
	if !ok {
		lib.LogErr("prediction without any outcomes")
		return nil
	}
	outcomes := outcomesInterfaces.([]interface{})
	if len(outcomes) > 0 {
		for _, outcomeInterface := range outcomes {
			outcome, ok := outcomeInterface.(map[string]interface{})
			if !ok {
				lib.LogErr(fmt.Sprintf("error converting single outcome: expected map[string]interface{} but got %T", outcomeInterface))
				return nil
			}

			outcomeID := h.readStringFromEvent(outcome, "id")
			outcomeName := h.readStringFromEvent(outcome, "title")
			outcomeColor := h.readStringFromEvent(outcome, "color")

			var outcomeDict PredictionOutcome
			outcomeDict.ID = outcomeID
			outcomeDict.Title = outcomeName
			outcomeDict.Color = outcomeColor

			if !onlyBeginning {
				outcomeUsers := h.readIntFromEvent(outcome, "users")
				outcomePoints := h.readIntFromEvent(outcome, "channel_points")

				var topPredictorsArray []TopPredictor
				topPredictorsInterface, ok := eventMsg.Payload.Event["top_predictors"]
				if !ok {
					lib.LogErr("prediction without any top_predictors")
				} else {
					topPredictors := topPredictorsInterface.([]interface{})
					if len(topPredictors) > 0 {
						for _, topPredictorInterface := range topPredictors {
							topPredictor, ok := topPredictorInterface.(map[string]interface{})
							if !ok {
								lib.LogErr(fmt.Sprintf("error converting single top predictor: expected map[string]interface{} but got %T", topPredictorInterface))
								return nil
							}

							tpID := h.readStringFromEvent(topPredictor, "user_id")
							tpName := h.readStringFromEvent(topPredictor, "user_name")
							tpPoints := h.readIntFromEvent(topPredictor, "channel_points_used")

							var topPredictorDict TopPredictor
							topPredictorDict.UserID = tpID
							topPredictorDict.UserName = tpName
							topPredictorDict.ChannelPointsUsed = tpPoints
							if withWinnings {
								tpWon := h.readIntFromEvent(topPredictor, "channel_points_won")
								topPredictorDict.ChannelPointsWon = tpWon
							}

							topPredictorsArray = append(topPredictorsArray, topPredictorDict)
						}
					}
				}

				outcomeDict.Users = outcomeUsers
				outcomeDict.ChannelPoints = outcomePoints
				outcomeDict.TopPredictors = topPredictorsArray
			}

			outcomesArray = append(outcomesArray, outcomeDict)
		}
	}

	return outcomesArray
}

package node

import (
	"fmt"
	"main/lib"
	"strconv"

	"grow.graphics/gd"
)

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

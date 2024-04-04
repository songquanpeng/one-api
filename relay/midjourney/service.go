// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: service/midjourney.go
package midjourney

import (
	mjProvider "one-api/providers/midjourney"
	"strconv"
	"strings"
)

func CoverActionToModelName(mjAction string) string {
	modelName := "mj_" + strings.ToLower(mjAction)
	if mjAction == mjProvider.MjActionSwapFace {
		modelName = "swap_face"
	}
	return modelName
}

func GetMjRequestModel(relayMode int, midjRequest *mjProvider.MidjourneyRequest) (string, *mjProvider.MidjourneyResponse, bool) {
	action := ""
	if relayMode == mjProvider.RelayModeMidjourneyAction {
		// plus request
		err := CoverPlusActionToNormalAction(midjRequest)
		if err != nil {
			return "", err, false
		}
		action = midjRequest.Action
	} else {
		switch relayMode {
		case mjProvider.RelayModeMidjourneyImagine:
			action = mjProvider.MjActionImagine
		case mjProvider.RelayModeMidjourneyDescribe:
			action = mjProvider.MjActionDescribe
		case mjProvider.RelayModeMidjourneyBlend:
			action = mjProvider.MjActionBlend
		case mjProvider.RelayModeMidjourneyShorten:
			action = mjProvider.MjActionShorten
		case mjProvider.RelayModeMidjourneyChange:
			action = midjRequest.Action
		case mjProvider.RelayModeMidjourneyModal:
			action = mjProvider.MjActionModal
		case mjProvider.RelayModeMidjourneySwapFace:
			action = mjProvider.MjActionSwapFace
		case mjProvider.RelayModeMidjourneySimpleChange:
			params := ConvertSimpleChangeParams(midjRequest.Content)
			if params == nil {
				return "", mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "invalid_request"), false
			}
			action = params.Action
		case mjProvider.RelayModeMidjourneyTaskFetch, mjProvider.RelayModeMidjourneyTaskFetchByCondition, mjProvider.RelayModeMidjourneyNotify:
			return "", nil, true
		default:
			return "", mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "unknown_relay_action"), false
		}
	}

	modelName := CoverActionToModelName(action)
	return modelName, nil, true
}

func CoverPlusActionToNormalAction(midjRequest *mjProvider.MidjourneyRequest) *mjProvider.MidjourneyResponse {
	// "customId": "MJ::JOB::upsample::2::3dbbd469-36af-4a0f-8f02-df6c579e7011"
	customId := midjRequest.CustomId
	if customId == "" {
		return mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "custom_id_is_required")
	}
	splits := strings.Split(customId, "::")
	var action string
	if splits[1] == "JOB" {
		action = splits[2]
	} else {
		action = splits[1]
	}

	if action == "" {
		return mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "unknown_action")
	}
	if strings.Contains(action, "upsample") {
		index, err := strconv.Atoi(splits[3])
		if err != nil {
			return mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "index_parse_failed")
		}
		midjRequest.Index = index
		midjRequest.Action = mjProvider.MjActionUpscale
	} else if strings.Contains(action, "variation") {
		midjRequest.Index = 1
		if action == "variation" {
			index, err := strconv.Atoi(splits[3])
			if err != nil {
				return mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "index_parse_failed")
			}
			midjRequest.Index = index
			midjRequest.Action = mjProvider.MjActionVariation
		} else if action == "low_variation" {
			midjRequest.Action = mjProvider.MjActionLowVariation
		} else if action == "high_variation" {
			midjRequest.Action = mjProvider.MjActionHighVariation
		}
	} else if strings.Contains(action, "pan") {
		midjRequest.Action = mjProvider.MjActionPan
		midjRequest.Index = 1
	} else if strings.Contains(action, "reroll") {
		midjRequest.Action = mjProvider.MjActionReRoll
		midjRequest.Index = 1
	} else if action == "Outpaint" {
		midjRequest.Action = mjProvider.MjActionZoom
		midjRequest.Index = 1
	} else if action == "CustomZoom" {
		midjRequest.Action = mjProvider.MjActionCustomZoom
		midjRequest.Index = 1
	} else if action == "Inpaint" {
		midjRequest.Action = mjProvider.MjActionInPaint
		midjRequest.Index = 1
	} else {
		return mjProvider.MidjourneyErrorWrapper(mjProvider.MjRequestError, "unknown_action:"+customId)
	}
	return nil
}

func ConvertSimpleChangeParams(content string) *mjProvider.MidjourneyRequest {
	split := strings.Split(content, " ")
	if len(split) != 2 {
		return nil
	}

	action := strings.ToLower(split[1])
	changeParams := &mjProvider.MidjourneyRequest{}
	changeParams.TaskId = split[0]

	if action[0] == 'u' {
		changeParams.Action = "UPSCALE"
	} else if action[0] == 'v' {
		changeParams.Action = "VARIATION"
	} else if action == "r" {
		changeParams.Action = "REROLL"
		return changeParams
	} else {
		return nil
	}

	index, err := strconv.Atoi(action[1:2])
	if err != nil || index < 1 || index > 4 {
		return nil
	}
	changeParams.Index = index
	return changeParams
}

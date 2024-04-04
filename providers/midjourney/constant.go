// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: relay/constant/relay_mode.go
package midjourney

const (
	RelayModeUnknown = iota
	RelayModeMidjourneyImagine
	RelayModeMidjourneyDescribe
	RelayModeMidjourneyBlend
	RelayModeMidjourneyChange
	RelayModeMidjourneySimpleChange
	RelayModeMidjourneyNotify
	RelayModeMidjourneyTaskFetch
	RelayModeMidjourneyTaskImageSeed
	RelayModeMidjourneyTaskFetchByCondition
	RelayModeAudioSpeech
	RelayModeAudioTranscription
	RelayModeAudioTranslation
	RelayModeMidjourneyAction
	RelayModeMidjourneyModal
	RelayModeMidjourneyShorten
	RelayModeMidjourneySwapFace
)

// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: constant/midjourney.go

const (
	MjErrorUnknown = 5
	MjRequestError = 4
)

const (
	MjActionImagine       = "IMAGINE"
	MjActionDescribe      = "DESCRIBE"
	MjActionBlend         = "BLEND"
	MjActionUpscale       = "UPSCALE"
	MjActionVariation     = "VARIATION"
	MjActionReRoll        = "REROLL"
	MjActionInPaint       = "INPAINT"
	MjActionModal         = "MODAL"
	MjActionZoom          = "ZOOM"
	MjActionCustomZoom    = "CUSTOM_ZOOM"
	MjActionShorten       = "SHORTEN"
	MjActionHighVariation = "HIGH_VARIATION"
	MjActionLowVariation  = "LOW_VARIATION"
	MjActionPan           = "PAN"
	MjActionSwapFace      = "SWAP_FACE"
)

var MidjourneyModel2Action = map[string]string{
	"mj_imagine":        MjActionImagine,
	"mj_describe":       MjActionDescribe,
	"mj_blend":          MjActionBlend,
	"mj_upscale":        MjActionUpscale,
	"mj_variation":      MjActionVariation,
	"mj_reroll":         MjActionReRoll,
	"mj_modal":          MjActionModal,
	"mj_inpaint":        MjActionInPaint,
	"mj_zoom":           MjActionZoom,
	"mj_custom_zoom":    MjActionCustomZoom,
	"mj_shorten":        MjActionShorten,
	"mj_high_variation": MjActionHighVariation,
	"mj_low_variation":  MjActionLowVariation,
	"mj_pan":            MjActionPan,
	"swap_face":         MjActionSwapFace,
}

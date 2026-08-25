package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	SoakWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	AnodepitSwellSettleWindow  = 45 * time.Second
	PitfireWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxAnodepitLevelPercent    = 95.0
	MinAnodepitLevelPercent    = 15.0
	TripAnodepitLowPercent     = 10.0
	TripAnodepitHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinPitframeO2Percent    = 2.5
	MaxPitframeO2Percent    = 6.0
	DefaultJournalCapacity = 512
)

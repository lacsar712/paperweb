package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	DewaterWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	CouchSwellSettleWindow  = 45 * time.Second
	HeadboxWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxCouchLevelPercent    = 95.0
	MinCouchLevelPercent    = 15.0
	TripCouchLowPercent     = 10.0
	TripCouchHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinWireframeO2Percent    = 2.5
	MaxWireframeO2Percent    = 6.0
	DefaultJournalCapacity = 512
)

package entity

import "time"

// Dashboard reports system catalogue metrics, with user imports counted separately.
type Dashboard struct {
	GeneratedAt  time.Time            `json:"generatedAt"`
	Days         int                  `json:"days"`
	Totals       DashboardTotals      `json:"totals"`
	Period       DashboardPeriod      `json:"period"`
	VideoStatus  DashboardVideoStatus `json:"videoStatus"`
	Daily        []DashboardDay       `json:"daily"`
	RecentVideos []DashboardVideo     `json:"recentVideos"`
}

type DashboardTotals struct {
	Users                 int `json:"users"`
	ActiveUsers           int `json:"activeUsers"`
	Videos                int `json:"videos"`
	UserImportedVideos    int `json:"userImportedVideos"`
	Channels              int `json:"channels"`
	ActiveChannels        int `json:"activeChannels"`
	Captions              int `json:"captions"`
	VideosWithoutCaptions int `json:"videosWithoutCaptions"`
	Languages             int `json:"languages"`
	Topics                int `json:"topics"`
	Levels                int `json:"levels"`
}

type DashboardPeriod struct {
	NewUsers       int `json:"newUsers"`
	NewVideos      int `json:"newVideos"`
	ActiveLearners int `json:"activeLearners"`
}

type DashboardVideoStatus struct {
	Published int `json:"published"`
	Draft     int `json:"draft"`
	Archived  int `json:"archived"`
}

type DashboardDay struct {
	Date      string `json:"date"`
	NewUsers  int    `json:"newUsers"`
	NewVideos int    `json:"newVideos"`
}

type DashboardVideo struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Status       string    `json:"status"`
	ChannelName  string    `json:"channelName"`
	LanguageCode string    `json:"languageCode"`
	CaptionCount int       `json:"captionCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

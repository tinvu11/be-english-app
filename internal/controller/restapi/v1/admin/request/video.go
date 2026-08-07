package request

type SaveVideo struct {
	Title           string `json:"title" validate:"required,max=255" example:"Learn English with Music"`
	YouTubeID       string `json:"youtubeId" validate:"required,max=50" example:"dQw4w9WgXcQ"`
	ThumbnailURL    string `json:"thumbnailUrl" example:"https://img.youtube.com/vi/dQw4w9WgXcQ/maxresdefault.jpg"`
	DurationSeconds int    `json:"durationSeconds" validate:"min=0" example:"300"`
	Status          string `json:"status" validate:"required,oneof=draft published archived" example:"draft"`
	LanguageID      int    `json:"languageId" validate:"required,gt=0" example:"1"`
	LevelID         int    `json:"levelId" validate:"required,gt=0" example:"2"`
	ChannelID       int    `json:"channelId" validate:"required,gt=0" example:"3"`
	TopicIDs        []int  `json:"topicIds" validate:"dive,gt=0" example:"5,6"`
}

type VideoStatus struct {
	Status string `json:"status" validate:"required,oneof=published archived" example:"published"`
}

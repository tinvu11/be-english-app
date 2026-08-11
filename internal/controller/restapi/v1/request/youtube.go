package request

type YouTubePreview struct {
	YouTubeURL string `json:"youtubeUrl" validate:"required,max=2048" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
}

type ImportYouTubeVideo struct {
	YouTubeURL          string `json:"youtubeUrl" validate:"required,max=2048" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
	CaptionLanguageCode string `json:"captionLanguageCode" validate:"required,max=35" example:"en"`
}

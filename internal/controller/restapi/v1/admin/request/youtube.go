package request

type YouTubePreview struct {
	YouTubeURL string `json:"youtubeUrl" validate:"required,max=2048" example:"https://www.youtube.com/watch?v=dQw4w9WgXcQ"`
}

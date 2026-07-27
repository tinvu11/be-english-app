package response

// ContentList -.
type ContentList[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

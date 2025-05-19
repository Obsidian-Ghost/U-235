package models

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PageSize    int   `json:"page_size"`
	TotalCount  int64 `json:"total_count"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}
type PaginatedUrlsResponse struct {
	Urls []ShortenedUrlInfoRes `json:"urls"`
	Meta PaginationMeta        `json:"meta"`
}

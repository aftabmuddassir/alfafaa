package dto

// SearchQuery represents search query parameters
type SearchQuery struct {
	Query string `form:"q" binding:"required,min=2,max=100"`
	Type  string `form:"type" binding:"omitempty,oneof=articles categories tags all"`
	PaginationQuery
}

// GetType returns the search type with a default
func (s *SearchQuery) GetType() string {
	if s.Type == "" {
		return "all"
	}
	return s.Type
}

// SearchResponse represents a search response
type SearchResponse struct {
	Articles   []ArticleListItemResponse `json:"articles,omitempty"`
	Categories []CategoryResponse        `json:"categories,omitempty"`
	Tags       []TagResponse             `json:"tags,omitempty"`
}

// SearchResultItem represents a single search result item
type SearchResultItem struct {
	Type        string      `json:"type"`
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Slug        string      `json:"slug"`
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

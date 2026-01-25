package dto

// PaginationQuery represents pagination query parameters
type PaginationQuery struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Sort    string `form:"sort" binding:"omitempty,oneof=newest oldest popular alphabetical"`
}

// GetPage returns the page number with a default value
func (p *PaginationQuery) GetPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetPerPage returns the per page value with a default
func (p *PaginationQuery) GetPerPage() int {
	if p.PerPage <= 0 {
		return 20
	}
	if p.PerPage > 100 {
		return 100
	}
	return p.PerPage
}

// GetOffset calculates the offset for database queries
func (p *PaginationQuery) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPerPage()
}

// GetSort returns the sort field with a default
func (p *PaginationQuery) GetSort() string {
	if p.Sort == "" {
		return "newest"
	}
	return p.Sort
}

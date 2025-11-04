package utils

type PaginationMeta struct {
	Total   int64 `json:"total"`
	Pages   int   `json:"pages"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

// NewPaginationMeta membuat meta pagination
func NewPaginationMeta(total int64, page, perPage int) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage)) // ceil
	return PaginationMeta{
		Total:   total,
		Pages:   pages,
		Page:    page,
		PerPage: perPage,
	}
}

package domain

type PaginatorResponse[T any] struct {
	Items      []T
	TotalItems int64
	TotalPages int64
	PageSize   int64
	Page       int64
}

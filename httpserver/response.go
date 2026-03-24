package httpserver

import "net/http"

// HandlerResponse is returned by HandlerFunc delegates.
// Set Status to override the default 200 OK (e.g., 201 Created, 202 Accepted).
type HandlerResponse[T any] struct {
	Status     int         `json:"-"`
	Data       T           `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// StatusOrDefault returns the handler's intended status code, defaulting to 200.
func (h *HandlerResponse[T]) StatusOrDefault() int {
	if h.Status > 0 {
		return h.Status
	}

	return http.StatusOK
}

type Pagination struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"pageSize" example:"10"`
	TotalCount int `json:"totalCount" example:"42"`
	TotalPages int `json:"totalPages" example:"5"`
}

type APIResponse[T any] struct {
	RequestID  string      `json:"requestId,omitempty" example:"3bf74527-8097-4217-8485-ffe05d16f82e"`
	Data       T           `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

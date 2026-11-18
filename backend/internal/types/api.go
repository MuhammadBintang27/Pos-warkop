package types

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool         `json:"success"`
	Status  int          `json:"status"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

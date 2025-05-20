package models

// TokenPair содержит пару токенов доступа и обновления
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Response стандартный формат ответа API
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// UserResponse содержит информацию о пользователе
type UserResponse struct {
	UserID string `json:"user_id"`
}

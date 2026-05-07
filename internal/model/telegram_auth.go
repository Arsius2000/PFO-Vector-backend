package model

// TelegramAuthResponse — ответ сервера
type TelegramAuthResponse struct {
	Token   string `json:"token"`
	UserID  int32  `json:"user_id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type TelegramCheckResponse struct {
	Exists  bool   `json:"exists"`
	UserID  int32  `json:"user_id,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
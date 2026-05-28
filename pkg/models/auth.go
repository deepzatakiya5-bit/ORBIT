package models

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
}

type LoginResponse struct {
	User
	AccessToken string `json:"access_token"`
}

type CreateUserResponse struct {
	User
	AccessToken string `json:"access_token"`
}

type LoginInput struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

package models

import "time"

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"-"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Email       string `json:"email"`
	CreatedAt   string `json:"created_at"`
}

type AuthSession struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

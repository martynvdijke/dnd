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
	UserID    int64
	Username  string
	Role      string
	ExpiresAt time.Time
	IP        string
}

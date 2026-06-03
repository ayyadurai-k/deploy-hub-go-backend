package accounts

import "time"

type CreateUserRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type UserResponse struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active"`
	DateJoined  time.Time `json:"date_joined"`
}

func ToUserResponse(u User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		DateJoined:  u.DateJoined,
	}
}
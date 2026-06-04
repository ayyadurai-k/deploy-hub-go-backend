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
	HasGoogle   bool      `json:"has_google"`
	HasGithub   bool      `json:"has_github"`
}

func ToUserResponse(u User, hasGoogle, hasGithub bool) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		DateJoined:  u.DateJoined,
		HasGoogle:   hasGoogle,
		HasGithub:   hasGithub,
	}
}
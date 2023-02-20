package hub

import (
	"golang.org/x/net/context"
	"time"
)

type UserAPI interface {
	GetUserProfile(ctx context.Context, userId string) (reply *UserProfile, err error)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReply struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

type UserProfile struct {
	Name string `json:"name"`
}

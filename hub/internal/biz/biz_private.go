package biz

import (
	hub "github.com/koyeo/nest/hub/internal/api"
	"golang.org/x/net/context"
)

var _ hub.PrivateAPI = (*PrivateBiz)(nil)

type PrivateBiz struct {
	*PublishBiz
}

func NewPrivateBiz() *PrivateBiz {
	return &PrivateBiz{PublishBiz: &PublishBiz{}}
}

func (p PrivateBiz) GetUserProfile(ctx context.Context, userId string) (reply *hub.UserProfile, err error) {
	//TODO implement me
	panic("implement me")
}

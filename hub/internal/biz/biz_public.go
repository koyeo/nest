package biz

import (
	hub "github.com/koyeo/nest/hub/internal/api"
	"golang.org/x/net/context"
)

var _ hub.PublicAPI = (*PublicBiz)(nil)

type PublicBiz struct {
}

func NewPublicBiz() *PublicBiz {
	return &PublicBiz{}
}

func (p PublicBiz) Login(ctx context.Context, request *hub.LoginRequest) (reply *hub.LoginReply, err error) {
	//TODO implement me
	panic("implement me")
}

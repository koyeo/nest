package hub

import (
	"golang.org/x/net/context"
)

const Namespace = "hub"

type PublicAPI interface {
	Login(ctx context.Context, request *LoginRequest) (reply *LoginReply, err error)
}

type PrivateAPI interface {
	UserAPI
	PublishAPI
}

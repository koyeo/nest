package biz

import (
	"fmt"
	hub "github.com/koyeo/nest/hub/internal/api"
	"golang.org/x/net/context"
)

var _ hub.PublishAPI = (*PublishBiz)(nil)

type PublishBiz struct {
}

func (p PublishBiz) Publish(ctx context.Context, req *hub.PublishRequest) (err error) {
	
	fmt.Println("ok")
	
	return
}

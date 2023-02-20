package hub

import (
	"github.com/koyeo/nest/hub/internal/protocol"
	"golang.org/x/net/context"
)

type PublishAPI interface {
	Publish(ctx context.Context, req *PublishRequest) (err error)
}

type PublishRequest struct {
	Protocol protocol.NestJson
}

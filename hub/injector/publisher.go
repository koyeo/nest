package injector

import (
	"github.com/google/wire"
	"github.com/koyeo/nest/hub/internal/config"
	"github.com/koyeo/nest/hub/internal/services/publisher"
)

var PublisherSet = wire.NewSet(NewPublisher)

func NewPublisher(conf *config.Config) *publisher.Publisher {
	return publisher.NewPublisher(conf.Publisher)
}

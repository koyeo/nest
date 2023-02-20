package injector

import (
	"github.com/google/wire"
	hub "github.com/koyeo/nest/hub/internal/api"
	"github.com/koyeo/nest/hub/internal/biz"
	"gorm.io/gorm"
)

var ApiProviderSet = wire.NewSet(NewPublicAPI, NewPrivateAPI)

func NewPublicAPI(db *gorm.DB) hub.PublicAPI {
	return biz.NewPublicBiz()
}

func NewPrivateAPI(db *gorm.DB) hub.PrivateAPI {
	return biz.NewPrivateBiz()
}

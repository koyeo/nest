package core

import (
	"fmt"
	"testing"
)

func TestGoPackages(t *testing.T) {
	res := GoPackages([]byte(`
package service

import (
	"go.uber.org/zap"
	"mix/test/pb/core/transaction"
	transaction "mix/test/pb/core-mi/transaction"
	_ "mix/test/pb/core-mi/transaction"
	. "mix/test/pb/core-mi/transaction"
)

import "iwallet/test1"
import . "iwallet/test2"
import _ "iwallet/test3"

import ("iwallet/test4")
import (ok "iwallet/test5")
import (. "iwallet/test6")
import (_ "iwallet/test7")
import (_ "core/test7")

func (p *Transaction) login(ctx *Context, in *transaction.LoginInput, out *transaction.LoginOutput) (err error) {

	// logger := ctx.logger.With(zap.String("func", "login"))
	// dbSession := ctx.dbSession

	return
}
`))
	for _, v := range res {
		is, _ := IsProjectPackage(v)
		fmt.Println(v, is)
	}
}

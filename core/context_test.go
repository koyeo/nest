package core

import (
	"nest/config"
	"testing"
)

func TestContext_CheckScript(t *testing.T) {
	context := new(Context)
	_ = context.AddScript(&config.Script{
		Id:   "init.sh",
		Name: "测试脚本",
	})
	_ = context.AddScript(&config.Script{
		Id:   "test.sh",
		Name: "测试脚本",
	})
	//t.Log(context.CheckScript("test.sh@before(1,2,3,4)"))
	//t.Log(context.CheckScript("test.sh@before(1,2,3,4"))
	//t.Log(context.CheckScript("test.sh@befor(1,2,3,4"))
	//t.Log(context.CheckScript("test.sh"))
	//t.Log(context.CheckScript("test.sh@after(1,2,3,4"))
	//t.Log(context.CheckScript("test.sh@after(1,2,3,4)"))
	t.Log(context.CheckScript("init.sh@after($app=1,$log=2)"))
	//t.Log(context.CheckScript("test-1@#$1.sh@after(1,2,3,4"))
}

package publish

import (
	"fmt"
	"github.com/gozelle/gin"
	"io"
	"net/http"
	"os"
)

func Publish(ctx *gin.Context) {
	ctx.Request.ParseMultipartForm(32 << 20)
	//获取上传文件
	file, handler, err := ctx.Request.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(ctx.Writer, "%v", handler.Header)
	//创建上传目录
	os.Mkdir("./upload", os.ModePerm)
	//创建上传文件
	f, err := os.Create("./upload/" + handler.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	ctx.Writer.WriteHeader(http.StatusCreated)
	io.WriteString(ctx.Writer, "Uploaded successfully")
}

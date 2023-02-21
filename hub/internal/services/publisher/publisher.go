package publisher

import (
	"fmt"
	"github.com/gozelle/fs"
	"github.com/gozelle/logging"
	"github.com/koyeo/nest/hub/internal/config"
	"path/filepath"
	
	"net"
	"os"
)

func NewPublisher(conf *config.Publisher) *Publisher {
	return &Publisher{
		conf: conf,
		log:  logging.Logger("publisher"),
	}
}

type Publisher struct {
	conf *config.Publisher
	log  *logging.ZapEventLogger
	root string
}

func (p Publisher) Serve(addr string) (err error) {
	
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return
	}
	defer func() {
		_ = listener.Close()
	}()
	
	p.log.Infof("publisher listen on: %s", addr)
	
	err = p.prepareLocalStorage()
	if err != nil {
		return
	}
	
	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			return
		}
		p.handleRequest(conn)
	}
}

func (p *Publisher) prepareLocalStorage() (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	
	ls := filepath.Join(pwd, *p.conf.Storage)
	if !fs.Exists(ls) {
		err = fs.MakeDir(ls)
		if err != nil {
			return
		}
	}
	p.root = ls
	p.log.Infof("prepare local storage: %s", ls)
	
	return
}

func (p Publisher) handleRequest(conn net.Conn) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	
	filename := string(buf[:n])
	if filename != "" {
		_, err = conn.Write([]byte("ok"))
		if err != nil {
			fmt.Println("conn.Write err:", err)
			return
		}
	} else {
		return
	}
	
	p.log.Debugf("上传文件名: %s", filename)
	fp := filepath.Join(p.root, filename)
	file, err := os.Create(fp)
	if err != nil {
		fmt.Println("os.Create err:", err)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	
	p.log.Debugf("创建文件: %s", fp)
	
	for {
		n, err := conn.Read(buf)
		if n == 0 {
			fmt.Println("文件读取完毕")
			fmt.Println("文件上传错误")
			break
		}
		if err != nil {
			fmt.Println("conn.Read err:", err)
			return
		}
		_, err = file.Write(buf[:n])
		if err != nil {
			return
		}
	}
}

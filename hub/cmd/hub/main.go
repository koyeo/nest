package main

import (
	"encoding/json"
	"fmt"
	"github.com/gozelle/exit"
	"github.com/gozelle/gin"
	"github.com/gozelle/mix"
	"github.com/gozelle/toml"
	logging "github.com/ipfs/go-log/v2"
	hub "github.com/koyeo/nest/hub/internal/api"
	"github.com/koyeo/nest/hub/internal/config"
	"github.com/koyeo/nest/hub/internal/services/publisher"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"
	"gorm.io/gorm"
	"os"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "nesthub"
	// Version is the version of the compiled software.
	Version string
	
	id, _ = os.Hostname()
)

var (
	rootCmd        *cobra.Command
	configFilePath string
)

var (
	log = logging.Logger("main")
)

func init() {
	rootCmd = &cobra.Command{
		Use:     "syncer [-c|--config /path/to/config.toml]",
		Short:   Name,
		Run:     runCmd,
		Args:    cobra.ExactArgs(0),
		Version: Version,
	}
	
	rootCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "配置文件路径")
	rootCmd.Flags().SortFlags = false
	_ = rootCmd.MarkFlagRequired("config")
}

func newApp(conf *config.Config, db *gorm.DB, public hub.PublicAPI, private hub.PrivateAPI, publisher *publisher.Publisher) *gin.Engine {
	
	go func() {
		err := publisher.Serve(*conf.PublisherListen)
		if err != nil {
			panic(err)
			return
		}
	}()
	
	s := mix.NewServer()
	// 注册公开接口
	s.RegisterAPI(s.Group("/api/v1/public"), hub.Namespace, public)
	
	// 注册受保护的接口
	pg := s.Group("/api/v1")
	s.RegisterAPI(pg, hub.Namespace, private)
	
	return s.Engine
}

func runCmd(_ *cobra.Command, _ []string) {
	
	// 读取配置文件
	conf := &config.Config{}
	err := toml.UnmarshalFile(configFilePath, conf)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("配置文件:")
	d, _ := json.MarshalIndent(conf, "", "\t")
	fmt.Println(string(d))
	
	app, cleanup, err := wireApp(conf)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	
	go func() {
		// start and wait for stop signal
		if e := app.Run(*conf.APIListen); e != nil {
			panic(e)
		}
	}()
	
	exit.Clean(func() error {
		cleanup()
		return nil
	})
	exit.Wait()
}

func main() {
	_ = rootCmd.Execute()
}

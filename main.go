package main

import (
	"alb-manager/conf"
	"alb-manager/router"
	"alb-manager/utils"
	"fmt"
	"log"
)

func init() {
	var err error

	if err = conf.InitConfig(); err != nil {
		panic(fmt.Errorf("Fatal error read config file: %w \n", err))
	}

	if err = utils.InitLogger(conf.ViperConfig.GetString("logfilepath")); err != nil {
		panic(fmt.Errorf("Fatal error open log file: %w \n", err))
	}

	log.Println("open log file success")
}

func main() {

	r := router.InitRouter()

	r.Run(conf.ViperConfig.GetString("listenaddr"))
}

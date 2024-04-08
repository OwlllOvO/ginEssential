package main

import (
	"fmt"
	"os"
	"owlllovo/ginessential/common"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	InitConfig()
	db := common.InitDB()
	sqlDB, err := db.DB()
	if err != nil {
		panic("Failed to get database connection")
	}
	defer sqlDB.Close()

	r := gin.Default()
	r.Static("/images", "./assets/images")
	r = CollectRoute(r)
	port := viper.GetString("server.port")

	configerr := common.LoadConfig("./config.json")
	if configerr != nil {
		fmt.Println("Failed to load config:", configerr)
		return
	}
	fmt.Println("MaxTokens from config:", common.AppConfig.MaxTokens)

	if port != "" {
		panic(r.Run(":" + port)) // listen and serve on specified port in yml
	} else {
		panic(r.Run()) // listen and serve on default port 8080
	}
}

func InitConfig() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

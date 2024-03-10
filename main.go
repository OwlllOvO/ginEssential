package main

import (
	"owlllovo/ginessential/common"

	"github.com/gin-gonic/gin"
)

func main() {
	db := common.InitDB()
	sqlDB, err := db.DB()
	if err != nil {
		panic("Failed to get database connection")
	}
	defer sqlDB.Close()

	r := gin.Default()
	r = CollectRoute(r)
	panic(r.Run()) // listen and serve on 0.0.0.0:8080
}

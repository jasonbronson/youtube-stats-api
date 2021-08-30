package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

// RUN: go run main.go [-maxResults=NumberOfResults]
// default maxResults=25
var (
	channelList  string
	key        string
	maxResults = flag.Int64("maxResults", 25, "The maximum number of playlist resources to include in the API response.")
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Error: error loading .env file")
		os.Exit(1)
	}
	channelList = os.Getenv("CHANNEL_LIST")
	key = os.Getenv("KEY")
	if channelList == "" || key == "" {
		fmt.Println("channel id and key secret is required. please check env file")
		return
	}

	c := cron.New(cron.WithSeconds())
	c.AddFunc("@every 300s", downloadStats)
	go c.Start()

	r := gin.Default()
	r.Use(cors.New(*corsConfig))
	r.GET("/stats", getStats)
	r.Run(":9898")

}


package main

import (
	"github.com/gin-contrib/cors"
)

var (corsConfig *cors.Config)

func init() {
	c := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		//ExposeHeaders:    []string{"authtoken"},
	}
	corsConfig = &c
}

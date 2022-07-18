package router

import (
	"github.com/codecodify/randimages/service"
	"github.com/gin-gonic/gin"
)

func Start() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("views/*")
	r.Static("/static", "./static")

	r.GET("/", service.Index)

	ws := r.Group("/ws")
	{
		ws.GET("", service.Ws)
	}

	return r
}

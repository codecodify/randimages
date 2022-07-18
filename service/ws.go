package service

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Request 定义客户端请求
type Request struct {
	Num int `json:"num"`
}

// Response 定义客户端响应
type Response struct {
	Images []string `json:"images"`
	// Total 总共获取张数
	Total int `json:"total"`
	// Num 一次获取数量
	Num int `json:"num"`
}

// GetImageResponse 获取第三方图片响应
type GetImageResponse struct {
	Code   int    `json:"code"`
	Imgurl string `json:"imgurl"`
}

// Client ws客户端
type Client struct {
	Conn *websocket.Conn
	// Total 总共获取张数
	Total int
	// Num 一次获取数量
	Num int
}

func (c *Client) Read(quit chan<- struct{}) {
	message, body, err := c.Conn.ReadMessage()
	if err != nil {
		log.Printf("读取信息失败:%s\n", err)
		quit <- struct{}{}
		return
	}

	// 客户端关闭
	if message == websocket.CloseMessage {
		quit <- struct{}{}
		return
	}

	var req Request
	if err = json.Unmarshal(body, &req); err != nil {
		log.Printf("读取信息序列化失败:%s", err)
		quit <- struct{}{}
		return
	}

	c.Num = correctNum(req.Num)
}

func (c *Client) Write(quit chan<- struct{}, pusher chan<- struct{}) {
	var resp Response
	resp.Images = make([]string, 0)
	if c.Num > 0 {
		resp.Images = batchFetch(c.Num)
		c.Total += len(resp.Images)
	}

	resp.Num = c.Num
	resp.Total = c.Total

	b, _ := json.Marshal(resp)
	if err := c.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
		fmt.Printf("写入io流错误: %s\n", err)
		quit <- struct{}{}
	}
	// 休息5秒
	time.Sleep(5 * time.Second)
	pusher <- struct{}{}
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Ws(ctx *gin.Context) {
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.HTML(http.StatusBadGateway, "ws服务器出错", nil)
		return
	}
	defer conn.Close()
	var num int
	num, err = strconv.Atoi(ctx.DefaultQuery("num", "1"))
	if err != nil {
		num = 15
	}
	num = correctNum(num)

	client := &Client{
		Conn: conn,
		Num:  num,
	}

	// 错误信号
	quit := make(chan struct{})
	// 推送信号
	pusher := make(chan struct{})
Quit:
	for {
		go client.Read(quit)
		go client.Write(quit, pusher)

		select {
		case <-quit:
			break Quit
		case <-pusher:
			fmt.Println("ws推送")
			continue Quit
		}
	}
}

func correctNum(num int) int {
	if num < 0 {
		num = 0
	}

	if num > 20 {
		num = 20
	}

	return num
}

// 批量获取图片
func batchFetch(num int) []string {
	container := make(chan string, num)
	images := make([]string, 0, num)
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			container <- getImage()
		}()
	}
	wg.Wait()
	close(container)

	for image := range container {
		images = append(images, image)
	}

	return images

}

func getImage() string {
	resp, err := http.Get("https://api.uomg.com/api/rand.img3?format=json")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		var imageResp GetImageResponse
		_ = json.Unmarshal(body, &imageResp)
		return imageResp.Imgurl
	}

	return ""

	// 测试地址
	//return "https://gw2.alicdn.com/tfscom/tuitui/TB2TkOIp79WBuNjSspeXXaz5VXa_!!0-rate.jpg"
}

# 项目说明
前几天在论坛看到获取淘宝买家秀图片源码，使用HTML纯静态，用JQuery轮训请求第三方获取图片。
使用轮训方式实现起来简单，但是获取图片速度慢，体验不是很好，于是我决定用`websocket`来批量推送图片。

# 技术栈
* `gin`
* `websocket`

# 说明
后台在获取客户端的`io`流时，我是开了两个协程分别去处理输入流和输出流，一般批量获取到指定数量的图片后才开始处理下一次的输出流，否则在循环里容易
引起系统奔溃。
```go
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
```

批量获取图片的思路是使用缓冲通道来并发请求第三方：

```go
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
```

PS: 获取淘宝买家秀图片接口可能以后会失效，若失效想测试效果就需要自行修改代码。
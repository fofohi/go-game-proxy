package main

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
)

func main() {

	var address = "127.0.0.1:19077"

	// -------------------------------------------------------
	//  fasthttp 的 handler 处理函数
	// -------------------------------------------------------
	var requestHandler = func(ctx *fasthttp.RequestCtx) {

		// -------------------------------------------------------
		// 处理 web client 的请求数据
		// -------------------------------------------------------
		// 取出 web client 请求进行 TCP 连接的连接 ID
		var connID = strconv.FormatUint(ctx.ConnID(), 10)
		fmt.Println(connID)
		// 取出 web client 请求 HTTP header 中的事务ID
		var tid = string(ctx.Request.Header.PeekBytes([]byte("TransactionID")))
		if len(tid) == 0 {
			tid = "12345678"
		}

		// 取出 web 访问的 URL/URI
		var uriPath = ctx.Path()
		{
			// 取出 URI
		}

		// 取出 web client 请求的 URL/URI 中的参数部分
		{
			var uri = ctx.URI().QueryString()
			fmt.Print(uri)
			ctx.URI().QueryArgs().VisitAll(func(key, value []byte) {
			})
		}
		// -------------------------------------------------------
		// 注意对比一下, 下面的代码段, 与 web client  中几乎一样
		// -------------------------------------------------------
		{
			// 取出 web client 请求中的 HTTP header
			{
				ctx.Request.Header.VisitAll(func(key, value []byte) {
					// l.Info("requestHeader", zap.String("key", gotils.B2S(key)), zap.String("value", gotils.B2S(value)))
				})

			}
			// 取出 web client 请求中的 HTTP payload
			{
			}
		}
		switch {
		// 如果访问的 URI 路由是 /uri 开头 , 则进行下面这个响应
		case len(uriPath) > 1:
			{

				// -------------------------------------------------------
				// 处理逻辑开始
				// -------------------------------------------------------

				// payload 是 []byte , 是 web response 返回的 HTTP payload
				var payload = bytes.NewBuffer([]byte("Hello, "))

				// 这是从 web client 取数据
				var who = ctx.QueryArgs().PeekBytes([]byte("who"))

				if len(who) > 0 {
					payload.Write(who)
				} else {
					payload.Write([]byte(" 中国 "))
				}

				// -------------------------------------------------------
				// 处理 HTTP 响应数据
				// -------------------------------------------------------
				// HTTP header 构造
				ctx.Response.Header.SetStatusCode(200)
				ctx.Response.Header.SetConnectionClose() // 关闭本次连接, 这就是短连接 HTTP
				ctx.Response.Header.SetBytesKV([]byte("Content-Type"), []byte("text/plain; charset=utf8"))
				ctx.Response.Header.SetBytesKV([]byte("TransactionID"), []byte(tid))
				// HTTP payload 设置
				// 这里 HTTP payload 是 []byte
				ctx.Response.SetBody(payload.Bytes())
			}

		// 访问路踊不是 /uri 的其他响应
		default:
			{

				// -------------------------------------------------------
				// 处理逻辑开始
				// -------------------------------------------------------

				// payload 是 []byte , 是 web response 返回的 HTTP payload
				var payload = bytes.NewBuffer([]byte("Hello, "))

				// 这是从 web client 取数据
				var who = ctx.QueryArgs().PeekBytes([]byte("who"))

				if len(who) > 0 {
					payload.Write(who)
				} else {
					payload.Write([]byte(" 中国 "))
				}

				// -------------------------------------------------------
				// 处理 HTTP 响应数据
				// -------------------------------------------------------
				// HTTP header 构造
				ctx.Response.Header.SetStatusCode(200)
				ctx.Response.Header.SetConnectionClose() // 关闭本次连接, 这就是短连接 HTTP
				ctx.Response.Header.SetBytesKV([]byte("Content-Type"), []byte("text/plain; charset=utf8"))
				ctx.Response.Header.SetBytesKV([]byte("TransactionID"), []byte(tid))
				// HTTP payload 设置
				// 这里 HTTP payload 是 []byte
				ctx.Response.SetBody(payload.Bytes())
			}
		}

		return

	}
	// -------------------------------------------------------
	// 创建 fasthttp 服务器
	// -------------------------------------------------------
	// Create custom server.
	s := &fasthttp.Server{
		Handler: requestHandler,       // 注意这里
		Name:    "hello-world server", // 服务器名称
	}
	// -------------------------------------------------------
	// 运行服务端程序
	// -------------------------------------------------------

	if err := s.ListenAndServe(address); err != nil {
		log.Fatal("error in ListenAndServe")
	}
}

// server.go
package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 处理连接
	fmt.Println("Accepted connection from", conn.RemoteAddr().String())

	// 发送欢迎消息给客户端
	conn.Write([]byte("Welcome to the server!\n"))
}

func main() {
	// 监听端口
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on :8080")

	for {
		// 等待连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			break
		}

		// 启动一个新的goroutine处理连接
		go handleConnection(conn)
	}
}

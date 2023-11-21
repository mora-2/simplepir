// client.go
package main

import (
	"fmt"
	"net"
)

func main() {
	// 连接服务器
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	// 读取服务器的欢迎消息
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	fmt.Println(len(buffer))
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	fmt.Print(string(buffer[:n]))
}

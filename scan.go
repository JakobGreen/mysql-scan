package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

// Get null terminated string
func read_cstr(buf []byte) string {
	pos := bytes.IndexByte(buf, 0)
	if pos == -1 {
		return string(buf[:])
	}

	return string(buf[:pos])
}

func main() {
	dialer := net.Dialer{}

	ctx, cancel := context.WithCancel(context.Background())
	conn, err := dialer.DialContext(ctx, "tcp", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	defer cancel()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)

		fmt.Println("Received: ", n)
		if err != nil {
			panic(err)
		}

		if n > 0 {
			break
		}
		time.Sleep(time.Second)
	}

	pktLen := int(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16)
	seq := buf[3]

	fmt.Println("Packet Length: ", pktLen)
	fmt.Println("Seq: ", seq)
	fmt.Println("MySQL Protocol: ", buf[4])
	fmt.Println("MySQL Version: ", read_cstr(buf[5:]))
}

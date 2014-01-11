package main

import (
	"fmt"
	"github.com/tvdburgt/passman/cache"
	"net"
	"net/rpc"
)

// -timeout=10m (update after each passman call)
func main() {
	c := make(cache.Cache)
	rpc.Register(c)
	l, err := net.Listen(cache.Network, cache.Address)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// TODO: check race conditions
		go rpc.ServeConn(conn)
	}
}

func handle(conn net.Conn) {
	msg := make([]byte, 1024);
	_, err := conn.Read(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println(msg)
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

func server(network, address string) {
	ln, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("no listen: ", err)
		return
	}

	defer ln.Close()

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("no accept", err)
	}

	var wg sync.WaitGroup
	var result string
	//var settings []byte

	for {	
		message, err := bufio.NewReader(conn).ReadString('|')
		if err != nil {
			fmt.Println("no accept: ", err)
			conn.Close()
			continue
		}

		fmt.Println("message >>>> ", message)

		str := strings.TrimRight(message, "|")

		fmt.Println("____________--------_________")
		fmt.Println(settings)
		fmt.Println("------------________---------")

		wg.Add(1)
		go func() {
			result, _ = processingClientRequest(str, &wg)
		}()
		wg.Wait()

		conn.Write([]byte(result))

		//time.Sleep(10 * time.Millisecond)
	}
}

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

	//var wg1 sync.WaitGroup
	var result string
	//var settings []byte

	var mutex, mutex2 sync.Mutex

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

/* 		wg.Add(1)
		go func() {
			result, _ = processingClientRequest(str, &wg)
		}()
		wg.Wait() */

		fmt.Println("client request >>>> ", str)

		str1 := strings.TrimRight(str, "|")
		pl := strings.Split(str1, " ")

		switch pl[0] {
		case "set_flow", "get_flow" :
			//wg1.Add(1)
			go func() {
				mutex2.Lock()
				result = processingClientRequest(str/* , &wg1 */)
				mutex2.Unlock()
			}()
			//wg1.Wait()

		case "get_raw_data", "get_ga", "set_ga", "get_ppm":
			go func() {
				mutex.Lock()
				result = processingClientRequest(str/* , &wg2 */)
				mutex.Unlock()
			}()
		}

		//result, _ = processingClientRequest(str/* , &wg */)

		conn.Write([]byte(result))
	}
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

func server(network, address string) {
	for {
		ln, err := net.Listen(network, address)
		if err != nil {
			fmt.Println("no listen: ", err)
			continue
		}
	
		defer ln.Close()
	
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("no accept", err)
			continue
		}
	
		var wg1, wg2 sync.WaitGroup
		var result string
	
		var mutex/* , mutex2 */ sync.Mutex
	
		for {
			message, err := bufio.NewReader(conn).ReadString('|')
			if err != nil {
				fmt.Println("no accept: ", err)
				conn.Close()
				break
			}
	
			fmt.Println("message >>>> ", message)
	
			str := strings.TrimRight(message, "|")
	
			fmt.Println("____________--------_________")
			fmt.Println(settings)
			fmt.Println("------------________---------")
	
			fmt.Println("client request >>>> ", str)
	
			str1 := strings.TrimRight(str, "|")
			pl := strings.Split(str1, " ")
	
			switch pl[0] {
			case "set_flow", "get_flow" :			
				go func() {
					mutex.Lock()
					result = processingClientRequest(str, &wg1)
					mutex.Unlock()
				}()
	
			case "get_raw_data", "get_ga", "set_ga", "get_ppm":
				if !gaFlag {
					go func() {
						conn.Write([]byte("busy |"))
						result = processingClientRequest(str, &wg2) + "|"
						conn.Write([]byte("free |"))
					}()
				} else {
					fmt.Println(">>>>........................ thread is busy")
				}
			}
	
			conn.Write([]byte(result + "|"))
			result = ""
		}
	}
}

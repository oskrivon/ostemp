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

	var wg1, wg2 sync.WaitGroup
	//var result string

	var mutex/* , mutex2  */sync.Mutex

	for {
		var result string
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
				mutex.Lock()
				result = processingClientRequest(str, &wg1)
				//conn.Write([]byte(result))
				mutex.Unlock()
			}()
			//wg1.Wait()

		case "get_raw_data", "get_ga", "set_ga", "get_ppm":
			//wg2.Add(1)
			if !gaFlag {
				go func() {
					result = processingClientRequest(str, &wg2)
					//conn.Write([]byte(result))
				}()
			} else {
				fmt.Println(">>>> thread is busy")
			}
			
			//wg2.Wait()
		}

		//result, _ = processingClientRequest(str/* , &wg */)

		conn.Write([]byte(result))
		result = ""
	}
}

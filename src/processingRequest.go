package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func processingClientRequest(request string/* , wg *sync.WaitGroup */)(string, error) {
	fmt.Println("client request >>>> ", request)

	str := strings.TrimRight(request, "|")
	pl := strings.Split(str, " ")

	//defer wg.Done()

	var response []byte
	var result string
	var err error

	var wg1, wg2 sync.WaitGroup

	//c := make(chan []byte)

	switch pl[0] {
	case "set_flow":
		wg1.Add(1)
		go func() {
			k := 1.41 / 125
			setPoint, _ := strconv.ParseFloat(pl[1], 32)
			value := setPoint / k
	
			fmt.Println(uint16(value))
			
			response, err = currentSystem.flowController[0].sendCommand(commands["set flow"], 0x2, uint16(value), "set")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(response)
	
			setPoint, _ = strconv.ParseFloat(pl[2], 32)
			value = setPoint / k
	
			fmt.Println(uint16(value))
			response, err = currentSystem.flowController[0].sendCommand(commands["set flow"], 0x3, uint16(value), "set")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(response)
		}()
		wg1.Wait()
		
	case "get_flow":
		wg1.Add(1)
		go func() {
			var flow1, flow2 string
			result = "get_flow"
			response, err = currentSystem.flowController[0].sendCommand(commands["get flow"], 0x2, 1, "get")
			if err != nil {
				result = "FC error"
				fmt.Println("response from fc >>>> ", err)
				flow1 = "0"
			} else {
				flow1, _ = parseResponse(response, "get flow")
			}
	
			response, err = currentSystem.flowController[0].sendCommand(commands["get flow"], 0x3, 1, "get")
			if err != nil {
				result = "FC error"
				fmt.Println("response from fc >>>> ", err)
				flow2 = "0"
			} else {
				flow2, _ = parseResponse(response, "get flow")
			}
	
			result = result + " " + flow1 + " " + flow2
		}()
		wg1.Wait()

	case "get_raw_data":
		wg2.Add(1)
		go func() {
			response =  currentSystem.gasAnalyzer[0].sendCommand(commands["get raw sensor data"], 0)

			result, err = parseResponse(response, "get raw sensor data")
	
			time.Sleep(1 * time.Second)
		}()
		wg2.Wait()

	case "get_ga":
		wg2.Add(1)
		go func() {
			response =  currentSystem.gasAnalyzer[0].sendCommand(commands["get ga options"], 0)

			settings = response
	
			result, _ = parseResponse(response, "get ga options")
	
			time.Sleep(1 * time.Second)
		}()
		wg2.Wait()

	case "set_ga":
		//f := func () []byte{return settings}
		wg2.Add(1)
		go func() {
			response = settings

			fmt.Println("response >>>> ", response)
	
			instruction := parsingDataFromClient(pl[1:])
			instruction = append([]byte{0x4f, 0x07, 0x8d, 0x7b}, instruction...)
			withCRC, _ := SignBytesLE(instruction)
	
			newCommand := command{
				instruction: withCRC,
				responseLenght: 0,
			}
	
			_ =  currentSystem.gasAnalyzer[0].sendCommand(newCommand, 0)
	
			result = "ok"
	
			time.Sleep(1 * time.Second)
		}()
		wg2.Wait()

	case "get_ppm":
		wg2.Add(1)
		go func() {
			response = currentSystem.gasAnalyzer[0].sendCommand(commands["get ppm"], 0)

			result, err = parseResponse(response, "get ppm")
	
			fmt.Println(result)
			time.Sleep(1 * time.Second)
		}()
		wg2.Wait()
	}

	return result, err
}
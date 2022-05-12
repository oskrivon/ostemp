package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func processingClientRequest(request string, wg *sync.WaitGroup) string {
	fmt.Println("client request >>>> ", request)
	var (
		response []byte
		result string
		err error
	)

	str := strings.TrimRight(request, "|")
	pl := strings.Split(str, " ")
	
	switch pl[0] {
	case "set_flow":
		k := 1.41 / 125
		setPoint, _ := strconv.ParseFloat(pl[1], 32)
		value := setPoint / k

		fmt.Println(uint16(value))
		
		response, err = currentSystem.flowController.sendCommand(commands["set flow"], currentSystem.fcId1, uint16(value), "set")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response)

		setPoint, _ = strconv.ParseFloat(pl[2], 32)
		value = setPoint / k

		fmt.Println(uint16(value))
		response, err = currentSystem.flowController.sendCommand(commands["set flow"], currentSystem.fcId2, uint16(value), "set")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response)
		
	case "get_flow":
		var flow1, flow2 string
		result = "get_flow"
		response, err = currentSystem.flowController.sendCommand(commands["get flow"], currentSystem.fcId1, 1, "get")
		if err != nil {
			result = "FC error"
			fmt.Println("response from fc >>>> ", err)
			flow1 = "0"
		} else {
			flow1, _ = parseResponse(response, "get flow")
		}

		response, err = currentSystem.flowController.sendCommand(commands["get flow"], currentSystem.fcId2, 1, "get")
		if err != nil {
			result = "FC error"
			fmt.Println("response from fc >>>> ", err)
			flow2 = "0"
		} else {
			flow2, _ = parseResponse(response, "get flow")
		}

		result = result + " " + flow1 + " " + flow2

	case "get_raw_data":
		gaFlag = true
		response =  currentSystem.gasAnalyzer.sendCommand(commands["get raw sensor data"], 0)

		result, err = parseResponse(response, "get raw sensor data")
		if err!= nil {
			fmt.Print("some error")
		}

		time.Sleep(3 * time.Second)
		gaFlag = false

	case "get_ga":
		gaFlag = true
		response =  currentSystem.gasAnalyzer.sendCommand(commands["get ga options"], 0)

		settings = response

		result, _ = parseResponse(response, "get ga options")

		time.Sleep(3 * time.Second)
		gaFlag = false

	case "set_ga":
		gaFlag = true
		response = settings

		fmt.Println("response >>>> ", response)

		instruction := parsingDataFromClient(pl[1:])
		instruction = append([]byte{0x4f, 0x07, 0x8d, 0x7b}, instruction...)
		withCRC, _ := SignBytesLE(instruction)

		newCommand := command{
			instruction: withCRC,
			responseLenght: 0,
		}

		_ =  currentSystem.gasAnalyzer.sendCommand(newCommand, 0)

		result = "ok"

		time.Sleep(3 * time.Second)
		gaFlag = false

	case "get_ppm":
		gaFlag = true
		response = currentSystem.gasAnalyzer.sendCommand(commands["get ppm"], 0)

		result, err = parseResponse(response, "get ppm")
		if err!= nil {
			fmt.Print("some error")
		}

		fmt.Println(result)
		time.Sleep(3 * time.Second)
		gaFlag = false

	case "clean_air":
		k := 1.41 / 125
		value1 := 1 / k
		value2 := 1 / k
		
		_, err = currentSystem.flowController.sendCommand(commands["set flow"], currentSystem.fcId1, uint16(value1), "set")
		if err != nil {
			fmt.Println(err)
		}

		_, err = currentSystem.flowController.sendCommand(commands["set flow"], currentSystem.fcId1, uint16(value2), "set")
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(result)
	return result
}
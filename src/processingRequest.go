package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func processingClientRequest(request string, wg *sync.WaitGroup)(string, error) {
	fmt.Println("client request >>>> ", request)

	str := strings.TrimRight(request, "|")
	pl := strings.Split(str, " ")

	defer wg.Done()

	var response []byte
	var result string
	var err error

	//c := make(chan []byte)

	switch pl[0] {
	case "set_flow":
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
		
	case "get_flow":
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

	case "get_raw_data":
		response =  currentSystem.gasAnalyzer[0].sendCommand(commands["get raw sensor data"], 0)

		result, err = parseResponse(response, "get raw sensor data")

		time.Sleep(1 * time.Second)

	case "get_ga":
		response =  currentSystem.gasAnalyzer[0].sendCommand(commands["get ga options"], 0)

		settings = response

		result, _ = parseResponse(response, "get ga options")

		time.Sleep(1 * time.Second)

	case "set_ga":
		//f := func () []byte{return settings}
		response = settings

		fmt.Println("response >>>> ", response)

/* 		var newSettings, firstPart, secondPart, modifiedPart []byte
		var beginPlace int

		for i := 0; i < 4; i++ {
			beginPlace = i * 33 + 9 + 29 + 3

			firstPart = response[3:beginPlace]
			secondPart = response[beginPlace + 4:]

			buf := new(bytes.Buffer)
			f, _ := strconv.ParseFloat(pl[i + 1], 32)
			fmt.Println("strconv.ParseFloat >>>>", f)
			
			err = binary.Write(buf, binary.LittleEndian, float32(f))
			if err != nil {
				fmt.Println("binary.Write failed:", err)
			}

			modifiedPart = buf.Bytes()

			newSettings = nil

			newSettings = append(newSettings, firstPart...)
			newSettings = append(newSettings, modifiedPart...)
			newSettings = append(newSettings, secondPart...)
			newSettings = newSettings[:141]

			settings = nil

			newSettings = append([]byte{0x4f, 0x07, 0x8d}, newSettings...)
			newSettings, _ := SignBytesLE(newSettings)

			response = newSettings
		}

		for i := 0; i < 4; i++ {
			beginPlace = i * 33 + 9 + 23 + 3

			firstPart = response[3:beginPlace]
			secondPart = response[beginPlace + 4:]

			buf := new(bytes.Buffer)
			f, _ := strconv.ParseFloat(pl[i + 5], 32)
			f = f / 1000 / 1000 / 1000
			fmt.Println("strconv.ParseFloat >>>>", f)
			
			err = binary.Write(buf, binary.LittleEndian, float32(f))
			if err != nil {
				fmt.Println("binary.Write failed:", err)
			}

			modifiedPart = buf.Bytes()

			newSettings = nil

			newSettings = append(newSettings, firstPart...)
			newSettings = append(newSettings, modifiedPart...)
			newSettings = append(newSettings, secondPart...)
			newSettings = newSettings[:141]

			settings = nil

			newSettings = append([]byte{0x4f, 0x07, 0x8d}, newSettings...)
			newSettings, _ := SignBytesLE(newSettings)

			response = newSettings
		} */

		newVersion := parsingDataFromClient(pl[1:])
		newVersion = append([]byte{0x4f, 0x07, 0x8d, 0x7b}, newVersion...)
		xxx, _ := SignBytesLE(newVersion)

		//newSettings, _ = SignBytesLE(newSettings)

		newCommand := command{
			instruction: xxx,
			responseLenght: 0,
		}

		fmt.Println("old version:::::::::::::", newCommand)
		fmt.Println(":::::::::::::")
		fmt.Println("new version:::::::::::::", xxx)

		_ =  currentSystem.gasAnalyzer[0].sendCommand(newCommand, 0)

		result = "ok"

		time.Sleep(1 * time.Second)

	case "get_ppm":
		response = currentSystem.gasAnalyzer[0].sendCommand(commands["get ppm"], 0)

		result, err = parseResponse(response, "get ppm")

		fmt.Println(result)
		time.Sleep(1 * time.Second)
	}

	return result, err
}
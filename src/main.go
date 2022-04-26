package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/modbus"
	"github.com/jacobsa/go-serial/serial"
)

var (
	logName string
	currentSystem systemComfig
	commands map[string]command
	settings []byte
	gaFlag bool  = false
)

//var wg sync.WaitGroup

type Material struct {
	Quantity int `yaml:"quantity"`
	TypeID   int `yaml:"typeID"`
}

func main() {
	var err error

	currentSystem = createSystem()

	logName, err = createLog()
	if err != nil {
		fmt.Println(err)
	}

	commands = safeCommands()


	go func ()  {
		listener, err := net.Listen("tcp", ":8082")
		if err != nil {
			fmt.Println(">>>>> new server is failed!!!!", err)
		}

		fmt.Println(">>>>> server is OK!")

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(">>>>>> new connection is failed!!!", err)
				return
			}

			input := make([]byte, 1024)

			n, err := conn.Read(input)
			if err != nil || n == 0 {
				fmt.Println(">>>> read error", err)
				break
			}

			fmt.Println(">>>>> input byte: ", input)
			fmt.Println(">>>>> input byte: ", string(input))

		}
	}()

	server("tcp", ":8081")
}

func (ga *gasAnalyzer) sendCommand(command command, id byte /* , c chan []byte */) []byte {
	port, err := serial.Open(currentSystem.gaConfig)
	if err != nil {
		fmt.Println(err)
		//c <- nil
		return nil
	}

	_, err = port.Write(command.instruction)
	fmt.Println("instruction >>>> ", command.instruction)
	if err != nil {
		fmt.Println(err)
		//c <- nil
		return nil
	}

	time.Sleep(100 * time.Millisecond)

	var sensorRerponse []byte

	for {
		var buf = make([]byte, 256)

		if len(sensorRerponse) >= command.responseLenght {
			//port.Close()
			break
		}

		n, err := port.Read(buf)
		if err != nil {
			//c <- nil
			fmt.Println(err)
		}
		//fmt.Println(sensorRerponse)

		sensorRerponse = append(sensorRerponse, buf[:n]...)
	}

	//defer close(c)
	defer port.Close()

	//fmt.Println(sensorRerponse)
	//returnString, _ := parseResponse(sensorRerponse, command.name)

	//c <- sensorRerponse
	err = VerifyTrailingBytesLE(sensorRerponse)
	if err != nil {
		return nil
	}

	return sensorRerponse
}

func (fc *flowController) sendCommand(command command, id byte, value uint16, tFlag string) (response []byte, err error) {
	handler := modbus.NewRTUClientHandler(currentSystem.fcConfig.PortName)
	handler.BaudRate = int(currentSystem.fcConfig.BaudRate)
	handler.DataBits = int(currentSystem.fcConfig.DataBits)
	handler.Parity = "E"
	handler.StopBits = int(currentSystem.fcConfig.StopBits)
	handler.SlaveId = id
	handler.Timeout = 1 * time.Second

	err = handler.Connect()
	if err != nil {
		return nil, err
	}

	defer handler.Close()

	client := modbus.NewClient(handler)
	time.Sleep(10 * time.Millisecond)

	var result []byte

	switch tFlag {
	case "set":
		bytes, err := client.WriteMultipleRegisters(command.adress, 1, []byte{byte(value), 0})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("set: ", bytes)
		result = nil
	case "get":
		result, err = client.ReadHoldingRegisters(command.adress, command.quantity)
		if err != nil {
			fmt.Println("error get FC >>>>", err)
		}
		fmt.Println("get: ", result)
	}

	return result, err
}

func clientMessageProcessing(message string, WG *sync.WaitGroup) (string, error) {
	fmt.Println("message from client: ", message)
	str := strings.TrimRight(message, "|")
	pl := strings.Split(str, " ")

	defer WG.Done()

	var result []byte
	var err error
	var wg sync.WaitGroup

	c := make(chan []byte)

	switch pl[0] {
	case "set_flow":
		/* targetFlow, _ := strconv.ParseInt(pl[2], 10, 2)
		targetConcentration := strconv.ParseInt(pl[3], 10, 2)

		// flow = A*x1 + B*x2
		// conc = x1/x2 -> x1 = conc*x2 -> flow = A*conc*x2 + B*x2 = x2*(A*conc + B)
		// x2 = flow/(A*conc+B), x1 = conc*flow/(A*conc + B) | A = B

		flow1 =

		result, err = system.flowController[0].sendCommand(system, commands[4], 1, uint16(v), "set")
		if err != nil {
			fmt.Println("error reseive")
		} */
	case "get_flow":
		result, err = currentSystem.flowController.sendCommand(commands["get flow"], 1, 1, "get")
		if err != nil {
			fmt.Println("request submission error")
		}
	case "set_ga":
		fmt.Println("ga settings", pl)

		//<-c
		wg.Add(1)
		go currentSystem.gasAnalyzer.sendCommand(commands["set ga options"], 0)
		result = <-c
		wg.Wait()

		//result = <-c

		if err != nil {
			return "error", err
		}

		xxx, _ := parseResponse(result, "")
		splitedResult := strings.Split(xxx, " ")
		fmt.Println("splited request: ", splitedResult)

		cellNumber, _ := strconv.ParseUint(pl[1], 10, 8)
		splitedResult[cellNumber*11+10] = pl[2]

		/* for i := 1; i < 5; i++ {
			splitedResult[i * 10] = "0.0"
		} */

		//fmt.Println("after setting: ", splitedResult, "lenght: ", len(splitedResult))

		var comandIntruction []byte

		for i, v := range splitedResult {
			switch i {
			case 0, 3, 4, 11, 14, 15, 22, 25, 26, 33, 36, 37:
				bufWrite := new(bytes.Buffer)
				//fmt.Println("v: ", v)
				r, _ := strconv.ParseUint(v, 10, 8)
				//fmt.Println("convert: ", r)
				_ = binary.Write(bufWrite, binary.LittleEndian, uint8(r))
				comandIntruction = append(comandIntruction, bufWrite.Bytes()...)
				//fmt.Println("bufWrite: ", bufWrite.Bytes())
			case 1, 2, 7, 8, 10, 12, 13, 18, 19, 21, 23, 24, 29, 30, 32, 34, 35, 40, 41:
				bufWrite := new(bytes.Buffer)
				r, _ := strconv.ParseFloat(v, 32)
				if (i != 8) && (i != 19) && (i != 30) && (i != 41) {
					_ = binary.Write(bufWrite, binary.LittleEndian, float32(r))
					comandIntruction = append(comandIntruction, bufWrite.Bytes()...)
				} else {
					_ = binary.Write(bufWrite, binary.LittleEndian, float32(r/1000/1000/1000))
					comandIntruction = append(comandIntruction, bufWrite.Bytes()...)
				}
				//fmt.Println("bufWrite: ", bufWrite.Bytes())

			case 5, 6, 16, 17, 27, 28, 38, 39:
				bufWrite := new(bytes.Buffer)
				r, _ := strconv.ParseUint(v, 10, 32)
				_ = binary.Write(bufWrite, binary.LittleEndian, uint32(r))
				comandIntruction = append(comandIntruction, bufWrite.Bytes()...)
			case 9, 20, 31, 43:
				bufWrite := new(bytes.Buffer)
				r, _ := strconv.ParseUint(v, 10, 16)
				_ = binary.Write(bufWrite, binary.LittleEndian, uint16(r))
				comandIntruction = append(comandIntruction, bufWrite.Bytes()...)
			}
			//fmt.Println(comandIntruction)
		}

		//fmt.Println(len(comandIntruction))
		lenBuf := make([]byte, 8)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(comandIntruction)))
		//fmt.Println("xxxx: ", lenBuf)
		//lengthBuff, _ := strconv.ParseUint(v, 10, 8)

		comandIntruction = append([]byte{0x4f, 0x07}, append(lenBuf[0:7], comandIntruction...)...)
		//fmt.Println("with suffix", comandIntruction)

		var commandToSettings command
		//commandToSettings.instruction = append(commandToSettings.instruction, lenBuf...)
		commandToSettings.instruction, _ = SignBytesLE(comandIntruction)
		//fmt.Println("with suffix", commandToSettings.instruction)
		//fmt.Println("length: ", len(commandToSettings.instruction))
		commandToSettings.responseLenght = 1

		/* result, err = system.gasAnalyzer[0].sendCommand(system, commandToSettings,0)
		if err != nil {
			return "error", err
		}
		time.Sleep(1000 * time.Millisecond)*/
		fmt.Println("set_ga command: ", result)

		//go newSystem.gasAnalyzer[0].sendCommand(commands[9], 0, c)

		if err != nil {
			return "error", err
		}
	case "get_raw_data":
		fmt.Println("RAW DATA")

		wg.Add(1)
		go currentSystem.gasAnalyzer.sendCommand(commands["get raw sensor data"], 0)
		result = <-c
		wg.Wait()

		if err != nil {
			return "error", err
		}
		fmt.Println("result", result)
	}

	return parseResponse(result, "")
}

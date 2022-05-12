package main

import (
	"fmt"
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
			break
		}

		n, err := port.Read(buf)
		if err != nil {
			fmt.Println(err)
		}

		sensorRerponse = append(sensorRerponse, buf[:n]...)
	}

	defer port.Close()

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
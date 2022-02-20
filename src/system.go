package main

import "github.com/jacobsa/go-serial/serial"

type gasAnalyzer struct {
}

type flowController struct {
	address byte
}

type system struct {
	port           string
	gasAnalyzer    []gasAnalyzer
	flowController []flowController
	//valve []Valve

	gaOptions serial.OpenOptions
}

type command struct {
	name           string
	instruction    []byte
	responseLenght int
	adress         uint16
	quantity       uint16
}

type GBObject interface {
	sendCommand(system, command, byte) (string, error)
}

func createSystem() system {
	return system{
		//port: "COM5",
		port: "/dev/ttyUSB1",
		gasAnalyzer: []gasAnalyzer{
			{},
		},
		flowController: []flowController{
			{
				address: 0,
			},
		},
		gaOptions: serial.OpenOptions{
			PortName:        "/dev/ttyUSB1",
			BaudRate:        9600,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 5,
			ParityMode:      serial.PARITY_NONE,
			InterCharacterTimeout: 10000,
		},
	}
}

type DataSensor struct {
	configID      uint8
	serial        string
	gasType       uint8
	v_ref         float32
	v_ref_comp    float32
	afe_bias      uint8
	afe_r_gain    uint8
	rangeMin      uint32
	rangeMax      uint32
	resolution    float32
	amp2ppm       float32
	responseTime  uint16
	baseLineShift float32
}

type ppmGas struct {
	id uint8
	ppm uint32
}

type ppmData struct {
	temperature int16
	humidity uint16
	pressure uint32
	gases []ppmGas
}
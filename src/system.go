package main

import (
	"fmt"
	"io/ioutil"

	"github.com/jacobsa/go-serial/serial"
	"gopkg.in/yaml.v3"
)

/* serial.OpenOptions{
	PortName:        "/dev/ttyUSB1",
	BaudRate:        9600,
	DataBits:        8,
	StopBits:        1,
	MinimumReadSize: 5,
	ParityMode:      serial.PARITY_NONE,
	InterCharacterTimeout: 10000,
}, */

type System struct {
	Server struct {
		Port int `yaml:"port"` 
	} `yaml:"server"`

	GasAnalyser struct {
		Port     string `yaml:"port"`
		BaudRate uint `yaml:"baudRate"`
		DataBits uint `yaml:"dataBits"`
		StopBits uint `yaml:"stopBits"`
		MinimumReadSize uint `yaml:"minimumReadSize"`
		ParityMode uint `yaml:"parityMode"`
		InterCharacterTimeout uint `yaml:"interCharacterTimeout"`
	} `yaml:"gasAnalyser"`

	FlowController struct {
		Port     string `yaml:"port"`
		BaudRate uint `yaml:"baudRate"`
		DataBits uint `yaml:"dataBits"`
		StopBits uint `yaml:"stopBits"`
		MinimumReadSize uint `yaml:"minimumReadSize"`
		ParityMode uint `yaml:"parityMode"`
		InterCharacterTimeout uint `yaml:"interCharacterTimeout"`
	} `yaml:"gasAnalyser"`
}

type gasAnalyzer struct {

}

type flowController struct {

}

type systemComfig struct {
	flowController []flowController
	gasAnalyzer []gasAnalyzer

	gaConfig serial.OpenOptions
	fcConfig serial.OpenOptions
}

type command struct {
	name           string
	instruction    []byte
	responseLenght int
	adress         uint16
	quantity       uint16
}

type GBObject interface {
	sendCommand(System, command, byte) (string, error)
}

func createSystem() systemComfig {
	var system System
	var systemComfig systemComfig

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil{
		fmt.Println("error with file reading", err)
	}

	err = yaml.Unmarshal(yamlFile, &system)
	if err != nil{
		fmt.Println("error with config unmarshalling", err)
	}

	systemComfig.gaConfig.PortName = system.GasAnalyser.Port
	systemComfig.gaConfig.BaudRate = system.GasAnalyser.BaudRate
	systemComfig.gaConfig.DataBits = system.GasAnalyser.DataBits
	systemComfig.gaConfig.StopBits = system.GasAnalyser.StopBits
	systemComfig.gaConfig.MinimumReadSize = system.GasAnalyser.MinimumReadSize
	systemComfig.gaConfig.ParityMode = serial.ParityMode(system.GasAnalyser.ParityMode)
	systemComfig.gaConfig.InterCharacterTimeout = system.GasAnalyser.InterCharacterTimeout

	systemComfig.fcConfig.PortName = system.FlowController.Port
	systemComfig.fcConfig.BaudRate = system.FlowController.BaudRate
	systemComfig.fcConfig.DataBits = system.FlowController.DataBits
	systemComfig.fcConfig.StopBits = system.FlowController.StopBits
	systemComfig.fcConfig.MinimumReadSize = system.FlowController.MinimumReadSize
	systemComfig.fcConfig.ParityMode = serial.ParityMode(system.FlowController.ParityMode)
	systemComfig.fcConfig.InterCharacterTimeout = system.FlowController.InterCharacterTimeout
	
	return systemComfig
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
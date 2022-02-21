package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

func parseResponse(in []byte, marker string) (response string, err error) {
	var resultString string

	switch marker {
	case "get raw sensor data":
		inData := in[11:]

		resultString = "raw_data "
		var log []string

		for i := 0; i < 4; i++ {
			begin := 1 + i*9
			end := begin + 4
			buf := bytes.NewReader(inData[begin:end])

			var r float32
			err = binary.Read(buf, binary.LittleEndian, &r)
			if err != nil {
				return "error", err
			}

			log = append(log, strconv.FormatFloat(float64(r), 'f', -1, 64))
			resultString = resultString + strconv.FormatFloat(float64(r) * 1000 * 1000 * 1000, 'f', -1, 64) + " "
		}
		record(log)
	case "get ga options":
		var ds DataSensor
		var responseLenght = 33

		inData := in[3:]

		resultString = "ga_options "

		buf := bytes.NewReader(inData[0:])
		_ = binary.Read(buf, binary.LittleEndian, &ds.configID)

		ds.serial = hex.EncodeToString(inData[1:9])

		resultString = resultString +
			strconv.FormatFloat(float64(ds.configID), 'f', -1, 64) + " " +
			ds.serial + " "

		for i := 0; i < 4; i++ {
			buf := bytes.NewReader(inData[9+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.gasType)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+1+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.v_ref)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+5+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.v_ref_comp)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+9+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.afe_bias)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+10+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.afe_r_gain)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+11+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.rangeMin)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+15+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.rangeMax)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+19+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.resolution)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+23+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.amp2ppm)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+27+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.responseTime)
			if err != nil {
				fmt.Println(err)
			}

			buf = bytes.NewReader(inData[9+29+i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ds.baseLineShift)
			if err != nil {
				fmt.Println(err)
			}

			resultString = resultString + strconv.FormatFloat(float64(ds.gasType), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.v_ref), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.v_ref_comp), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.afe_bias), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.afe_r_gain), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.rangeMin), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.rangeMax), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.resolution), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.amp2ppm)*1000*1000*1000, 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.responseTime), 'f', -1, 64) + " " +
				strconv.FormatFloat(float64(ds.baseLineShift), 'f', -1, 64) + " "
			//fmt.Println("ds:", resultString)
		}

	case "set ga options":


	case "get flow":
		resultString = "get_flow "
		
		k := 1.41 / 125
		r := uint8(in[0])
		//l := uint8(in[1])

		f := k * float64(r)
		//r := uint8(binary.LittleEndian.Uint16.Uint16(in[0]))
		//err = binary.Read(buf, binary.LittleEndian, &r)
		
		//fmt.Println("int: ", f)
/* 		if err != nil {
			fmt.Println("___flow___", err, "error!")
			return "error", err
		} */
		//fmt.Println("___flow___", f)

		resultString = strconv.FormatFloat(f, 'f', -1, 64)
	case "get ppm":
		resultString = "get_ppm "
		var responseLenght = 5

		inData := in[3:]

		var ppmData ppmData

		buf := bytes.NewReader(inData[0:])
		err = binary.Read(buf, binary.LittleEndian, &ppmData.temperature)
		if err != nil {
			fmt.Println(err)
		}

		resultString = resultString + strconv.FormatFloat(float64(ppmData.temperature)/100, 'f', -1, 64) + " "

		buf = bytes.NewReader(inData[2:])
		err = binary.Read(buf, binary.LittleEndian, &ppmData.humidity)
		if err != nil {
			fmt.Println(err)
		}

		resultString = resultString + strconv.FormatFloat(float64(ppmData.humidity)/100, 'f', -1, 64) + " "

		buf = bytes.NewReader(inData[4:])
		err = binary.Read(buf, binary.LittleEndian, &ppmData.pressure)
		if err != nil {
			fmt.Println(err)
		}

		resultString = resultString + strconv.FormatFloat(float64(ppmData.pressure)/100, 'f', -1, 64) + " "

		for i := 0; i < 4; i++ {
			var ppmGas ppmGas

			buf = bytes.NewReader(inData[8 + i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ppmGas.id)
			if err != nil {
				fmt.Println(err)
			}

			resultString = resultString + strconv.FormatFloat(float64(ppmGas.id), 'f', -1, 64) + " "

			buf = bytes.NewReader(inData[8 + 1 + i*responseLenght:])
			err = binary.Read(buf, binary.LittleEndian, &ppmGas.ppm)
			if err != nil {
				fmt.Println(err)
			}

			resultString = resultString + strconv.FormatFloat(float64(ppmGas.ppm)/100, 'f', -1, 64) + " "

			ppmData.gases = append(ppmData.gases, ppmGas)
		}
	}

	fmt.Println("result: ", resultString)
	return resultString, nil
}

func parsingDataFromClient(data []string) []byte {
	var (
		err error
		result []byte
	)

	result = append(result, []byte(data[0])...)
	data = data[1:]

	for i := 0; i < 4; i++ {
		for j := 0; j < 11; j++ {
			switch j {
			case 0, 3, 4:
				var x uint64
				x, err = strconv.ParseUint(data[i * 11 + j], 10, 8)
				if err != nil {
					x = 0
					fmt.Println("error parsing client data", err)
				}
				result = append(result, toByte(uint8(x))...)

			case 1, 2, 7, 10:
				var x float64
				x, err = strconv.ParseFloat(data[i * 11 + j], 32)
				if err != nil {
					x = 0
					fmt.Println("error parsing client data", err)
				}
				result = append(result, toByte(float32(x))...)

			case 8:
				var x float64
				x, err = strconv.ParseFloat(data[i * 11 + j], 32)
				if err != nil {
					x = 0
					fmt.Println("error parsing client data", err)
				}
				result = append(result, toByte(float32(x))...)

			case 5, 6:
				var x uint64
				x, err = strconv.ParseUint(data[i * 11 + j],10, 32)
				if err != nil {
					x = 0
					fmt.Println("error parsing client data", err)
				}
				result = append(result, toByte(uint32(x))...)

			case 9:
				var x uint64
				x, err = strconv.ParseUint(data[i * 11 + j],10, 32)
				if err != nil {
					x = 0
					fmt.Println("error parsing client data", err)
				}
				result = append(result, toByte(uint16(x))...)
				
			}
		}
	}

	return result
}

func toByte(in interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, in)
	if err != nil {
		in = 0
		fmt.Println("error parsing client data", err)
	}

	fmt.Println("___", in, ":::", buf.Bytes())

	return buf.Bytes()
}
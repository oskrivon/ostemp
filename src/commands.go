package main

func safeCommands() map[string]command {
	m := make(map[string]command)
	m["get raw sensor data"] = command{
		name:           "get raw sensor data",
		instruction:    []byte{0x4f, 0x05, 0x00, 0x43, 0x47},
		responseLenght: 44,
	}
	m["set ga options"] = command{
		name:        "set ga options",
		instruction: []byte{0x4f, 0x07, 0x00, 0x43, 0xb7},
		//responseLenght: 44,
	}
	m["get ga options"] = command{
		name:           "get ga options",
		instruction:    []byte{0x4f, 0x06, 0x00, 0x43, 0xb7},
		responseLenght: 143,
	}
	m["open valve"] = command{}
	m["close vavle"] = command{}
	m["set flow"] = command{
		name:     "set flow",
		adress:   33,
		quantity: 1,
	}
	m["get flow"] = command{
		name:     "get flow",
		adress:   32,
		quantity: 1,
	}
	m["capacity unit"] = command{
		name:     "capacity unit",
		adress:   33272,
		quantity: 7,
	}
	m["set FC options"] = command{}
	m["get FC options"] = command{}
	m["get ppm"] = command{
		name:           "get ppm",
		instruction:    []byte{0x4f, 0x02, 0x00, 0x41, 0x77},
		responseLenght: 31,
	}

	return m
}
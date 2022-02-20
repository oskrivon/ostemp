package main

/* func backgroundDataReading(conn net.Conn, system system, fcCommand, gaCommand command) {
	for {
		for _, fc := range system.flowController {
			r, err := fc.sendCommand(fcCommand, 3, 0, "get")
			fmt.Println(r)
			if err != nil {
				r = "error"
				fmt.Println("error fc")
			}
		}

		for _, ga := range system.gasAnalyzer {
			r, err := ga.sendCommand(gaCommand, 1)
			if err != nil {
				r = "error"
				fmt.Println("error ga")
			}

			var arr []string
			arr = append(arr, time.Now().Format("02-01-2006 15:04:05"))
			rForLog := strings.TrimSuffix(r, " ")
			arrayForLog := strings.Split(rForLog, " ")

			arr = append(arr, arrayForLog...)

			record(arr)
			fmt.Println(arr)

			conn.Write([]byte(r))
		}

		time.Sleep(100 * time.Millisecond)
	}
} */

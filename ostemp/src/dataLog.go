package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func createLog() (string, error){
	records := [][]string{
		{"timestamp", "sensor_1", "sensor_2", "sensor_3", "sensor_4"},
		{},
	}

	nowTime := time.Now().Format("02.01.2006-15.04.05")
	filename := nowTime + ".csv"

	csvFile, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)

	for _, rows := range records {
		_ = csvwriter.Write(rows)
	}

	csvwriter.Flush()

	return filename, err
}

func record(data []string) {
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(data); err != nil {
		fmt.Println("error record: ", err)
	}
}
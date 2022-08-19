// This app will convert the JSON emitted by a NR bulk export job to CSV.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type nrJSONStruct []struct {
	Attributes struct {
		EventType string `json:"eventType"`
		User      string `json:"user"`
		Account   string `json:"account"`
		Timestamp int64  `json:"timestamp"`
	} `json:"attributes"`
}

func main() {
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	jsonFileName := flag.String("file", "", "Name of the file to convert")
	flag.Parse()

	if *logVerbose {
		fmt.Println("JSON convert util v1.0")
		fmt.Println("Verbose logging enabled")
	}

	jsonFile, err := ioutil.ReadFile(*jsonFileName)
	if err != nil {
		fmt.Println(err)
	}

	var exportJSON nrJSONStruct
	if err := json.Unmarshal(jsonFile, &exportJSON); err != nil {
		fmt.Println(err)
	}

	csvFile, err := os.Create(*jsonFileName + ".csv")
	for _, record := range exportJSON {
		if err != nil {
			fmt.Println(err)
		}

		csvRecord := fmt.Sprintf("%v,%v,%v,%v\n", record.Attributes.EventType, record.Attributes.User, record.Attributes.Account, record.Attributes.Timestamp)
		csvFile.WriteString(csvRecord)
	}
	csvFile.Close()

}

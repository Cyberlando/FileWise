package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func createJSON(data []FileInfo, name string) {
	// Marshal the file information to JSON.
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error Creating JSON: %v", err)
	}

	// Write the JSON data to a file.
	path := fmt.Sprintf("../results/%v.json", name)
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	} else {
		_temp := fmt.Sprintf("Data successfully written to %s.json", name)
		fmt.Println(_temp)
	}

}

type FileInfo struct {
	Name      string
	Path      string
	Size      float64 //in kilobytes
	Ext       string
	Encoding  string
	Is_Binary bool
}

func main() {
	var files []FileInfo
	var binaryTrue []FileInfo
	var binaryFalse []FileInfo

	file, err := os.Open("../results/results.json")
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
	}

	decode := json.NewDecoder(file)
	err = decode.Decode(&files)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
	}

	for _, obj := range files {

		if obj.Is_Binary {
			binaryTrue = append(binaryTrue, obj)
		} else {
			binaryFalse = append(binaryFalse, obj)
		}
	}

	createJSON(binaryTrue, "binaryTrue")
	createJSON(binaryFalse, "binaryFalse")

	fmt.Printf("\nTotal Files: %v \n", len(files))
	fmt.Printf("Percentage of Text Files: %.2f%%\n", float64(len(binaryFalse))/float64(len(files))*100)
	fmt.Printf("Percentage of Binary Files: %.2f%%\n", float64(len(binaryTrue))/float64(len(files))*100)
}

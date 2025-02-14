package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/saintfish/chardet"
)

type FileInfo struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Size     float64 `json:"size"` // in kilobytes
	Ext      string  `json:"ext"`
	Encoding string  `json:"encoding"`  // detected file encoding
	IsBinary bool    `json:"is_binary"` // true if file appears to be binary
}

// File size in kilobytes
func getFileSize(filePath string) (float64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("error getting file size for %s: %w", filePath, err)
	}
	return float64(info.Size()) / 1024, nil
}

// isBinaryFromData applies two methods on the provided data:
//  1. check if null byte (0x00) is present
//  2. If more than 30% of the bytes are non-printable(excluding tab, newline, and carriage return)

func isBinaryFromData(data []byte) bool {
	nonPrintable := 0

	//Check 1: Check for null byte
	for _, b := range data {
		if b == 0 {
			return true
		}
	}

	// Check 2: Count non-printable characters
	// Special Considerations : ASCII 32-126 , tab (9), newline (10), and carriage return (13)
	for _, b := range data {

		if (b < 32 || b > 126) && b != 9 && b != 10 && b != 13 {
			nonPrintable++
		}
	}
	if len(data) > 0 && float64(nonPrintable)/float64(len(data)) > 0.3 {
		return true
	}
	return false
}

// reads the first 1024 bytes:
// 1. Use chardet library to detect encoding
// 2. Calls isBinary..... with the same data to determine if the file is binary
func detectEncodingAndBinary(filePath string) (string, bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", false, err
	}
	defer f.Close()

	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", false, err
	}
	data := buf[:n]

	// Use chardet to detect encoding
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return "", false, err
	}
	encoding := result.Charset

	//Calls isBinary... with same data in order to acess the file once
	binary := isBinaryFromData(data)
	return encoding, binary, nil
}

// worker processes directories from dirsChan ->
// For each file, it gathers information -> results
func worker(dirsChan chan string, results chan<- FileInfo, wg *sync.WaitGroup) {
	for dir := range dirsChan {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("Error reading directory %s: %v", dir, err)
			wg.Done()
			continue
		}

		for _, entry := range entries {
			fullPath := filepath.Join(dir, entry.Name())
			if entry.IsDir() { //subdirectory: increment the WaitGroup and send its path
				wg.Add(1)
				go func(subDir string) {
					dirsChan <- subDir
				}(fullPath)
			} else {
				info, err := entry.Info()
				if err != nil {
					log.Printf("Error getting info for %s: %v", fullPath, err)
					continue
				}

				//check for symlink, resolve if it exists
				if info.Mode()&os.ModeSymlink != 0 {
					resolved, err := filepath.EvalSymlinks(fullPath)
					if err != nil {
						log.Printf("Error resolving symlink %s: %v", fullPath, err)
						continue
					}
					fullPath = resolved
				}

				size, err := getFileSize(fullPath)
				if err != nil {
					log.Printf("Error getting file size for %s: %v", fullPath, err)
					continue
				}

				// detect encoding and check if file is binary
				encoding, binary, err := detectEncodingAndBinary(fullPath)
				if err != nil {
					log.Printf("Error detecting encoding for %s: %v", fullPath, err)
					encoding = "unknown"
					binary = false
				}

				ext := filepath.Ext(entry.Name())
				if ext != "" {
					ext = strings.TrimPrefix(ext, ".")
				} else {
					ext = "{BLANK}"
				}

				results <- FileInfo{
					Name:     entry.Name(),
					Path:     filepath.ToSlash(fullPath),
					Size:     size,
					Ext:      ext,
					Encoding: encoding,
					IsBinary: binary,
				}
			}
		} // Finished processing A directory
		wg.Done()
	}
}

func createJSON(data []FileInfo, name string) {
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v", err)
	}

	// Write JSON data to file
	path := fmt.Sprintf("./results/%v.json", name)
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v \n", err)
	} else {
		_temp := fmt.Sprintf("Data successfully written to %s.json", name)
		fmt.Println(_temp)
	}

}

// Seperates Binary & Non-Binary files
func splitter(data []FileInfo) {

	var binaryTrue []FileInfo
	var binaryFalse []FileInfo

	for _, file := range data {
		if file.IsBinary {
			binaryTrue = append(binaryTrue, file)
		} else {
			binaryFalse = append(binaryFalse, file)
		}
	}
	createJSON(binaryTrue, "binaryTrue")
	createJSON(binaryFalse, "binaryFalse")
	fmt.Printf("Percentage of non-binary files: %.3f%% \n", float64(len(binaryFalse))/float64(len(data))*100)
	fmt.Printf("Percentage of binary files: %.3f%% \n", float64(len(binaryTrue))/float64(len(data))*100)
	fmt.Println("File seperation successful!")
}

func main() {
	// Default to the current directory if no path is given
	rootDir := flag.String("p", ".", "Directory to scan")
	flag.Parse()

	absPath, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", *rootDir, err)
	}
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		log.Fatalf("Invalid directory provided: %s", absPath)
	}

	// Create buffered channels.
	dirsChan := make(chan string, 10000)
	results := make(chan FileInfo, 10000)
	var wg sync.WaitGroup

	// Start processing root directory
	wg.Add(1)
	dirsChan <- absPath

	// Launch scale able number of worker goroutines: (50 worked well for a laptop)
	numWorkers := 50
	for i := 0; i < numWorkers; i++ {
		go worker(dirsChan, results, &wg)
	}

	// Close channels when all directories have been processed.
	go func() {
		wg.Wait()
		close(dirsChan)
		close(results)
	}()

	// Collect file information from the result channel
	var files []FileInfo
	for file := range results {
		files = append(files, file)
	}
	//Creates Master Result File
	createJSON(files, "results")
	splitter(files)
}

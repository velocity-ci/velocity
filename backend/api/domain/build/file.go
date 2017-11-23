package build

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/docker/go/canonical/json"
)

type fileManager struct {
}

// TODO: include query ability
func (m *fileManager) GetByID(id string) []*StreamLine {
	outputStream := []*StreamLine{}

	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		var line StreamLine
		err := json.Unmarshal(lineBytes, &line)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(line.Output)

		outputStream = append(outputStream, &line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return outputStream
}

func (m *fileManager) Save(streamLine *StreamLine) *StreamLine {
	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", streamLine.OutputStream.ID)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()
	jsonLine, err := json.Marshal(ResponseStreamLine{
		LineNumber: streamLine.LineNumber,
		Timestamp:  streamLine.Timestamp,
		Output:     streamLine.Output,
	})
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.WriteString(string(jsonLine))
	if err != nil {
		log.Fatal(err)
	}

	return streamLine
}

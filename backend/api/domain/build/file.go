package build

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/docker/go/canonical/json"
)

type fileManager struct {
	logger *log.Logger
}

func newFileManager() *fileManager {
	err := os.MkdirAll("/tmp/velocity-workspace/logs", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return &fileManager{
		logger: log.New(os.Stdout, "[file:build]", log.Lshortfile),
	}
}

// TODO: include query ability
func (m *fileManager) GetByID(id string) ([]StreamLine, uint64) {
	outputStream := []StreamLine{}
	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
	file, err := os.Open(filePath)
	if err != nil {
		m.logger.Println(err)
		return []StreamLine{}, uint64(0)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := uint64(0)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		var line StreamLine
		err := json.Unmarshal(lineBytes, &line)
		if err != nil {
			m.logger.Fatal(err)
		}
		fmt.Println(line.Output)

		outputStream = append(outputStream, line)
		count++
	}

	if err := scanner.Err(); err != nil {
		m.logger.Fatal(err)
	}

	return outputStream, count
}

func (m *fileManager) DeleteByID(id string) {
	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
	os.RemoveAll(filePath)
}

func (m *fileManager) SaveStreamLine(streamLine StreamLine) StreamLine {
	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", streamLine.BuildStepStreamID)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		m.logger.Fatal(err)
		return streamLine
	}
	defer file.Close()
	jsonLine, err := json.Marshal(ResponseStreamLine{
		LineNumber: streamLine.LineNumber,
		Timestamp:  streamLine.Timestamp,
		Output:     streamLine.Output,
	})
	if err != nil {
		m.logger.Fatal(err)
	}
	_, err = file.WriteString(fmt.Sprintf("%s\n", jsonLine))
	if err != nil {
		m.logger.Fatal(err)
	}

	m.logger.Printf("wrote to %s:%d", filePath, streamLine.LineNumber)

	return streamLine
}

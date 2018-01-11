package build

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/docker/go/canonical/json"
)

type logFile struct {
	contents   [][]byte
	totalLines uint64

	needsFlush bool
	mux        sync.RWMutex
}

type fileManager struct {
	logger   *log.Logger
	logFiles map[string]*logFile // id: logFile

	wg   *sync.WaitGroup
	stop bool
}

func NewFileManager(wg *sync.WaitGroup) *fileManager {
	err := os.MkdirAll("/tmp/velocity-workspace/logs", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fM := &fileManager{
		logger:   log.New(os.Stdout, "[file:build]", log.Lshortfile),
		logFiles: map[string]*logFile{},
		wg:       wg,
		stop:     false,
	}

	return fM
}

func (m *fileManager) StartWorker() {
	m.wg.Add(1)
	for m.stop == false {
		for id, lF := range m.logFiles {
			lF.mux.Lock()
			if lF.needsFlush {
				filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
				file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
				err = file.Truncate(0)
				if err != nil {
					log.Panic(err)
				}
				writer := bufio.NewWriter(file)
				for _, s := range lF.contents {
					writer.Write(s)
				}
				writer.Flush()
				lF.needsFlush = false
				m.logger.Printf("flushed logs for %s", id)
			}
			lF.mux.Unlock()
		}
		time.Sleep(3 * time.Second)
	}
	m.logger.Println("stopped file manager")
	m.wg.Done()
}

func (m *fileManager) StopWorker() {
	m.stop = true
}

func (m *fileManager) GetByID(id string, q StreamLineQuery) ([]StreamLine, uint64) {
	outputStream := []StreamLine{}
	logFile := m.getStreamLogFile(id)
	skipCounter := uint64(0)

	for _, l := range logFile.contents {
		if uint64(len(outputStream)) >= q.Amount {
			break
		}
		if skipCounter < (q.Page-1)*q.Amount {
			skipCounter++
			break
		}
		var sL StreamLine
		err := json.Unmarshal(l, &sL)
		if err != nil {
			log.Println(err)
		}
		// parts := strings.SplitN(l, " ", 2)
		// timestampUnixNano, _ := strconv.ParseInt(parts[0], 10, 64)
		// outputStream = append(outputStream, StreamLine{
		// 	BuildStepStreamID: id,
		// 	LineNumber:        uint64(i),
		// 	Timestamp:         time.Unix(0, timestampUnixNano),
		// 	Output:            parts[1],
		// })

		outputStream = append(outputStream, sL)
	}
	return outputStream, logFile.totalLines
}

func (m *fileManager) SaveStreamLine(streamLine StreamLine) StreamLine {
	logFile := m.getStreamLogFile(streamLine.BuildStepStreamID)
	logFile.mux.Lock()
	defer logFile.mux.Unlock()

	jsonSL, err := json.Marshal(&streamLine)
	if err != nil {
		log.Println(err)
	}

	if streamLine.LineNumber >= logFile.totalLines {
		logFile.contents = append(logFile.contents, jsonSL)
		logFile.totalLines++
	} else {
		logFile.contents[streamLine.LineNumber] = jsonSL
	}
	logFile.needsFlush = true
	return streamLine
}

func (m *fileManager) getStreamLogFile(id string) *logFile {
	if _, ok := m.logFiles[id]; !ok {
		filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
		file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Panic(err)
			return nil
		}
		contents := [][]byte{}
		totalLines := uint64(0)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			contents = append(contents, scanner.Bytes())
			totalLines++
		}
		m.logFiles[id] = &logFile{
			contents:   contents,
			totalLines: totalLines,
			needsFlush: false,
		}
		file.Close()
	}

	return m.logFiles[id]
}

func (m *fileManager) DeleteByID(id string) {
	filePath := fmt.Sprintf("/tmp/velocity-workspace/logs/%s", id)
	os.RemoveAll(filePath)
	if _, ok := m.logFiles[id]; ok {
		delete(m.logFiles, id)
	}
}

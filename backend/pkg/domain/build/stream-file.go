package build

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/Sirupsen/logrus"
)

type logFile struct {
	contents   [][]byte
	totalLines int

	needsFlush bool
	mux        sync.RWMutex
}

type StreamFileManager struct {
	logDirectory string
	logFiles     map[string]*logFile

	wg   *sync.WaitGroup
	stop bool
}

func NewStreamFileManager(
	wg *sync.WaitGroup,
	logDirectory string,
) *StreamFileManager {
	err := os.MkdirAll(logDirectory, os.ModePerm)
	if err != nil {
		logrus.Fatal(err)
	}
	fM := &StreamFileManager{
		logFiles:     map[string]*logFile{},
		logDirectory: logDirectory,
		wg:           wg,
		stop:         false,
	}

	return fM
}

func (m *StreamFileManager) StartWorker() {
	m.wg.Add(1)
	for m.stop == false {
		for id, lF := range m.logFiles {
			lF.mux.Lock()
			if lF.needsFlush {
				filePath := fmt.Sprintf("%s/%s", m.logDirectory, id)
				file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
				err = file.Truncate(0)
				if err != nil {
					logrus.Error(err)
				}
				writer := bufio.NewWriter(file)
				for _, s := range lF.contents {
					writer.Write(s)
					writer.WriteRune('\n')
				}
				writer.Flush()
				lF.needsFlush = false
				logrus.Infof("flushed logs for %s", id)
			}
			lF.mux.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
	logrus.Info("stopped file manager")
	m.wg.Done()
}

func (m *StreamFileManager) StopWorker() {
	m.stop = true
}

func (m *StreamFileManager) getLinesByStream(s *Stream, q *domain.PagingQuery) (r []*StreamLine, t int) {
	logFile := m.getStreamLogFile(s.ID)
	skipCounter := 0

	for _, l := range logFile.contents {
		if q.Limit > 0 && len(r) >= q.Limit {
			break
		}
		if q.Limit > 0 && skipCounter < (q.Page-1)*q.Limit {
			skipCounter++
			break
		}
		var sL StreamLine
		err := json.Unmarshal(l, &sL)
		if err != nil {
			logrus.Error(err)
		}

		r = append(r, &sL)
	}
	return r, logFile.totalLines
}

func (m *StreamFileManager) saveStreamLine(streamLine *StreamLine) *StreamLine {
	logFile := m.getStreamLogFile(streamLine.StreamID)
	logFile.mux.Lock()
	defer logFile.mux.Unlock()

	jsonSL, err := json.Marshal(&streamLine)
	if err != nil {
		logrus.Error(err)
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

func (m *StreamFileManager) getStreamLogFile(id string) *logFile {
	if _, ok := m.logFiles[id]; !ok {
		filePath := fmt.Sprintf("%s/%s", m.logDirectory, id)
		file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			logrus.Error(err)
			return nil
		}
		contents := [][]byte{}
		totalLines := 0
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

func (m *StreamFileManager) deleteByID(id string) {
	filePath := fmt.Sprintf("%s/%s", m.logDirectory, id)
	os.RemoveAll(filePath)
	if _, ok := m.logFiles[id]; ok {
		delete(m.logFiles, id)
	}
}

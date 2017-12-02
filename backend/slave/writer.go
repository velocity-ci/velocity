package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

type SlaveWriter struct {
	ws             *websocket.Conn
	status         string
	OutputStreamID string
	LineNumber     uint64
}

func NewSlaveWriter(ws *websocket.Conn) *SlaveWriter {
	return &SlaveWriter{
		ws: ws,
	}
}

func (w SlaveWriter) Write(p []byte) (n int, err error) {
	lM := slave.SlaveBuildLogMessage{
		OutputStreamID: w.OutputStreamID,
		LineNumber:     w.LineNumber,
		Status:         w.status,
		Output:         string(p),
	}
	m := slave.SlaveMessage{
		Type: "log",
		Data: lM,
	}
	err = w.ws.WriteJSON(m)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *SlaveWriter) SetLineNumber(num uint64) {
	w.LineNumber = num
}

func (w *SlaveWriter) SetStatus(s string) {
	w.status = s
}

func (w *SlaveWriter) SetOutputStreamID(id string) {
	w.OutputStreamID = id
}

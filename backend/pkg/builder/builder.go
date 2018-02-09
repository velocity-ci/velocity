package builder

import (
	"sync"

	"github.com/gorilla/websocket"
)

// TODO: move to REST package and introduce interface.
type safeWebsocket struct {
	ws        *websocket.Conn
	writeLock sync.RWMutex
	readLock  sync.RWMutex
}

func (sws *safeWebsocket) WriteJSON(m interface{}) error {
	sws.writeLock.Lock()
	defer sws.writeLock.Unlock()

	return sws.ws.WriteJSON(m)
}

func (sws *safeWebsocket) ReadJSON(m interface{}) error {
	sws.readLock.Lock()
	defer sws.readLock.Unlock()

	return sws.ws.ReadJSON(m)
}

func (sws *safeWebsocket) Close() error {
	sws.writeLock.Lock()
	defer sws.writeLock.Unlock()

	return sws.ws.Close()
}

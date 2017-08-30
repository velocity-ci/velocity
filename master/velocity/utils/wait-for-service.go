package utils

import (
	"fmt"
	"log"
	"net"
	"time"
)

// WaitForService - Waits for a given service on TCP port to be ready.
func WaitForService(address string, logger *log.Logger) bool {

	for i := 0; i < 12; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			logger.Println("Connection error:", err)
		} else {
			conn.Close()
			logger.Println(fmt.Sprintf("Connected to %s", address))
			return true
		}
		time.Sleep(5 * time.Second)
	}

	return false
}

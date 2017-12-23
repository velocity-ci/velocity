package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

func main() {
	// Register Slave POST /v1/slaves {"id": ""}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	masterAddress := os.Getenv("MASTER_ADDRESS") // http://master || https://master
	if masterAddress == "" {
		log.Fatal("Missing MASTER_ADDRESS environment variable")
		os.Exit(1)
	}
	slaveSecret := os.Getenv("SLAVE_SECRET")
	if slaveSecret == "" {
		log.Fatal("Missing SLAVE_SECRET environment variable")
		os.Exit(1)
	}

	if masterAddress[:5] != "https" {
		log.Println("WARNING: Builds are not protected by TLS.")
	}

	if !waitForService(client, masterAddress) {
		log.Fatalf("Could not connect to: %s", masterAddress)
		os.Exit(1)
	}

	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	uniqueID := fmt.Sprintf("%d", randNumber.Int63())
	registerPayload, _ := json.Marshal(map[string]string{"id": uniqueID})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/slaves", masterAddress), bytes.NewBuffer(registerPayload))
	req.Header.Set("Authorization", fmt.Sprintf("basic %s", slaveSecret))

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if res.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		log.Fatalf("Registration failed. StatusCode: %d. Message: %s", res.StatusCode, string(bodyBytes))
		res.Body.Close()
		os.Exit(1)
	}

	a := auth{}
	json.NewDecoder(res.Body).Decode(&a)
	log.Printf("Registered %s.", uniqueID)

	// Connect to WebSocket on successful registration GET /v1/slaves/ws with authToken in header
	websocketAddress := strings.Replace(masterAddress, "http", "ws", 1)
	var dialer *websocket.Dialer
	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("bearer %s", a.Token))

	websocketConn, _, err := dialer.Dial(fmt.Sprintf("%s/v1/slaves/ws", websocketAddress), headers)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Printf("Connected to %s", websocketAddress)

	monitorCommands(websocketConn)
}

type auth struct {
	Token string `json:"authToken"`
}

func waitForService(client *http.Client, address string) bool {

	for i := 0; i < 6; i++ {
		_, err := client.Get(address)
		if err != nil {
			log.Println("Connection error:", err)
		} else {
			log.Println(fmt.Sprintf("Connected to %s", address))
			return true
		}
		time.Sleep(5 * time.Second)
	}

	return false
}

func monitorCommands(ws *websocket.Conn) {
	for {
		command := &slave.CommandMessage{}
		err := ws.ReadJSON(command)
		if err != nil {
			log.Println(err)
			log.Println("Closing WebSocket")
			ws.Close()
			main()
			return
		}

		if command.Command == "build" {
			log.Printf("Got Build: %v", command.Data)
			runBuild(command.Data.(*slave.BuildCommand), ws)
		} else if command.Command == "known-hosts" {
			log.Printf("Got known hosts: %v", command.Data)
			updateKnownHosts(command.Data.(*slave.KnownHostCommand))
		}
	}
}

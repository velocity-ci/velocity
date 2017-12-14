package knownhost

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
)

type FileManager struct {
	logger         *log.Logger
	knownHostsPath string
}

func NewFileManager() *FileManager {
	processUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(fmt.Sprintf("%s/.ssh/", processUser.HomeDir), os.ModePerm)
	fM := &FileManager{
		logger:         log.New(os.Stdout, "[file:knownhost]", log.Lshortfile),
		knownHostsPath: fmt.Sprintf("%s/.ssh/known_hosts", processUser.HomeDir),
	}
	f, err := os.OpenFile(fM.knownHostsPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	return fM
}

func (m FileManager) Save(e *KnownHost) error {
	if m.Exists(e.Entry) {
		return nil
	}
	f, err := os.OpenFile(m.knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s\n", e.Entry))
	if err != nil {
		return err
	}

	m.logger.Printf("Wrote %s to %s", e.Entry, m.knownHostsPath)

	return nil
}

func (m FileManager) Exists(e string) bool {
	f, err := os.OpenFile(m.knownHostsPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		if scanner.Text() == e {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return false
}

func (m FileManager) All() []KnownHost {
	knownHosts := []KnownHost{}

	f, err := os.OpenFile(m.knownHostsPath, os.O_RDONLY|os.O_SYNC, 0644)
	if err != nil {
		log.Fatal(err)
		return knownHosts
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		knownHost := KnownHost{
			Entry: scanner.Text(),
		}
		knownHosts = append(knownHosts, knownHost)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return knownHosts

}

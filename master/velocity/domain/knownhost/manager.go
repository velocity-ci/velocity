package knownhost

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

const knownHostsPath = "/root/.ssh/known_hosts"

type Manager struct {
	logger *log.Logger
}

func NewManager(fileLogger *log.Logger) *Manager {
	os.MkdirAll("/root/.ssh/", os.ModePerm)
	f, err := os.OpenFile(knownHostsPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	return &Manager{
		logger: fileLogger,
	}
}

func (m Manager) Save(e *domain.KnownHost) error {
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s\n", e.Entry))
	if err != nil {
		return err
	}

	m.logger.Printf("Wrote %s to %s", e.Entry, knownHostsPath)

	return nil
}

func (m Manager) Exists(e string) bool {
	f, err := os.OpenFile(knownHostsPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		if scanner.Text() == e {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return false
}

func (m Manager) All() []domain.KnownHost {
	knownHosts := []domain.KnownHost{}

	f, err := os.OpenFile(knownHostsPath, os.O_RDONLY|os.O_SYNC, 0644)
	if err != nil {
		log.Fatal(err)
		return knownHosts
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		knownHost := domain.KnownHost{
			Entry: scanner.Text(),
		}
		knownHosts = append(knownHosts, knownHost)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return knownHosts

}

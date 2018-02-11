package knownhost

import (
	"fmt"
	"os"
	"os/user"

	"github.com/Sirupsen/logrus"
)

type FileManager struct {
	knownHostsPath string
}

func NewFileManager(homedir string) *FileManager {
	if homedir == "" {
		processUser, err := user.Current()
		if err != nil {
			logrus.Error(err)
		}
		homedir = processUser.HomeDir
	}
	os.MkdirAll(fmt.Sprintf("%s/.ssh/", homedir), os.ModePerm)
	fM := &FileManager{
		knownHostsPath: fmt.Sprintf("%s/.ssh/known_hosts", homedir),
	}
	f, err := os.OpenFile(fM.knownHostsPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error(err)
	}
	defer f.Close()

	return fM
}

func (m FileManager) save(e *KnownHost) error {
	// if m.Exists(e.Entry) {
	// 	return nil
	// }
	f, err := os.OpenFile(m.knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s\n", e.Entry))
	if err != nil {
		return err
	}

	logrus.Debugf("Wrote %s to %s", e.Entry, m.knownHostsPath)

	return nil
}

func (m FileManager) clear() error {
	f, err := os.OpenFile(m.knownHostsPath, os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Truncate(0)
	return nil
}

func (m FileManager) WriteAll(kHs []*KnownHost) {
	if err := m.clear(); err != nil {
		logrus.Error(err)
	}
	for _, k := range kHs {
		m.save(k)
	}
}

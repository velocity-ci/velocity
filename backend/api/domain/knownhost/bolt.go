package knownhost

// import (
// 	"encoding/json"
// 	"log"
// 	"os"

// 	"github.com/boltdb/bolt"
// )

// type Manager struct {
// 	logger      *log.Logger
// 	bolt        *bolt.DB
// 	fileManager *FileManager
// }

// func NewManager(
// 	bolt *bolt.DB,
// 	fileManager *FileManager,
// ) *Manager {
// 	m := &Manager{
// 		logger:      log.New(os.Stdout, "[bolt-knownhost]", log.Lshortfile),
// 		bolt:        bolt,
// 		fileManager: fileManager,
// 	}

// 	for _, h := range fileManager.All() {
// 		m.Save(&h)
// 	}

// 	for _, h := range m.FindAll() {
// 		m.fileManager.Save(&h)
// 	}

// 	return m
// }

// func (m *Manager) Save(h *KnownHost) error {
// 	if m.Exists(h.Entry) {
// 		return nil
// 	}

// 	m.fileManager.Save(h)

// 	tx, err := m.bolt.Begin(true)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	m.logger.Printf("saving known host: %s", h.Entry)

// 	knownHostsBucket, err := tx.CreateBucketIfNotExists([]byte("known-hosts"))
// 	if err != nil {
// 		return err
// 	}
// 	if knownHostsBucket == nil {
// 		knownHostsBucket = tx.Bucket([]byte("known-hosts"))
// 	}

// 	knownHostJSON, err := json.Marshal(h)
// 	id, _ := knownHostsBucket.NextSequence()

// 	knownHostsBucket.Put(Itob(int(id)), knownHostJSON)

// 	if err := tx.Commit(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (m *Manager) Exists(entry string) bool {
// 	tx, err := m.bolt.Begin(false)
// 	if err != nil {
// 		log.Fatal(err)
// 		return false
// 	}
// 	defer tx.Rollback()
// 	knownHostsBucket := tx.Bucket([]byte("known-hosts"))

// 	if knownHostsBucket == nil {
// 		return false
// 	}

// 	c := knownHostsBucket.Cursor()
// 	for k, _ := c.First(); k != nil; k, _ = c.Next() {
// 		v := knownHostsBucket.Get(k)
// 		h := KnownHost{}
// 		json.Unmarshal(v, &h)
// 		if h.Entry == entry {
// 			return true
// 		}
// 	}

// 	return false

// }

// func (m *Manager) FindAll() []KnownHost {
// 	knownHosts := []KnownHost{}

// 	tx, err := m.bolt.Begin(false)
// 	if err != nil {
// 		return knownHosts
// 	}
// 	defer tx.Rollback()

// 	b := tx.Bucket([]byte("known-hosts"))
// 	if b == nil {
// 		return knownHosts
// 	}

// 	c := b.Cursor()
// 	for k, _ := c.First(); k != nil; k, _ = c.Next() {
// 		v := b.Get(k)
// 		h := KnownHost{}
// 		err := json.Unmarshal(v, &h)
// 		if err == nil {
// 			knownHosts = append(knownHosts, h)
// 		}
// 	}

// 	return knownHosts
// }

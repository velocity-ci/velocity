package persistence

import (
	"log"
	"os"
	"sync"

	"github.com/velocity-ci/velocity/backend/architect/domain"
)

var once sync.Once

func init() {
	once.Do(func() {
		_, err := GetUser(domain.User{
			Username: "admin",
		})
		if err != nil {
			var password string
			if os.Getenv("ADMIN_PASSWORD") != "" {
				password = os.Getenv("ADMIN_PASSWORD")
			} else {
				password = GenerateRandomString(16)
			}
			user, err := domain.NewUser("admin", password)
			if err != nil {
				log.Fatal(err)
			}
			SaveUser(user)
			log.Printf("\n\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n\n", password)
		}

	})
}

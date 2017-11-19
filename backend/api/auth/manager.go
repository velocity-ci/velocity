package auth

import (
	"log"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/user"
)

func EnsureAdminUser(m *user.Manager) {
	_, err := m.GetByUsername("admin")
	if err != nil {
		var password string
		if os.Getenv("ADMIN_PASSWORD") != "" {
			password = os.Getenv("ADMIN_PASSWORD")
		} else {
			password = GenerateRandomString(16)
		}
		user := &user.User{Username: "admin"}
		user.HashPassword(password)
		m.Save(user)
		log.Printf("\n\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n\n", password)
	}
}

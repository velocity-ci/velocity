package user

import (
	"log"

	"github.com/VJftw/velocity/master/velocity/domain"
	"github.com/VJftw/velocity/master/velocity/utils"
	"github.com/jinzhu/gorm"
)

// DBManager - Manages User entities on the Database.
type DBManager struct {
	dbLogger *log.Logger
	gorm     *gorm.DB
}

func createAdminUser(dbManager *DBManager) {
	_, err := dbManager.FindByUsername("admin")
	if err != nil {
		password := utils.GenerateRandomString(16)
		user := &domain.User{Username: "admin", Password: password}
		user.HashPassword()
		dbManager.Save(user)
		dbManager.dbLogger.Printf("\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n", password)
	}
}

// NewDBManager - Returns a new DBManager.
func NewDBManager(
	dbLogger *log.Logger,
	gorm *gorm.DB,
) *DBManager {
	dbManager := &DBManager{
		dbLogger: dbLogger,
		gorm:     gorm,
	}
	createAdminUser(dbManager)
	return dbManager
}

// Save - Saves a given user to the Database.
func (m *DBManager) Save(u *domain.User) error {

	existingDBUser := domain.User{}

	tx := m.gorm.Begin()

	err := tx.Where("username = ?", u.Username).First(&existingDBUser).Error
	if err != nil { // Not found, create
		err = tx.Create(u).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		tx.Save(u)
	}

	tx.Commit()

	m.dbLogger.Printf("Saved user %s", u.Username)

	return nil
}

// FindByUsername - Returns a user given their username, nil if not found.
func (m *DBManager) FindByUsername(username string) (*domain.User, error) {
	dbUser := &domain.User{}

	err := m.gorm.Where("username = ?", username).First(dbUser).Error
	if err != nil {
		m.dbLogger.Printf("user: %s %v", username, err)
		return nil, err
	}

	m.dbLogger.Printf("Got user %s", dbUser.Username)
	return dbUser, nil
}

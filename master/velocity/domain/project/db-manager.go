package project

import (
	"log"

	"github.com/VJftw/velocity/master/velocity/domain"
	"github.com/jinzhu/gorm"
)

// DBManager - Manages Project entities on the Database.
type DBManager struct {
	dbLogger *log.Logger
	gorm     *gorm.DB
}

// NewDBManager - Returns a new DBManager.
func NewDBManager(
	dbLogger *log.Logger,
	gorm *gorm.DB,
) *DBManager {
	return &DBManager{
		dbLogger: dbLogger,
		gorm:     gorm,
	}
}

// Save - Saves a given project to the Database.
func (m *DBManager) Save(p *domain.Project) error {

	existingProject := domain.Project{}

	tx := m.gorm.Begin()

	err := tx.Where("id = ?", p.ID).First(&existingProject).Error
	if err != nil { // Not found, create
		err = tx.Create(p).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		tx.Save(p)
	}

	tx.Commit()

	m.dbLogger.Printf("Saved project %s", p.ID)

	return nil
}

// FindByID - Returns a project given its ID, nil if not found.
func (m *DBManager) FindByID(ID string) (*domain.Project, error) {
	project := &domain.Project{}

	err := m.gorm.Where("id = ?", ID).First(&project).Error
	if err != nil {
		m.dbLogger.Printf("project: %s %v", ID, err)
		return nil, err
	}

	m.dbLogger.Printf("Got project %s", project.ID)
	return project, nil
}

func (m *DBManager) FindAll() []domain.Project {
	projects := []domain.Project{}

	err := m.gorm.Find(&projects).Error
	if err != nil {
		m.dbLogger.Printf("projects: %v", err)
		return []domain.Project{}
	}

	m.dbLogger.Printf("Got %d projects", len(projects))
	return projects
}

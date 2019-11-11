package db

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

type project struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	Address       string    `db:"address"`
	SSHPrivateKey string    `db:"ssh_private_key"`
	SSHHostKey    string    `db:"ssh_host_key"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func toProject(p *project) (*v1.Project, error) {
	createdAt, err := ptypes.TimestampProto(p.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.TimestampProto(p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &v1.Project{
		Id:   p.ID,
		Name: p.Name,
		Repository: &v1.Repository{
			Address: p.Address,
			SshConfig: &v1.SSHConfig{
				PrivateKey: p.SSHPrivateKey,
				HostKey:    p.SSHHostKey,
			},
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func toDBProject(p *v1.Project) (*project, error) {
	createdAt, err := ptypes.Timestamp(p.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.Timestamp(p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &project{
		ID:            p.GetId(),
		Name:          p.GetName(),
		Address:       p.GetRepository().GetAddress(),
		SSHPrivateKey: p.GetRepository().GetSshConfig().GetPrivateKey(),
		SSHHostKey:    p.GetRepository().GetSshConfig().GetHostKey(),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// CreateProject creates a given project
func (db *DB) CreateProject(ctx context.Context, p *v1.Project) (*v1.Project, error) {
	dbP, err := toDBProject(p)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	_, err = tx.NamedExec(`INSERT INTO projects
	(id, name, address, ssh_private_key, ssh_host_key, created_at, updated_at)
VALUES
	(:id, :name, :address, :ssh_private_key, :ssh_host_key, :created_at, :updated_at);`,
		map[string]interface{}{
			"id":              dbP.ID,
			"name":            dbP.Name,
			"address":         dbP.Address,
			"ssh_private_key": dbP.SSHPrivateKey,
			"ssh_host_key":    dbP.SSHHostKey,
			"created_at":      dbP.CreatedAt,
			"updated_at":      dbP.UpdatedAt,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return p, tx.Commit()
}

// UpdateProject updates a given project
func (db *DB) UpdateProject(ctx context.Context, p *v1.Project) (*v1.Project, error) {
	dbP, err := toDBProject(p)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	_, err = tx.NamedExec(`UPDATE projects
SET
	name=:name,
	address=:address,
	ssh_private_key=:ssh_private_key,
	ssh_host_key=:ssh_host_key,
	updated_at=:updated_at
WHERE id=:id;`,
		map[string]interface{}{
			"id":              dbP.ID,
			"name":            dbP.Name,
			"address":         dbP.Address,
			"ssh_private_key": dbP.SSHPrivateKey,
			"ssh_host_key":    dbP.SSHHostKey,
			"updated_at":      dbP.UpdatedAt,
		},
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return p, tx.Commit()
}

// GetProjectByUUID returns a project referenced by the given UUID
func (db *DB) GetProjectByUUID(ctx context.Context, uuid string) (*v1.Project, error) {
	dbP := project{}
	if err := db.Get(&dbP, `SELECT * FROM projects WHERE id=$1;`, uuid); err != nil {
		return nil, err
	}

	return toProject(&dbP)
}

// GetProjects returns the projects
func (db *DB) GetProjects(ctx context.Context) ([]*v1.Project, error) {
	dbPs := []project{}
	if err := db.Select(&dbPs, `SELECT * FROM projects;`); err != nil {
		return nil, err
	}
	res := []*v1.Project{}
	for _, dbP := range dbPs {
		p, err := toProject(&dbP)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}

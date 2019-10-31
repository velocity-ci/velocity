package db

import (
	"context"
	"log"
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

func (db *DB) CreateProject(ctx context.Context, p *v1.Project) (*v1.Project, error) {
	dbP, err := toDBProject(p)
	if err != nil {
		return nil, err
	}
	log.Println("converted Project to dbProject")
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Println("started transaction")
	_, err = tx.Exec(`INSERT INTO
projects
(id, name, address, ssh_private_key, ssh_host_key, created_at, updated_at)
VALUES
($1, $2, $3, $4, $5, $6, $7);`,
		dbP.ID, dbP.Name, dbP.Address, dbP.SSHPrivateKey, dbP.SSHHostKey, dbP.CreatedAt, dbP.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	log.Println("executed insert")
	return p, tx.Commit()
}

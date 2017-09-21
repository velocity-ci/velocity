package project

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	git "gopkg.in/src-d/go-git.v4"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

func FromRequest(b io.ReadCloser) (*domain.Project, error) {
	p := &domain.Project{}

	err := json.NewDecoder(b).Decode(p)
	if err != nil {
		return nil, err
	}
	p.ID = strings.Replace(strings.ToLower(p.Name), " ", "-", -1)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	return p, nil
}

func ValidatePOST(p *domain.Project, m *BoltManager) (bool, *middlewares.ResponseErrors) {
	hasErrors := false

	errs := projectErrors{}

	if len(p.Name) < 3 || len(p.Name) > 128 {
		errs.Name = []string{"Invalid name"}
		hasErrors = true
	}

	if len(p.Repository) < 8 || len(p.Repository) > 128 {
		errs.Repository = []string{"Invalid repository address"}
		hasErrors = true
	}

	_, err := ssh.ParsePrivateKey([]byte(p.PrivateKey))

	if err != nil {
		fmt.Println("SSH Parse Private Key Error:")
		fmt.Println(err)
		errs.PrivateKey = []string{"Invalid private key"}
		hasErrors = true
	}

	if hasErrors {
		return false, &middlewares.ResponseErrors{
			Errors: &errs,
		}
	}
	_, err = m.FindByID(p.ID)

	if err == nil {
		return false, &middlewares.ResponseErrors{
			Errors: &projectErrors{
				Name: []string{"Name already taken."},
			},
		}
	}

	return validateRepository(p)
}

type projectErrors struct {
	Name       []string `json:"name"`
	Repository []string `json:"repository"`
	PrivateKey []string `json:"key"`
}

func validateRepository(p *domain.Project) (bool, *middlewares.ResponseErrors) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", p.ID))
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	signer, _ := ssh.ParsePrivateKey([]byte(p.PrivateKey))
	auth := &gitssh.PublicKeys{User: "git", Signer: signer}

	// Clones the repository into the given dir, just as a normal git clone does
	_, err = git.PlainClone(dir, true, &git.CloneOptions{
		URL:   p.Repository,
		Depth: 1,
		Auth:  auth,
	})

	if err != nil {

		fmt.Println("Repository Clone Error:")
		fmt.Println(err.Error())

		return false, &middlewares.ResponseErrors{
			Errors: &projectErrors{
				Repository: []string{"Failed to clone repository."},
			},
		}

	}

	return true, nil
}

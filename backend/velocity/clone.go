package velocity

import (
	"fmt"
)

type Clone struct {
	BaseStep   `yaml:",inline"`
	Address    string `yaml:"-"`
	PrivateKey string `yaml:"-"`
	Submodule  bool   `json:"submodule" yaml:"submodule"`
}

func NewClone() *Clone {
	return &Clone{
		Submodule: false,
	}
}

func (c Clone) GetType() string {
	return "clone"
}

func (c Clone) GetDescription() string {
	return c.Description
}

func (c Clone) GetDetails() string {
	return fmt.Sprintf("submodule: %v", c.Submodule)
}

func (c *Clone) Execute(emitter Emitter, params map[string]Parameter) error {
	emitter.Write([]byte(fmt.Sprintf("%s\n## %s\n\x1b[0m", infoANSI, c.Description)))

	// log.Printf("Cloning %s", c.RepositoryAddress)

	// repo, dir, err := GitClone(c.Build.Project, false, true, c.Submodule, emitter)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }
	// log.Println("Done.")
	// // defer os.RemoveAll(dir)

	// w, err := repo.Worktree()
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }
	// log.Printf("Checking out %s", c.Build.CommitHash)
	// err = w.Checkout(&git.CheckoutOptions{
	// 	Hash: plumbing.NewHash(c.Build.CommitHash),
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }
	// log.Println("Done.")

	// os.Chdir(dir)
	return nil
}

func (cdB *Clone) Validate(params map[string]Parameter) error {
	return nil
}

func (c *Clone) SetParams(params map[string]Parameter) error {
	return nil
}

// func (c *Clone) SetBuild(b *Build) error {
// 	// c.Build = b
// 	return nil
// }

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

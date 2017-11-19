package slave

// import (
// 	"github.com/velocity-ci/velocity/backend/api/project"
// 	"github.com/velocity-ci/velocity/backend/velocity"
// )

// type CommandMessage struct {
// 	Command string  `json:"command"`
// 	Data    Message `json:"data"`
// }

// type BuildCommand struct {
// 	*velocity.Build
// 	Task *velocity.Task `json:"task"`
// }

// func NewBuildCommand(p *project.Project, t *velocity.Task, commitHash string, buildId uint64) *CommandMessage {
// 	return &CommandMessage{
// 		Command: "build",
// 		Data: BuildCommand{
// 			Task:  t,
// 			Build: velocity.NewBuild(p.ToTaskProject(), commitHash, buildId),
// 		},
// 	}
// }

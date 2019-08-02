package build

type Stoppable interface {
	GracefulStop() error
}

func eventBuildStart(plan *ConstructionPlan) {

}

func eventBuildComplete(plan *ConstructionPlan) {

}

func eventBuildFail(plan *ConstructionPlan, task *Task, err error) {

}

func eventBuildSuccess(plan *ConstructionPlan) {

}

func eventTaskStart(plan *ConstructionPlan, task *Task) {

}

func eventTaskComplete(plan *ConstructionPlan, task *Task) {

}

func eventTaskFail(plan *ConstructionPlan, task *Task, err error) {

}

func eventTaskSuccess(plan *ConstructionPlan, task *Task) {

}

package aicraft

type Agent struct {
	ID        string
	Name      string
	Tasks     []*Task
	DependsOn []string
	Output    map[string]interface{}
	Stream    <-chan interface{}
}

func NewAgent(id, name string, dependsOn []string) *Agent {
	return &Agent{
		ID:        id,
		Name:      name,
		Tasks:     []*Task{},
		DependsOn: dependsOn,
		Output:    make(map[string]interface{}),
	}
}

func (a *Agent) AddTask(task *Task) {
	a.Tasks = append(a.Tasks, task)
}

func (a *Agent) ExecuteTasks() error {
	for _, task := range a.Tasks {
		err := task.Execute()
		if err != nil {
			return err
		}
		a.Output[task.ID] = task.Result
		a.Stream = task.Stream
	}
	return nil
}

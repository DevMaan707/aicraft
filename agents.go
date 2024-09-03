package aicraft

import "sync"

type Agent struct {
	ID        string
	Name      string
	Tasks     []*Task
	DependsOn []string               // New field for specifying dependencies
	Output    map[string]interface{} // New field for storing the agent's output
}

func NewAgent(id, name string, dependsOn []string) *Agent {
	return &Agent{
		ID:        id,
		Name:      name,
		Tasks:     []*Task{},
		DependsOn: dependsOn,
		Output:    make(map[string]interface{}), // Initialize the output map
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
		a.Output[task.ID] = task.Result // Store task result in the agent's output
	}
	return nil
}

func (a *Agent) ExecuteTasksConcurrently() error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, len(a.Tasks))

	for _, task := range a.Tasks {
		wg.Add(1)
		go func(task *Task) {
			defer wg.Done()
			err := task.Execute()
			if err != nil {
				errors <- err
				return
			}
			mu.Lock()
			a.Output[task.ID] = task.Result // Store task result in the agent's output
			mu.Unlock()
		}(task)
	}
	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		return <-errors
	}
	return nil
}

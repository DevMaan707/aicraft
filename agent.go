package aicraft

import "sync"

type Agent struct {
	ID    string
	Name  string
	Tasks []*Task
}

func NewAgent(id, name string) *Agent {
	return &Agent{
		ID:    id,
		Name:  name,
		Tasks: []*Task{},
	}
}

func (a *Agent) AddTask(task *Task) {
	a.Tasks = append(a.Tasks, task)
}

func (a *Agent) ExecuteTasks() {
	for _, task := range a.Tasks {
		result := task.Execute()
		task.Result = result
	}
}

func (a *Agent) ExecuteTasksConcurrently() {
	var wg sync.WaitGroup
	for _, task := range a.Tasks {
		wg.Add(1)
		go func(task *Task) {
			defer wg.Done()
			result := task.Execute()
			task.Result = result
		}(task)
	}
	wg.Wait()
}

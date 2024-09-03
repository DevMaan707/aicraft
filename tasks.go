package aicraft

import "fmt"

type Task struct {
	ID     string
	Name   string
	Tool   *Tool
	Inputs map[string]interface{}
	Result interface{}
	Stream <-chan interface{}
}

func NewTask(id, name string, tool *Tool, inputs map[string]interface{}) *Task {
	return &Task{
		ID:     id,
		Name:   name,
		Tool:   tool,
		Inputs: inputs,
	}
}

func (t *Task) Execute() error {
	if t.Tool == nil {
		return fmt.Errorf("task %s has no tool assigned", t.Name)
	}

	result, stream, err := t.Tool.Execute(t.Inputs)
	if err != nil {
		return err
	}
	t.Result = result
	t.Stream = stream
	return nil
}

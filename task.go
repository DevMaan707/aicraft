package aicraft

type Task struct {
	ID     string
	Name   string
	Tool   *Tool
	Action func(tool *Tool) string
	Result string
}

func NewTask(id, name string, tool *Tool, action func(tool *Tool) string) *Task {
	return &Task{
		ID:     id,
		Name:   name,
		Tool:   tool,
		Action: action,
	}
}

func (t *Task) Execute() string {
	return t.Action(t.Tool)
}

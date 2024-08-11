package aicraft

type Tool struct {
	ID   string
	Name string
}

func NewTool(id, name string) *Tool {
	return &Tool{
		ID:   id,
		Name: name,
	}
}

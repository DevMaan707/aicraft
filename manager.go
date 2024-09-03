package aicraft

import (
	"log"
	"sync"
)

type Manager struct {
	Agents map[string]*Agent
	Tasks  map[string]*Task
	Tools  map[string]*Tool
	mu     sync.Mutex
}

func NewManager() *Manager {
	m := &Manager{
		Agents: make(map[string]*Agent),
		Tasks:  make(map[string]*Task),
		Tools:  make(map[string]*Tool),
	}
	m.initializePredefinedTools()
	return m
}

func (m *Manager) initializePredefinedTools() {
	m.Tools[TextToPDFTool.ID] = TextToPDFTool
	m.Tools[ImageGeneratorTool.ID] = ImageGeneratorTool
	m.Tools[PDFToEmbeddingsTool.ID] = PDFToEmbeddingsTool
	m.Tools[OpenAIContentGeneratorTool.ID] = OpenAIContentGeneratorTool
	m.Tools[QueryToEmbeddingTool.ID] = QueryToEmbeddingTool
	m.Tools[PDFExtractorTool.ID] = PDFExtractorTool
}

func (m *Manager) CreateAgent(id, name string, dependsOn []string) *Agent {
	agent := NewAgent(id, name, dependsOn)
	m.Agents[id] = agent
	return agent
}

func (m *Manager) CreateTask(id, name string, toolID string, inputs map[string]interface{}) *Task {
	tool, ok := m.Tools[toolID]
	if !ok {
		log.Fatalf("Error creating task %s: tool with ID %s not found", name, toolID)
		return nil
	}
	task := NewTask(id, name, tool, inputs)
	m.Tasks[id] = task
	return task
}

func (m *Manager) AssignTaskToAgent(agentID, taskID string) {
	agent := m.Agents[agentID]
	task := m.Tasks[taskID]
	agent.AddTask(task)
}

func (m *Manager) ExecuteWorkflow() error {
	executed := make(map[string]bool)

	for len(executed) < len(m.Agents) {
		for _, agent := range m.Agents {
			if executed[agent.ID] {
				continue
			}

			canExecute := true
			for _, dep := range agent.DependsOn {
				if !executed[dep] {
					canExecute = false
					break
				}
			}

			if canExecute {

				err := agent.ExecuteTasks()
				if err != nil {
					return err
				}

				for _, task := range agent.Tasks {
					if task.Stream != nil {
						go func(ch <-chan interface{}) {

						}(task.Stream)
					}
				}

				executed[agent.ID] = true
			}
		}
	}

	return nil
}

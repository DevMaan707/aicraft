package aicraft

import (
	"fmt"
	"log"
	"sync"
)

type WorkflowConfig struct {
	Tasks  []TaskConfig
	Agents []AgentConfig
}
type TaskConfig struct {
	ID     string
	Name   string
	ToolID string
	Inputs map[string]interface{}
}

type AgentConfig struct {
	ID        string
	Name      string
	DependsOn []string
	Tasks     []string
}

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
func (m *Manager) InitializeWorkflow(config WorkflowConfig) error {
	// Initialize Tasks
	for _, taskConfig := range config.Tasks {
		task := m.CreateTask(taskConfig.ID, taskConfig.Name, taskConfig.ToolID, taskConfig.Inputs)
		if task == nil {
			return fmt.Errorf("failed to create task: %s", taskConfig.Name)
		}
	}

	// Initialize Agents and Assign Tasks
	for _, agentConfig := range config.Agents {
		agent := m.CreateAgent(agentConfig.ID, agentConfig.Name, agentConfig.DependsOn)
		for _, taskID := range agentConfig.Tasks {
			m.AssignTaskToAgent(agent.ID, taskID)
		}
	}

	return nil
}
func (m *Manager) ExecuteAllWorkflows() error {
	executed := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for len(executed) < len(m.Agents) {
		for _, agent := range m.Agents {
			mu.Lock()
			if executed[agent.ID] {
				mu.Unlock()
				continue
			}
			canExecute := true
			for _, dep := range agent.DependsOn {
				if !executed[dep] {
					canExecute = false
					break
				}
			}
			mu.Unlock()

			if canExecute {
				wg.Add(1)
				go func(agent *Agent) {
					defer wg.Done()

					mu.Lock()
					for _, dep := range agent.DependsOn {
						for _, task := range agent.Tasks {
							if output, ok := m.Agents[dep].Output["task_extract_text"]; ok {
								task.Inputs["pdf_content"] = output
							}
						}
					}
					mu.Unlock()

					err := agent.ExecuteTasks()
					if err != nil {
						log.Printf("Error executing tasks for agent %s: %v", agent.ID, err)
					}

					mu.Lock()
					executed[agent.ID] = true
					mu.Unlock()

				}(agent)
			}
		}
		wg.Wait()
	}

	return nil
}

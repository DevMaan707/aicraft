package aicraft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const OpenAIURL = "https://api.openai.com/v1/chat/completions"

type Manager struct {
	Agents map[string]*Agent
	Tasks  map[string]*Task
	Tools  map[string]*Tool
}

func NewManager() *Manager {
	return &Manager{
		Agents: make(map[string]*Agent),
		Tasks:  make(map[string]*Task),
		Tools:  make(map[string]*Tool),
	}
}

func (m *Manager) CreateAgent(id, name string) *Agent {
	agent := NewAgent(id, name)
	m.Agents[id] = agent
	return agent
}

func (m *Manager) CreateTask(id, name string, toolID string, action func(tool *Tool) string) *Task {
	tool := m.Tools[toolID]
	task := NewTask(id, name, tool, action)
	m.Tasks[id] = task
	return task
}

func (m *Manager) CreateTool(id, name string) *Tool {
	tool := NewTool(id, name)
	m.Tools[id] = tool
	return tool
}

func (m *Manager) AssignTaskToAgent(agentID, taskID string) {
	agent := m.Agents[agentID]
	task := m.Tasks[taskID]
	agent.AddTask(task)
}

func (m *Manager) ExecuteOpenAITask(prompt, apiKey string) (string, error) {
	data := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", OpenAIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Debug: print the body to check the response format
	fmt.Println("API Response:", string(body))

	var response OpenAIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Check if Choices is non-empty
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from the API")
	}

	return response.Choices[0].Message.Content, nil
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

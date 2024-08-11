package main

import (
	"fmt"
	"log"

	"github.com/DevMaan707/aicraft"
)

func main() {
	apiKey := "your-api-key"
	// Initialize the Manager
	manager := aicraft.NewManager()

	// Create tools and store them in variables (used for demonstration purposes)
	codeAnalysisTool := manager.CreateTool("tool1", "Code Analysis Tool")
	storyWritingTool := manager.CreateTool("tool2", "Story Writing Tool")

	// Demonstrate the use of these variables (you could print them or use them in another context)
	fmt.Printf("Created tool: %s (%s)\n", codeAnalysisTool.Name, codeAnalysisTool.ID)
	fmt.Printf("Created tool: %s (%s)\n", storyWritingTool.Name, storyWritingTool.ID)

	// Create agents
	codeAnalysisAgent := manager.CreateAgent("agent1", "Code Analysis Agent")
	storyWritingAgent := manager.CreateAgent("agent2", "Story Writing Agent")

	// Define actions for tasks
	codeAnalysisAction := func(tool *aicraft.Tool) string {
		codeToAnalyze := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

		prompt := fmt.Sprintf("Analyze the following code for performance improvements and bugs:\n%s", codeToAnalyze)

		result, err := manager.ExecuteOpenAITask(prompt, apiKey)
		if err != nil {
			log.Fatalf("Error in code analysis task: %v", err)
		}
		return result
	}

	storyWritingAction := func(tool *aicraft.Tool) string {
		title := "The Brave Little Robot"
		prompt := fmt.Sprintf("Write a short story based on the following title: %s", title)

		result, err := manager.ExecuteOpenAITask(prompt, apiKey)
		if err != nil {
			log.Fatalf("Error in story writing task: %v", err)
		}
		return result
	}

	// Create tasks
	codeAnalysisTask := manager.CreateTask("task1", "Analyze Code", "tool1", codeAnalysisAction)
	storyWritingTask := manager.CreateTask("task2", "Write Story", "tool2", storyWritingAction)

	// Assign tasks to agents
	manager.AssignTaskToAgent("agent1", "task1")
	manager.AssignTaskToAgent("agent2", "task2")

	// Execute tasks
	fmt.Println("Executing tasks sequentially...")
	codeAnalysisAgent.ExecuteTasks()
	storyWritingAgent.ExecuteTasks()

	// Print results
	fmt.Println("Code Analysis Result:")
	fmt.Println(codeAnalysisTask.Result)

	fmt.Println("Story Writing Result:")
	fmt.Println(storyWritingTask.Result)

	codeAnalysisAgent.ExecuteTasksConcurrently()
	storyWritingAgent.ExecuteTasksConcurrently()

	fmt.Println("Concurrent Code Analysis Result:")
	fmt.Println(codeAnalysisTask.Result)

	fmt.Println("Concurrent Story Writing Result:")
	fmt.Println(storyWritingTask.Result)
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/DevMaan707/aicraft"
	"github.com/joho/godotenv"
)

var verbose bool

func main() {
	// Parse the verbose flag
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	cwd, _ := os.Getwd()
	fmt.Println("Current Working Directory:", cwd)
	err := godotenv.Load(cwd + "/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalf("Error: OPENAI_API_KEY not found in environment variables")
	}
	if verbose {
		fmt.Println("API KEY => " + apiKey)
	}

	manager := aicraft.NewManager()

	workflowConfig := aicraft.WorkflowConfig{
		Tasks: []aicraft.TaskConfig{
			{
				ID:     "task_extract_text",
				Name:   "Extract Text from PDF",
				ToolID: aicraft.PDFExtractorTool.ID,
				Inputs: map[string]interface{}{
					"pdf_url": "https://www.nber.org/system/files/working_papers/w27392/w27392.pdf",
					"verbose": verbose,
				},
			},
			{
				ID:     "task_convert_pdf",
				Name:   "Convert PDF to Embeddings",
				ToolID: aicraft.PDFToEmbeddingsTool.ID,
				Inputs: map[string]interface{}{
					"chunkSize":    800,
					"chunkOverlap": 100,
					"api_key":      apiKey,
					"verbose":      verbose,
				},
			},
			{
				ID:     "task_query_embedding",
				Name:   "Convert Query to Embedding",
				ToolID: aicraft.QueryToEmbeddingTool.ID,
				Inputs: map[string]interface{}{
					"query":   "Give me the summary of the context provided.",
					"api_key": apiKey,
					"verbose": verbose,
				},
			},
			{
				ID:     "task_optimize_query",
				Name:   "Optimize Query",
				ToolID: aicraft.OpenAIContentGeneratorTool.ID,
				Inputs: map[string]interface{}{
					"query":        "Give me the summary of the context provided in 500 words.",
					"context":      "Sample context here...",
					"chunkSize":    800,
					"chunkOverlap": 100,
					"api_key":      apiKey,
					"verbose":      verbose,
				},
			},
		},
		Agents: []aicraft.AgentConfig{
			{
				ID:   "agent1",
				Name: "Text Extractor",
				Tasks: []string{
					"task_extract_text",
				},
			},
			{
				ID:        "agent2",
				Name:      "PDF Processor",
				DependsOn: []string{"agent1"},
				Tasks: []string{
					"task_convert_pdf",
				},
			},
			{
				ID:        "agent3",
				Name:      "Query Embedding",
				DependsOn: []string{"agent2"},
				Tasks: []string{
					"task_query_embedding",
				},
			},
			{
				ID:        "agent4",
				Name:      "Query Optimizer",
				DependsOn: []string{"agent3"},
				Tasks: []string{
					"task_optimize_query",
				},
			},
		},
	}

	log.Println("Step 1: Initializing tasks and agents...")
	err = manager.InitializeWorkflow(workflowConfig)
	if err != nil {
		log.Fatalf("Error during workflow initialization: %v", err)
	}

	log.Println("Step 2: Executing all tasks and workflows...")
	err = manager.ExecuteAllWorkflows()

	if err != nil {
		log.Fatalf("Workflow completed with errors: %v", err)
	} else {
		// Extract the stream from agent4 for query optimization task
		agent := manager.Agents["agent4"]
		taskOptimizeQuery := manager.Tasks["task_optimize_query"]
		stream := agent.Stream

		if stream == nil {
			log.Fatalf("Error: No stream found for task %s", taskOptimizeQuery.ID)
		}

		// Print the stream content (the optimized query output)
		for content := range stream {
			fmt.Print(content)
		}

		log.Println("Workflow completed successfully.")
	}
}

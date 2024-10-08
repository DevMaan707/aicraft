package main

// import (
// 	"flag"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/DevMaan707/aicraft"
// 	"github.com/joho/godotenv"
// )

// var verbose bool

// func main() {

// 	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
// 	flag.Parse()

// 	cwd, _ := os.Getwd()
// 	fmt.Println("Current Working Directory:", cwd)
// 	err := godotenv.Load(cwd + "/.env")
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}

// 	apiKey := os.Getenv("OPENAI_API_KEY")
// 	if apiKey == "" {
// 		log.Fatalf("Error: OPENAI_API_KEY not found in environment variables")
// 	}
// 	if verbose {
// 		fmt.Println("API KEY => " + apiKey)
// 	}

// 	manager := aicraft.NewManager()
// 	pdf1 := "https://www.nber.org/system/files/working_papers/w27392/w27392.pdf"
// 	//pdf2 := "https://noobsverse.blr1.digitaloceanspaces.com/upload/files/2024/05/6iG6o3X319WqcTSfPnUa_26_5186d700361f07c13bdab2a3a6d8c679_file.pdf"

// 	workflowConfig := aicraft.WorkflowConfig{
// 		Tasks: []aicraft.TaskConfig{

// 			{
// 				ID:     "task_extract_text",
// 				Name:   "Extract Text from PDF",
// 				ToolID: aicraft.PDFExtractorTool.ID,
// 				Inputs: map[string]interface{}{
// 					"pdf_url": pdf1,
// 					"verbose": verbose,
// 				},
// 			},

// 			{
// 				ID:     "task_convert_pdf",
// 				Name:   "Convert PDF to Embeddings",
// 				ToolID: aicraft.PDFToEmbeddingsTool.ID,
// 				Inputs: map[string]interface{}{
// 					"chunkSize":    800,
// 					"chunkOverlap": 100,
// 					"api_key":      apiKey,
// 					"verbose":      verbose,
// 				},
// 			},

// 			{
// 				ID:     "task_query_embedding",
// 				Name:   "Convert Query to Embedding",
// 				ToolID: aicraft.QueryToEmbeddingTool.ID,
// 				Inputs: map[string]interface{}{
// 					"query":   "Give me the summary of the context provided.",
// 					"api_key": apiKey,
// 					"verbose": verbose,
// 				},
// 			},
// 			{
// 				ID:     "get_results",
// 				Name:   "Get Results",
// 				ToolID: aicraft.OpenAIContentGeneratorTool.ID,
// 				Inputs: map[string]interface{}{
// 					"query":        "Give me the summary of the context provided",
// 					"chunkSize":    800,
// 					"chunkOverlap": 100,
// 					"api_key":      apiKey,
// 					"verbose":      verbose,
// 				},
// 			},

// 			{
// 				ID:     "task_generate_image",
// 				Name:   "Generate Image from Text",
// 				ToolID: aicraft.ImageGeneratorTool.ID,
// 				Inputs: map[string]interface{}{
// 					"description": "A futuristic city skyline at sunset.",
// 					"api_key":     apiKey,
// 					"verbose":     verbose,
// 				},
// 			},
// 		},
// 		Agents: []aicraft.AgentConfig{
// 			{
// 				ID:   "agent1",
// 				Name: "Text Extractor",
// 				Tasks: []string{
// 					"task_extract_text",
// 				},
// 			},
// 			{
// 				ID:        "agent2",
// 				Name:      "PDF Processor",
// 				DependsOn: []string{"agent1"},
// 				Tasks: []string{
// 					"task_convert_pdf",
// 				},
// 			},
// 			{
// 				ID:        "agent3",
// 				Name:      "Query Embedding",
// 				DependsOn: []string{"agent2"},
// 				Tasks: []string{
// 					"task_query_embedding",
// 				},
// 			},
// 			{
// 				ID:        "agent4",
// 				Name:      "Image Generator",
// 				DependsOn: []string{"agent3"},
// 				Tasks: []string{
// 					"get_results",
// 				},
// 			},
// 			{
// 				ID:        "agent5",
// 				Name:      "Image Generator",
// 				DependsOn: []string{"agent3"},
// 				Tasks: []string{
// 					"task_generate_image",
// 				},
// 			},
// 		},
// 	}

// 	log.Println("Step 1: Initializing tasks and agents...")
// 	err = manager.InitializeWorkflow(workflowConfig)
// 	if err != nil {
// 		log.Fatalf("Error during workflow initialization: %v", err)
// 	}

// 	log.Println("Step 2: Executing all tasks and workflows...")
// 	err = manager.ExecuteAllWorkflows()

// 	if err != nil {
// 		log.Fatalf("Workflow completed with errors: %v", err)
// 	} else {

// 		log.Println("Step 3: Performing similarity search between query and PDF content embeddings...")

// 		pdfEmbeddings := manager.Agents["agent2"].Output["task_convert_pdf"].([][]float64)
// 		queryEmbedding := manager.Agents["agent3"].Output["task_query_embedding"].([]float64)

// 		mostSimilarChunkIndex := aicraft.FindMostSimilarChunk(queryEmbedding, pdfEmbeddings)
// 		log.Printf("Most Similar Chunk Index: %d\n", mostSimilarChunkIndex)

// 		extractedText := manager.Agents["agent1"].Output["task_extract_text"].(string)
// 		relevantText := aicraft.ExtractRelevantText(extractedText, mostSimilarChunkIndex, 800)

// 		agent := manager.Agents["agent4"]
// 		results := manager.Tasks["get_results"]
// 		stream := agent.Stream
// 		if stream == nil {
// 			log.Fatalf("Error: No stream found for task %s", results.ID)
// 		}
// 		for content := range stream {
// 			fmt.Print(content)
// 		}

// 		log.Printf("Relevant text for image generation: %s\n", relevantText)

// 		manager.Tasks["task_generate_image"].Inputs["description"] = relevantText

// 		// imageAgent := manager.Agents["agent5"]
// 		// imageTask := manager.Tasks["task_generate_image"]
// 		// imageURL, ok := imageAgent.Output[imageTask.ID].(string)
// 		// if !ok {
// 		// 	log.Fatalf("Error: Image URL is invalid or not found.")
// 		// }

// 		// fmt.Printf("Generated Image URL: %s\n", imageURL)
// 		// log.Println("Workflow completed successfully.")
// 	}
// }

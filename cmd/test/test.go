package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DevMaan707/aicraft"
	"github.com/joho/godotenv"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("Current Working Directory:", cwd)
	err := godotenv.Load(cwd + "/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	fmt.Println("API KEY => " + apiKey)

	manager := aicraft.NewManager()

	log.Println("Step 1: Creating task to extract text from PDF...")
	taskExtractText := manager.CreateTask("task_extract_text", "Extract Text from PDF", aicraft.PDFExtractorTool.ID, map[string]interface{}{
		"pdf_url": "https://noobsverse.blr1.digitaloceanspaces.com/upload/files/2024/05/6iG6o3X319WqcTSfPnUa_26_5186d700361f07c13bdab2a3a6d8c679_file.pdf",
	})
	if taskExtractText == nil {
		log.Fatalf("Error: Failed to create task for extracting text from PDF.")
	}
	log.Println("Created Task 1 = Extract Text from PDF")

	agentTextExtractor := manager.CreateAgent("agent1", "Text Extractor", nil)
	manager.AssignTaskToAgent(agentTextExtractor.ID, taskExtractText.ID)

	log.Println("Step 2: Executing the text extraction task...")
	err = manager.ExecuteWorkflow()
	if err != nil {
		log.Fatalf("Error during workflow execution: %v", err)
	}
	log.Println("Text extraction task executed successfully.")

	extractedText, ok := manager.Agents[agentTextExtractor.ID].Output[taskExtractText.ID].(string)
	if !ok || extractedText == "" {
		log.Fatalf("Error: Extracted text is empty or invalid.")
	}

	log.Println("Step 3: Creating task to convert extracted text to embeddings...")
	taskConvertPDF := manager.CreateTask("task_convert_pdf", "Convert PDF to Embeddings", aicraft.PDFToEmbeddingsTool.ID, map[string]interface{}{
		"pdf_content":  extractedText,
		"chunkSize":    800,
		"chunkOverlap": 100,
		"api_key":      apiKey,
	})
	if taskConvertPDF == nil {
		log.Fatalf("Error: Failed to create task for converting PDF to embeddings.")
	}
	log.Println("Created Task 2 = Convert PDF to Embeddings")

	agentPDFProcessor := manager.CreateAgent("agent2", "PDF Processor", []string{"agent1"})
	manager.AssignTaskToAgent(agentPDFProcessor.ID, taskConvertPDF.ID)

	log.Println("Step 4: Executing the PDF to Embeddings conversion task...")
	err = manager.ExecuteWorkflow()
	if err != nil {
		log.Fatalf("Error during workflow execution: %v", err)
	}
	log.Println("PDF to Embeddings conversion task executed successfully.")

	docEmbeddings, ok := manager.Agents[agentPDFProcessor.ID].Output[taskConvertPDF.ID].([][]float64)
	if !ok || len(docEmbeddings) == 0 {
		log.Fatalf("Error: Document embeddings are empty or invalid.")
	}

	log.Println("Step 5: Creating task to convert user query to embeddings...")
	taskQueryEmbedding := manager.CreateTask("task_query_embedding", "Convert Query to Embedding", aicraft.QueryToEmbeddingTool.ID, map[string]interface{}{
		"query":   "What does the context contain?",
		"api_key": apiKey,
	})
	if taskQueryEmbedding == nil {
		log.Fatalf("Error: Failed to create task for converting query to embedding.")
	}
	log.Println("Created Task 3 = Convert Query to Embedding")

	agentQueryEmbedding := manager.CreateAgent("agent3", "Query Embedding", []string{"agent2"})
	manager.AssignTaskToAgent(agentQueryEmbedding.ID, taskQueryEmbedding.ID)

	log.Println("Step 6: Executing the query embedding task...")
	err = manager.ExecuteWorkflow()
	if err != nil {
		log.Fatalf("Error during workflow execution: %v", err)
	}
	log.Println("Query embedding task executed successfully.")

	queryEmbedding, ok := manager.Agents[agentQueryEmbedding.ID].Output[taskQueryEmbedding.ID].([]float64)
	if !ok || len(queryEmbedding) == 0 {
		log.Fatalf("Error: Query embedding is empty or invalid.")
	}
	log.Println("Retrieved Query Embedding")

	log.Println("Step 7: Performing similarity search to find the most relevant text chunk...")
	mostSimilarChunkIndex := aicraft.FindMostSimilarChunk(queryEmbedding, docEmbeddings)
	log.Printf("Most Similar Chunk Index: %d\n", mostSimilarChunkIndex)

	relevantText := aicraft.ExtractRelevantText(extractedText, mostSimilarChunkIndex, 800)

	log.Println("Step 8: Creating task to optimize user query with context...")
	taskOptimizeQuery := manager.CreateTask("task_optimize_query", "Optimize Query", aicraft.OpenAIContentGeneratorTool.ID, map[string]interface{}{
		"query":        "Give me the summary of the context provided in 500 words.",
		"context":      relevantText,
		"chunkSize":    800,
		"chunkOverlap": 100,
		"api_key":      apiKey,
	})
	if taskOptimizeQuery == nil {
		log.Fatalf("Error: Failed to create task for optimizing query.")
	}
	log.Println("Created Task 4 = Optimize Query with Context")

	agentQueryOptimizer := manager.CreateAgent("agent4", "Query Optimizer", []string{"agent3"})
	manager.AssignTaskToAgent(agentQueryOptimizer.ID, taskOptimizeQuery.ID)

	log.Println("Step 9: Executing the query task...")
	err = manager.ExecuteWorkflow()
	if err != nil {
		log.Fatalf("Error during workflow execution: %v", err)
	}
	log.Println("Query optimization with context task executed successfully.")

	stream := manager.Agents[agentQueryOptimizer.ID].Stream
	if !ok {
		log.Fatalf("Error: No stream found for task %s", taskOptimizeQuery.ID)
	}

	for content := range stream {
		fmt.Print(content)
	}

}

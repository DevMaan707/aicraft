package aicraft

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"net/http"
	"os"

	"github.com/ledongthuc/pdf"
)

const maxTokens = 8000

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Tool struct {
	ID      string
	Name    string
	Execute func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error)
}

var (
	TextToPDFTool = &Tool{
		ID:   "text_to_pdf",
		Name: "Text to PDF",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			text, ok := inputs["text"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'text' is required and must be a string")
			}
			fmt.Println("Converting text to PDF...")
			return fmt.Sprintf("PDF Content from: %s", text), nil, nil
		},
	}

	OpenAIContentGeneratorTool = &Tool{
		ID:   "openai_content_generator",
		Name: "OpenAI Content Generator",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			query, ok := inputs["query"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'query' is required and must be a string")
			}
			chunkSize, ok := inputs["chunkSize"].(int)
			if !ok {
				return nil, nil, fmt.Errorf("input 'chunkSize' is required and must be an int")
			}
			chunkOverlap, ok := inputs["chunkOverlap"].(int)
			if !ok {
				return nil, nil, fmt.Errorf("input 'chunkOverlap' is required and must be an int")
			}

			context, ok := inputs["context"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'context' is required and must be a string")
			}

			apiKey, ok := inputs["api_key"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'api_key' is required and must be a string")
			}

			// Retrieve the verbose flag
			verbose, _ := inputs["verbose"].(bool)

			model := "gpt-3.5-turbo"
			if m, ok := inputs["model"].(string); ok && m != "" {
				model = m
			}

			contextChunks := SplitTextIntoChunks(context, chunkSize, chunkOverlap)

			prompt := fmt.Sprintf("Context: %s\n\nQuery: %s", contextChunks[0], query)
			if EstimateTokens(prompt) > maxTokens {
				prompt = truncateTextToTokenLimit(prompt, maxTokens-500)
			}

			if verbose {
				log.Printf("Generated prompt: %s", prompt)
			}

			data := map[string]interface{}{
				"model":  model,
				"stream": true,
				"messages": []map[string]string{
					{"role": "user", "content": prompt},
				},
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal request data: %v", err)
			}

			req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to execute request: %v", err)
			}

			contentChannel := make(chan interface{})

			go func() {
				defer resp.Body.Close()
				defer close(contentChannel)

				reader := bufio.NewReader(resp.Body)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Printf("failed to read stream: %v", err)
						break
					}

					if strings.TrimSpace(line) == "data: [DONE]" {
						break
					}

					if strings.HasPrefix(line, "data: ") {
						line = line[len("data: "):]
					}

					var streamResponse struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}

					if err := json.Unmarshal([]byte(line), &streamResponse); err != nil {
						continue
					}

					if len(streamResponse.Choices) > 0 {
						content := streamResponse.Choices[0].Delta.Content

						contentChannel <- content
					}
				}
			}()

			return nil, contentChannel, nil
		},
	}

	ImageGeneratorTool = &Tool{
		ID:   "image_generator",
		Name: "Image Generator",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			description, ok := inputs["description"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'description' is required and must be a string")
			}

			apiKey, ok := inputs["api_key"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'api_key' is required and must be a string")
			}

			verbose, _ := inputs["verbose"].(bool)

			data := map[string]interface{}{
				"prompt": description,
				"n":      1,
				"size":   "1024x1024",
			}

			if verbose {
				log.Printf("Generating image with description: %s", description)
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal request data: %v", err)
			}

			req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			var response struct {
				Data []struct {
					URL string `json:"url"`
				} `json:"data"`
			}
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode response: %v", err)
			}

			if len(response.Data) == 0 {
				return nil, nil, fmt.Errorf("no images returned from OpenAI API")
			}

			if verbose {
				log.Printf("Generated image URL: %s", response.Data[0].URL)
			}

			return response.Data[0].URL, nil, nil
		},
	}

	QueryToEmbeddingTool = &Tool{
		ID:   "query_to_embedding",
		Name: "Query to Embedding",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			query, ok := inputs["query"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'query' is required and must be a string")
			}

			apiKey, ok := inputs["api_key"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'api_key' is required and must be a string")
			}

			// Retrieve the verbose flag
			verbose, _ := inputs["verbose"].(bool)

			model := "text-embedding-ada-002"
			if m, ok := inputs["model"].(string); ok && m != "" {
				model = m
			}

			data := map[string]interface{}{
				"model": model,
				"input": query,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal request data: %v", err)
			}

			req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			if verbose {
				log.Printf("Sending query to OpenAI Embedding API: %s", query)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			var response struct {
				Data []struct {
					Embedding []float64 `json:"embedding"`
				} `json:"data"`
			}

			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode response: %v", err)
			}

			if len(response.Data) == 0 {
				return nil, nil, fmt.Errorf("no embeddings returned from OpenAI API")
			}

			if verbose {
				log.Printf("Query embedding generated successfully, embedding length: %d", len(response.Data[0].Embedding))
			}

			return response.Data[0].Embedding, nil, nil
		},
	}

	PDFToEmbeddingsTool = &Tool{
		ID:   "pdf_to_embeddings",
		Name: "PDF to Embeddings",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			pdfContent, _ := inputs["pdf_content"].(string)
			log.Println("PDF CONTENT LENGTH => " + fmt.Sprintf("%d", len(pdfContent)))
			chunkSize, _ := inputs["chunkSize"].(int)
			chunkOverlap, ok := inputs["chunkOverlap"].(int)
			if !ok {
				return nil, nil, fmt.Errorf("input 'chunkOverlap' is required and must be an int")
			}

			apiKey, ok := inputs["api_key"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'api_key' is required and must be a string")
			}

			verbose, _ := inputs["verbose"].(bool)

			chunks := SplitTextIntoChunks(pdfContent, chunkSize, chunkOverlap)
			var allEmbeddings [][]float64

			for i, chunk := range chunks {
				if verbose {
					log.Printf("Processing chunk %d: %s", i+1, chunk)
				}

				if len(chunk) == 0 {
					if verbose {
						log.Printf("Chunk %d is empty, skipping.", i+1)
					}
					continue
				}

				data := map[string]interface{}{
					"model": "text-embedding-ada-002",
					"input": chunk,
				}

				jsonData, err := json.Marshal(data)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to marshal request data: %v", err)
				}

				req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create request: %v", err)
				}

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+apiKey)

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to execute request: %v", err)
				}
				defer resp.Body.Close()

				var response struct {
					Data []struct {
						Embedding []float64 `json:"embedding"`
					} `json:"data"`
				}
				err = json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to decode response: %v", err)
				}

				if len(response.Data) == 0 {
					if verbose {
						log.Printf("Failed to produce embeddings for the provided text chunk: %s", chunk)
					}
					return nil, nil, fmt.Errorf("no embeddings returned from OpenAI API")
				}

				allEmbeddings = append(allEmbeddings, response.Data[0].Embedding)
				if verbose {
					log.Printf("Chunk %d processed successfully, embedding length: %d", i+1, len(response.Data[0].Embedding))
				}
			}

			if len(allEmbeddings) == 0 {
				return nil, nil, fmt.Errorf("no valid embeddings were produced")
			}

			if verbose {
				log.Printf("Total embeddings generated: %d", len(allEmbeddings))
			}

			return allEmbeddings, nil, nil
		},
	}

	PDFExtractorTool = &Tool{
		ID:   "pdf_extractor",
		Name: "PDF Extractor",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			pdfURL, ok := inputs["pdf_url"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'pdf_url' is required and must be a string")
			}

			verbose, _ := inputs["verbose"].(bool)

			pdfFilePath, err := DownloadPDF(pdfURL)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to download PDF: %v", err)
			}
			defer os.Remove(pdfFilePath)

			text, err := ExtractTextFromPDF(pdfFilePath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to extract text from PDF: %v", err)
			}

			if verbose {
				log.Println("Extracted text from PDF:", text)
			}

			return text, nil, nil
		},
	}

	ImageNeedCheckerTool = &Tool{
		ID:   "image_need_checker",
		Name: "Image Need Checker",
		Execute: func(inputs map[string]interface{}) (interface{}, <-chan interface{}, error) {
			content, ok := inputs["content"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'content' is required and must be a string")
			}

			apiKey, ok := inputs["api_key"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("input 'api_key' is required and must be a string")
			}
			model := "text-embedding-ada-002"
			if m, ok := inputs["model"].(string); ok && m != "" {
				model = m
			}
			data := map[string]interface{}{
				"model": model,
				"messages": []map[string]string{
					{"role": "system", "content": "You are an assistant that identifies the need for diagrams or flowcharts in text."},
					{"role": "user", "content": fmt.Sprintf("Given the following content, identify if any diagrams or flowcharts are needed and provide descriptions: %s. THE OUTPUT TO BE STRICTLY A SLICE OF STRINGS OR ARRAY OF STRINGS", content)},
				},
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal request data: %v", err)
			}

			req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			var response OpenAIResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode response: %v", err)
			}

			if len(response.Choices) == 0 {
				return nil, nil, fmt.Errorf("no choices returned from OpenAI API")
			}

			return response.Choices[0].Message.Content, nil, nil
		},
	}
)

func CosineSimilarity(vec1, vec2 []float64) float64 {
	var dotProduct, magA, magB float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		magA += vec1[i] * vec1[i]
		magB += vec2[i] * vec2[i]
	}
	return dotProduct / (math.Sqrt(magA) * math.Sqrt(magB))
}

func FindMostSimilarChunk(queryEmbedding []float64, docEmbeddings [][]float64) int {
	bestIndex := 0
	bestSimilarity := -1.0
	for i, docEmbedding := range docEmbeddings {
		similarity := CosineSimilarity(queryEmbedding, docEmbedding)
		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestIndex = i
		}
	}
	return bestIndex
}

// const (
// 	chunkSize    = 800
// 	chunkOverlap = 100
// )

func EstimateTokens(text string) int {

	return len(text) / 4
}

func SplitTextIntoChunks(text string, chunkSize int, chunkOverlap int) []string {
	words := strings.Fields(text)
	var chunks []string
	start := 0

	for start < len(words) {
		end := start + chunkSize

		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[start:end], " ")

		for EstimateTokens(chunk) > chunkSize {
			end--
			chunk = strings.Join(words[start:end], " ")
		}

		chunks = append(chunks, chunk)

		start += chunkSize - chunkOverlap

		if start >= len(words) {
			break
		}
	}

	return chunks
}

func truncateTextToTokenLimit(text string, maxTokens int) string {
	words := strings.Fields(text)
	if len(words) > maxTokens {
		return strings.Join(words[:maxTokens], " ")
	}
	return text
}

func ExtractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
		}

		buf.WriteString(pageText)

		truncatedText := truncateTextToTokenLimit(buf.String(), maxTokens)
		if len(truncatedText) < buf.Len() {
			buf.Reset()
			buf.WriteString(truncatedText)
			break
		}
	}

	return buf.String(), nil
}

func DownloadPDF(pdfURL string) (string, error) {

	tmpFile, err := os.CreateTemp("", "downloaded-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	resp, err := http.Get(pdfURL)
	if err != nil {
		return "", fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download PDF: status code %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF to temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

func ExtractDescriptions(content string) []string {

	lines := strings.Split(content, "\n")
	var descriptions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !strings.HasPrefix(line, "No images needed") {
			descriptions = append(descriptions, line)
		}
	}
	return descriptions
}

func ExtractRelevantText(extractedText string, index int, chunkSize int) string {
	words := strings.Fields(extractedText)
	start := index * chunkSize
	end := start + chunkSize

	if end > len(words) {
		end = len(words)
	}

	return strings.Join(words[start:end], " ")
}
func FlattenAndConvertToFloat32(embeddings [][]float64) ([]float32, []int) {
	var flattened []float32
	var lengths []int

	for _, vec := range embeddings {
		lengths = append(lengths, len(vec))
		for _, val := range vec {
			flattened = append(flattened, float32(val))
		}
	}

	return flattened, lengths
}

func ReconstructToFloat64(flattened []float32, lengths []int) [][]float64 {
	var embeddings [][]float64
	offset := 0

	for _, length := range lengths {
		vec := make([]float64, length)
		for j := 0; j < length; j++ {
			vec[j] = float64(flattened[offset+j])
		}
		embeddings = append(embeddings, vec)
		offset += length
	}

	return embeddings
}

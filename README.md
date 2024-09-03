# AICraft: Workflow Automation with AI

AICraft is a Go package that automates workflows using predefined tools, such as OpenAI's GPT-4 for content generation, DALL·E for image generation, and `text-embedding-ada-002` for generating embeddings from PDF content. This package allows you to create agents, assign tasks, and execute complex workflows involving AI-driven tools.

## **Features**

- **PDF to Embeddings Conversion:** Convert PDF content into embeddings using OpenAI's `text-embedding-ada-002`.
- **Content Generation:** Generate or optimize content using OpenAI's GPT-4.
- **Image Generation:** Create images or diagrams using OpenAI's DALL·E.
- **Workflow Automation:** Chain tasks together into workflows with dependency management.

## **Installation**

To install the package, simply run:

```bash
go get github.com/DevMaan707/aicraft
```
## Tools

- **PDFToEmbeddingsTool**: Converts PDF content into embeddings using OpenAI's text-embedding-ada-002.

- **OpenAIContentGeneratorTool**: Generates or optimizes content using OpenAI's GPT-4.

- **ImageGeneratorTool**: Generates images or diagrams using OpenAI's DALL·E.

- **TextToPDFTool**: Converts generated text and images into a PDF.

## Agents

Agents are responsible for executing tasks. Each agent can depend on other agents, ensuring tasks are executed in the correct order.
## Documentation
#### **Key Components**

1. **Manager:** 
   - Responsible for creating agents, tasks, and tools. 
   - Manages the execution of workflows and ensures tasks are executed in the correct order.

2. **Agent:**
   - An entity that executes one or more tasks.
   - Can have dependencies on other agents to ensure proper execution order.

3. **Task:**
   - Represents a unit of work that uses a specific tool.
   - Takes input parameters and produces output after execution.

4. **Tool:**
   - A predefined function or API call used to perform specific tasks, such as content generation or image creation.

#### **Key Functions**

1. **NewManager():** Creates a new `Manager` instance.

2. **CreateAgent(id, name string, dependsOn []string):** Creates a new `Agent` with a unique ID, name, and optional dependencies.

3. **CreateTask(id, name, toolID string, inputs map[string]interface{}):** Creates a new `Task` with a unique ID, name, and associated tool.

4. **AssignTaskToAgent(agentID, taskID string):** Assigns a task to an agent.

5. **ExecuteWorkflow():** Executes all agents and tasks within the manager, respecting dependencies.

#### **Predefined Tools**

1. **PDFToEmbeddingsTool:**
   - Converts PDF content into embeddings using OpenAI's `text-embedding-ada-002` model.
   - Inputs: `pdf_content` (string), `api_key` (string).

2. **OpenAIContentGeneratorTool:**
   - Generates or optimizes content using OpenAI's GPT-4 model.
   - Inputs: `query` (string), `api_key` (string).

3. **ImageGeneratorTool:**
   - Generates images or diagrams using OpenAI's DALL·E model.
   - Inputs: `description` (string), `api_key` (string).

4. **TextToPDFTool:**
   - Converts text into a PDF document.
   - Inputs: `text` (string).

#### **Example Workflow**

An example workflow can be set up to convert a PDF into embeddings, optimize a query, generate related images, and compile everything into a final PDF document.

Refer to the `README.md` for a step-by-step example of how to create and execute such a workflow.

#### **Advanced Features**

- **Dependency Management:** Agents can depend on other agents, allowing for complex workflows where tasks are executed in a specific order.
- **Error Handling:** The `ExecuteWorkflow` function ensures that errors are properly handled and reported during execution.

#### **Extending AICraft**

Users can extend the `aicraft` package by defining their own tools and tasks. This allows for greater flexibility and customization of workflows.

#### **Conclusion**

The `aicraft` package provides a powerful and flexible way to automate complex workflows involving AI-driven tasks. By leveraging predefined tools and the ability to define dependencies between agents, users can create sophisticated processes that handle everything from text generation to PDF creation.


## License
This project is licensed under the MIT License. See the LICENSE file for more details.

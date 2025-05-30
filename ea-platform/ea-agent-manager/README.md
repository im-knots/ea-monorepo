# Ea Agent Manager API

The Ea Platform Agent Manager API manages the creation of AI agents via an API that powers a Node-based Agent Builder UI. An Agent is a collection of Nodes and Edges that define a workflow to accomplish a specific task. This document outlines how nodes are categorized, how edges link them, and how our schema distinguishes between Node Definitions (“templates”) and Agent Nodes (“instances”).

## Terminology

- **Nodes**: A unit of work, such as a prompt, an LLM, or a component of an Agent.
- **Edge**: A connection between Nodes that defines the workflow.
- **Agent**: A collection of Nodes that execute a task or set of tasks.

## Nodes

Nodes are grouped into general types depending on what they do. 

| Type              | Use                                              | 
|-------------------|--------------------------------------------------|
| **Trigger**       | Initiates an agent workflow. (ROADMAP)                    | 
| **Input**         | Provides input data to the workflow.             | 
| **Worker**        | Processes data or performs specific actions.     | 
| **Utils**         | Utility nodes for various small tasks in a workflow | 
| **Destination**   | Outputs results to internal/external systems or storage.  |


### Trigger Nodes (ROADMAP)

| Name                           | Use                                                               |
|--------------------------------|-------------------------------------------------------------------|
| **trigger.internal.timed**              | Triggers an agent workflow at a set time. Uses cron syntax in JSON. |
| **trigger.internal.manual**             | Triggers an agent workflow manually.                              |
| **trigger.internal.loop.do**            | Repeats an agent workflow in a `do` loop.                        |
| **trigger.internal.loop.for**           | Repeats an agent workflow in a `for` loop with conditions.       |
| **trigger.external.slack**   | Triggers an agent workflow via a mention in a slack channel.          |
| **trigger.external.irc**   | Triggers an agent workflow via a mention in an irc channel.          |
| **trigger.external.webhook**   | Triggers an agent workflow via an external webhook URL.          |
| **trigger.external.aws.TODO**  | TODO: Define AWS-specific triggers.                             |
| **trigger.external.gcp.TODO**  | TODO: Define GCP-specific triggers.                             |
| **trigger.external.azure.TODO**| TODO: Define Azure-specific triggers.                           |
| **trigger.external.digitalocean.TODO** | TODO: Define Digital Ocean-specific triggers.             |
| **trigger.external.alibaba.TODO**     | TODO: Define Alibaba Cloud-specific triggers.             |


### Input Nodes

| Name            | Use                                              |
|-----------------|--------------------------------------------------|
| **input.internal.text**  | Accepts a text input for the agent workflow. Can be used to pass prompts or other textual data. |
| **input.external.image** | (ROADMAP) Accepts an image input for the agent workflow. Used for tasks like image generation or classification. |
| **input.external.video** | (ROADMAP) Accepts a video input for the agent workflow. Used for tasks like video processing or analysis. |
| **input.external.audio** | (ROADMAP) Accepts an audio input for the agent workflow. Used for tasks like transcription or audio analysis. |
| **input.external.model** | (ROADMAP) Accepts a model as input for the agent workflow. Used for training or fine tuning workflows. Accepts .safetensors. TODO other formats |
| **input.external.file** | (ROADMAP) Accepts an abritrary file input for the agent workflow. Used for more general tasks such as importing data via csv or tfrecord files|
| **input.external.github** | Accepts a github repo URL and takes the contents of a github repo as input. Used for coding tasks. Private repos require github API key setup under user profile |
| **input.external.jira** | (ROADMAP) Accepts jira stories as input. Usually used in combination with input.external.github and triggers to do coding tasks. Requires Jira API key setup under user profile |
| **input.external.web** | (ROADMAP) Accepts an arbitrary public webpage as input. |



### Worker Nodes

| Name                         | Use                                              |
|-----------------------------|--------------------------------------------------|
| **worker.inference.llm.ollama**       | Uses an LLM powered by Ollama for tasks like generating text or extracting tags. |
| **worker.inference.llm.openai**      | Uses OpenAI's models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.llm.anthropic**   | Uses Anthropic's models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.llm.google**   | Uses Google's Gemini models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.llm.xai**   | Uses xAI's models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.stable-diffusion.video** | (ROADMAP) Leverages Stable Diffusion to generate videos from text prompts. Supports model-specific settings. |
| **worker.inference.stable-diffusion.image** | (ROADMAP) Leverages Stable Diffusion to generate images from text prompts. Supports model-specific settings. |
| **worker.train**            | (ROADMAP) Executes a model training operation |
| **worker.finetune**            | (ROADMAP) Executes a model tuning operation |
| **worker.custom**           | (ROADMAP) Executes custom scripts or AI models for specific use cases. Requires user-provided code or configuration. |

### Utility Nodes

| Name                              | Use                                              |
|----------------------------------|--------------------------------------------------|
| **utils.internal.base64.encode** | Takes data and encodes it to base64 |
| **utils.internal.base64.decode** | Takes base64 encoded data and decodes it |
| **utils.external.github.pr**     | Creates a github pull request |



### Destination Nodes

| Name                              | Use                                              |
|----------------------------------|--------------------------------------------------|
| **destination.external.social.instagram** | (ROADMAP) Posts content (e.g., videos, images, text) to Instagram. |
| **destination.external.social.facebook** | (ROADMAP) Posts content (e.g., videos, images, text) to Facebook. |
| **destination.external.social.x** | (ROADMAP) Posts content (e.g., videos, images, text) to x.com. |
| **destination.external.social.reddit** | (ROADMAP) Posts content (e.g., videos, images, text) to Reddit. |
| **destination.external.social.linkedin** | (ROADMAP) Posts content (e.g., videos, images, text) to Linkedin. |
| **destination.external.social.pinterest** | (ROADMAP) Posts content (e.g., videos, images, text) to Pinterest. |
| **destination.external.social.tiktok** | (ROADMAP) Posts content (e.g., videos, images, text) to TikTok. |
| **destination.external.cloud._**    | (ROADMAP) Stores output files in a cloud storage solution (e.g., S3, GCS). |
| **destination.external.github**  | Posts output to a github repository as a git commit
| **destination.external.webhook** | (ROADMAP) Sends output to an external system via a webhook. |
| **destination.internal.ea**       | (ROADMAP) Stores output files within the Ea platform's storage system. |
| **destination.internal.text**       | Shows text output in a text box in the workflow, used for debugging or intermediate checks |
| **destination.internal.image**       | (ROADMAP) Shows image output in an image box in the workflow, used for debugging or intermediate checks |
| **destination.internal.video**       | (ROADMAP) Shows video output in a video box in the workflow, used for debugging or intermediate checks |



## Edges
Edges connect nodes and define the workflow's data flow and execution order.

| Property   | Description                             | Examples                  |
|------------|-----------------------------------------|--------------------------|
| **from**   | alias of the source node.                 | `"from": ["input"]`   |
| **to**     | aliad of the destination node(s).       | `"to": ["ollama"]` |


## Schema and Data Model
To keep the workflow flexible yet maintainable, we separate a Node’s definition from its instance in an Agent:

### Node Definition (the “template”)
-   Stored in a dedicated Mongo collection (e.g., nodeDefs).
-   Defines how to call an API or perform a function (base URL, method, headers, enumerated parameters, etc.).
-   Includes documentation metadata (description, tags, references).


### Agent (the “instance”)
-   References Node Definitions via a definition_ref.
-   Only overrides or provides values for the parameters needed.
-   Stores a graph of Node Instances (nodes) and Edges (edges) that define the workflow.

## API Documentation

### Endpoints Overview

| Method | Path                     | Description                                |
|--------|--------------------------|--------------------------------------------|
| **GET**   | `/api/v1/nodes`          | Retrieve all nodes with their `id` |
| **GET**   | `/api/v1/nodes/{id}`     | Retrieve a specific node by its `id`.      |
| **POST**  | `/api/v1/nodes`          | Create a new node definition.             |
| **PUT**   | `/api/v1/nodes/{id}`     | Update a specific node definition by its `id`. |
| **DELETE** | `/api/v1/nodes/{id}`    | Delete a specific node definition by its `id`. |
| **GET**   | `/api/v1/agents`         | Retrieve all agents with their `id` |
| **GET**   | `/api/v1/agents/{id}`    | Retrieve a specific agent by its `id`.    |
| **POST**  | `/api/v1/agents`         | Create a new agent.                       |
| **PUT**   | `/api/v1/agents/{id}`    | Update a specific agent by its `id`.      |
| **DELETE** | `/api/v1/agents/{id}`   | Delete a specific agent by its `id`.      |

---

### Required Headers
All requests to this API coming into the cluster via the api gateway must include an authorization header containing an authenticated user's JWT

```
Authorization: Bearer <YOUR JWT>
```

Internal systems within the cluster (behind Kong) can access this service by providing

(**Note: network level access is restricted in the cluster via NetworkPolicies**)

```
x-consumer-username: internal
```


---

### Nodes

#### `POST /api/v1/nodes`
Create a new node definition.

**Request Body Example:**
```json
{
  "type": "worker.inference.llm.ollama",
  "name": "Ollama LLM Inference",
  "creator": "<UUID OF CREATOR USER>",
  "api": {
    "base_url": "https://ollama.ea-platform.svc.cluster.local:11434",
    "endpoint": "/api/generate",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    }
  },
  "parameters": [
    {
      "key": "model",
      "type": "string",
      "description": "Name of the model to use, e.g. 'llama2-7b'.",
      "enum": ["llama3.2", "deepseek-r1:8b"],
      "default": "llama3.2"
    },
    {
      "key": "prompt",
      "type": "string",
      "description": "User prompt to be sent to the model.",
      "default": "Hello world"
    },
    {
      "key": "stream",
      "type": "bool",
      "description": "enable the full stream response, we have to disable this",
      "default": false
    },
    {
      "key": "temperature",
      "type": "number",
      "description": "Controls randomness in generation (0.0 - 1.0).",
      "default": 0.7
    }
  ],
  "outputs": [
    {
      "key": "textoutput",
      "type": "string",
      "description": "the result of the prompt",
      "enum": ["some", "promptoutput"],
      "default": "someoutput"
    }
  ],
  "metadata": {
    "description": "Makes an inference call to an Ollama instance for text generation.",
    "tags": ["worker", "llm", "ollama", "inference"],
    "additional": {
      "documentation_url": "https://github.com/ollama/ollama/blob/main/docs/api.md",
      "timeout": 30
    }
  }
}
```

---

#### `GET /api/v1/nodes`
Retrieve a list of all nodes. Use the `creator_id` query parameter to filter by creator.

**Request Example:**
- All nodes: `/api/v1/nodes`
- Nodes by creator: `/api/v1/nodes?creator_id=<SOME CREATOR UUID>`

**Response Example:**
```json
[
    {
        "creator":"<SOME CREATOR UUID>",
        "id":"<SOME NODE UUID>",
        "type":"worker.inference.llm.ollama"
    },
    {
        "creator":"<SOME CREATOR UUID>",
        "id":"<SOME NODE UUID>",
        "type":"worker.inference.llm.openai"
    }
]
```
---

#### `GET /api/v1/nodes/{id}`
Retrieve a specific node definition by its `id`.

**Response Example:**
```json
{
    "id":"c6520f08-ea04-4899-aeab-672cc01ff500",
    "name":"Ollama LLM Inference",
    "creator":"<SOME CREATOR UUID>",
    "type":"worker.inference.llm.ollama",
    "api":{
        "baseurl":"https://ollama.ea-platform.svc.cluster.local:11434",
        "endpoint":"/api/generate",
        "headers":{
            "Content-Type":"application/json"
        },
        "method":"POST"
    },
    "metadata":{
        "additional":null,
        "createdat":"2025-02-04T17:15:57.804Z",
        "description":"",
        "tags":null,
        "updatedat":"2025-02-04T17:15:57.804Z"
    },
    "parameters":[
        {
            "default":"llama3.2",
            "description":"Name of the model to use, e.g. 'llama2-7b'.",
            "enum":["llama3.2","deepseek-r1:8b"],
            "key":"model",
            "type":"string"},
        {
            "default":"Hello world",
            "description":"User prompt to be sent to the model.",
            "enum":null,
            "key":"prompt",
            "type":"text"
        },
        {
            "default":0.7,
            "description":"Controls randomness in generation (0.0 - 1.0).",
            "enum":null,
            "key":"temperature",
            "type":"number"
        }
    ],
    "outputs":[
        {
            "default":"someoutput",
            "description":"the result of the prompt",
            "enum":["some","promptoutput"],
            "key":"textoutput",
            "type":"string"
        }
    ]
}
```

---


**Response Example:**
```json
{
    "node_id":"9fb7ef94-9aba-4c8c-b085-f17b008ab9ed",
    "creator":"<SOME CREATOR UUID>",
    "message":"Node definition created"  
}
```



#### `PUT /api/v1/nodes/{id}`
Update an existing node definition.

**Request Body Example:**
```json
{
  "id": "c6520f08-ea04-4899-aeab-672cc01ff500",
  "type": "worker.inference.llm.ollama",
  "name": "Updated Ollama LLM Inference",
  "creator": "<UUID OF CREATOR USER>",
  "api": {
    "base_url": "https://ollama.ea-platform.svc.cluster.local:11434",
    "endpoint": "/api/generate",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    }
  },
  "parameters": [
    {
      "key": "model",
      "type": "string",
      "description": "Updated model selection",
      "enum": ["llama3.2", "deepseek-r1:8b"],
      "default": "deepseek-r1:8b"
    }
  ],
  "metadata": {
    "description": "Updated description",
    "tags": ["worker", "llm", "update"]
  }
}
```
---
**Response Example (Success):**
```json
{
  "message": "Node definition updated successfully",
  "node_id": "c6520f08-ea04-4899-aeab-672cc01ff500"
}
```




#### `DELETE /api/v1/nodes/{id}`
Delete a specific node definition by its `id`.

**Response Example (Success):**
```json
{
    "message": "Node definition deleted successfully",
    "node_id": "c6520f08-ea04-4899-aeab-672cc01ff500"
}
```

**Response Example (Not Found):**
```json
{
    "error": "Node definition not found"
}
```

### Agents

#### `POST /api/v1/agents`
Create a new agent.

**Request Body Example:**
```json
{
  "name": "My Sample Ollama Agent",
  "creator": "<UUID OF CREATOR USER FROM EA-AINU-MANAGER>",
  "description": "An example agent using the Ollama LLM definition.",
  "nodes": [
    {
      "type": "worker.inference.llm.ollama",
      "alias": "ollama",
      "parameters": {
        "model": "llama2-13b",
        "prompt": "Tell me a short story about a flying cat."
      }
    },
    {
      "type": "destination.internal.text",
      "alias": "textbox",
      "parameters": {}
    }
  ],
  "edges": [
    { "from": ["ollama"],"to": ["textbox"] }
  ]
}
```

**Response Example:**
```json
{
    "agent_id":"cac871c8-5f72-4e6c-9bc8-9eb006597d31",
    "creator":"<SOME CREATOR UUID>",
    "message":"Agent created"
}
```

---


#### `GET /api/v1/agents`
Retrieve a list of all agents. Use the `creator_id` query parameter to filter by creator.

**Request Example:**
- All agents: `/api/v1/agents`
- Agents by creator: `/api/v1/agents?creator_id=<SOME CREATOR UUID>`

**Response Example:**
```json
[
    {
        "creator": "marco@erulabs.ai",
        "id": "34ef1000-d6d0-44a6-ac37-3937d42ce0e2",
        "name": "My Sample Ollama Agent"
    },
    {
        "creator": "someuser@example.com",
        "id": "00000000-0000-0000-0000-000000000000",
        "name": "agent 2"
    }
]
```

---

#### `GET /api/v1/agents/{id}`
Retrieve a specific agent by its `id`.

**Response Example:**
```json
{
    "id":"25b218a7-b260-4212-9b3f-62b9ecfd43f6",
    "name":"My Sample Ollama Agent",
    "creator":"marco@erulabs.ai",
    "description":"An example agent using the Ollama LLM definition.",  
    "nodes":[
        {
            "alias": "ollama",
            "type":"worker.inference.llm.ollama",
            "parameters":{
                "model":"llama2-13b",
                "prompt":"Tell me a short story about a flying cat."
            }
        },
        {
            "alias": "textbox",
            "type":"destination.internal.text",
            "parameters":{
              "input": "{{ollama.response}}"
            }
        }
    ],
    "edges":[
        {"from":["ollama"],"to":["textbox"]}
    ],
    "metadata":{
        "createdat":"2025-02-04T17:56:27.169Z",
        "updatedat":"2025-02-04T17:56:27.169Z"
    }
}
```

---

#### `PUT /api/v1/agents/{id}`
Update an existing agent.

**Request Body Example:**
```json
{
  "id": "cac871c8-5f72-4e6c-9bc8-9eb006597d31",
  "name": "Updated Agent Name",
  "creator": "<UUID OF CREATOR USER>",
  "description": "Updated description of the agent",
  "nodes": [
    {
      "type": "worker.inference.llm.ollama",
      "alias": "updated_ollama",
      "parameters": {
        "model": "llama3.2",
        "prompt": "Updated prompt for agent."
      }
    }
  ],
  "edges": [
    { "from": ["updated_ollama"], "to": ["textbox"] }
  ]
}
```

**Response Example (Success):**
```json
{
  "message": "Agent updated successfully",
  "agent_id": "cac871c8-5f72-4e6c-9bc8-9eb006597d31"
}
```
---

#### `DELETE /api/v1/agents/{id}`
Delete a specific agent by its `id`.

**Response Example (Success):**
```json
{
    "message": "Agent deleted successfully",
    "agent_id": "34ef1000-d6d0-44a6-ac37-3937d42ce0e2"
}
```

**Response Example (Not Found):**
```json
{
    "error": "Agent not found"
}
```
---


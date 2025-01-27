# Ea Agent Manager API

The Ea Platform Agent Manager API manages the creation of AI agents via an API that powers a Node-based Agent Builder UI. An Agent is a collection of Nodes and Edges that define a workflow to accomplish a specific task. This document outlines how nodes are categorized, how edges link them, and how our schema distinguishes between Node Definitions (“templates”) and Agent Nodes (“instances”).

## Terminology

- **Nodes**: A unit of work, such as a prompt, an LLM, or a component of an Agent.
- **Edge**: A connection between Nodes that defines the workflow.
- **Agent**: A collection of Nodes that execute a task or set of tasks.


## Nodes

Nodes are grouped into general types depending on what they do. 

| Type              | Use                                              | Examples                                                              |
|-------------------|--------------------------------------------------|-----------------------------------------------------------------------|
| **Trigger**       | Initiates an agent workflow.                     | `timed`, `manual`, `loop`                                             |
| **Input**         | Provides input data to the workflow.             | `text`, `image`, `video`, `audio`                                     |
| **Worker**        | Processes data or performs specific actions.     | `ollama`, `chatgpt`, `stable-diffusion`, custom scripts or models     |
| **Destination**   | Outputs results to external systems or storage.  | `cloud storage`, `Ea storage`, `social media`                        |


### Trigger Nodes

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
| **input.external.image** | Accepts an image input for the agent workflow. Used for tasks like image generation or classification. |
| **input.external.video** | Accepts a video input for the agent workflow. Used for tasks like video processing or analysis. |
| **input.external.audio** | Accepts an audio input for the agent workflow. Used for tasks like transcription or audio analysis. |
| **input.external.model** | Accepts a model as input for the agent workflow. Used for training or fine tuning workflows. Accepts .safetensors. TODO other formats |
| **input.external.file** | Accepts an abritrary file input for the agent workflow. Used for more general tasks such as importing data via csv or tfrecord files|
| **input.external.github** | Accepts a github repo URL and takes the contents of a github repo as input. Used for coding tasks. Private repos require github API key setup under user profile |
| **input.external.jira** | Accepts jira stories as input. Usually used in combination with input.external.github and triggers to do coding tasks. Requires Jira API key setup under user profile |
| **input.external.web** | Accepts an arbitrary public webpage as input. |



### Worker Nodes

| Name                         | Use                                              |
|-----------------------------|--------------------------------------------------|
| **worker.inference.llm.ollama**       | Uses an LLM powered by Ollama for tasks like generating text or extracting tags. |
| **worker.inference.llm.openai**      | Uses OpenAI's models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.llm.anthropic**   | Uses Anthropic's models to perform tasks like generating descriptions or answering questions. |
| **worker.inference.stable-diffusion.video** | Leverages Stable Diffusion to generate videos from text prompts. Supports model-specific settings. |
| **worker.inference.stable-diffusion.image** | Leverages Stable Diffusion to generate images from text prompts. Supports model-specific settings. |
| **worker.train**            | Executes a model training operation |
| **worker.finetune**            | Executes a model tuning operation |
| **worker.custom**           | Executes custom scripts or AI models for specific use cases. Requires user-provided code or configuration. |


### Destination Nodes

| Name                              | Use                                              |
|----------------------------------|--------------------------------------------------|
| **destination.external.social.instagram** | Posts content (e.g., videos, images, text) to Instagram. |
| **destination.external.social.facebook** | Posts content (e.g., videos, images, text) to Facebook. |
| **destination.external.social.x** | Posts content (e.g., videos, images, text) to x.com. |
| **destination.external.social.reddit** | Posts content (e.g., videos, images, text) to Reddit. |
| **destination.external.social.linkedin** | Posts content (e.g., videos, images, text) to Linkedin. |
| **destination.external.social.pinterest** | Posts content (e.g., videos, images, text) to Pinterest. |
| **destination.external.social.tiktok** | Posts content (e.g., videos, images, text) to TikTok. |
| **destination.external.cloud**    | Stores output files in a cloud storage solution (e.g., S3, GCS). |
| **destination.external.github**  | Posts output to a github repository
| **destination.external.webhook** | Sends output to an external system via a webhook. |
| **destination.internal.ea**       | Stores output files within the Ea platform's storage system. |
| **destination.internal.text**       | Shows text output in a text box in the workflow, used for debugging or intermediate checks |
| **destination.internal.image**       | Shows image output in an image box in the workflow, used for debugging or intermediate checks |
| **destination.internal.video**       | Shows video output in a video box in the workflow, used for debugging or intermediate checks |
| **destination.internal.log**       | Shows log messages of connected nodes, used for debugging |


## Edges
Edges connect nodes and define the workflow's data flow and execution order.

| Property   | Description                             | Examples                  |
|------------|-----------------------------------------|--------------------------|
| **from**   | ID(s) of the source node.                 | `"from": ["input.text"]`   |
| **to**     | ID(s) of the destination node(s).       | `"to": ["worker.train"]` |


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

## Reasoning
- Maintainability: Centralizing each node’s API logic (parameters, endpoints, etc.) in a single definition means that if the underlying API changes, you only update the Node Definition once—rather than in every Agent that uses it.
- Simplicity in Agent: Agent documents only store the minimal data—which node definition they reference, plus any parameter values. This keeps the agent’s JSON lightweight.
- Scalability: Multiple agents can reuse the same Node Definition. For example, “worker.inference.llm.ollama” can appear in hundreds of different Agent workflows without duplicating configuration.
- Consistency: By enumerating possible parameter values (enum) or providing defaults in the Node Definition, you guide users to valid settings. The Agent just overrides them if needed.
- Versatility: If you decide to add new Node Definitions (like “worker.inference.llm.openai”), you don’t need to alter the schema—just create a new Node Definition. Agents can reference it immediately.

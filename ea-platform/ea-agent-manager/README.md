# Ea Agent Manager API

Ea Platform Agent Manager API

Manages the creation of AI agents with APIs driving the Frontend Node-based Agent Builder.


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




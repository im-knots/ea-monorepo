curl -X POST http://localhost:8084/api/v1/agents \
-H "Content-Type: application/json" \
-d '{
    "name": "My First Agent",
    "user": "knots",
    "description": "A test agent",
    "nodes": [
        {
            "id": "input",
            "type": "input",
            "data": "promptInput",
            "provider": "ea",
            "model": "some-tokenizer"
        },
        {
            "id": "worker",
            "type": "llm",
            "data": "promptInput",
            "provider": "ollama",
            "model": "llama3.3"
        },
        {
            "id": "output",
            "type": "output",
            "data": "textOutput",
            "provider": "ea",
            "model": "some-model"
        }
    ],
    "edges": [
        { "from": "input", "to": "worker" },
        { "from": "worker", "to": "output" }
    ]
}'

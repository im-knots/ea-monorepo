# eru-labs-monorepo
A monorepo for all things eru labs

## Contents
- Eru Labs brand webpage front/backends 
    - Built with Bootstrap for frontend
    - Built with `go 1.22.2` for backend
- Ea platform front/backends
    - Built with Bootstrap for frontend
    - Built with `go 1.22.2` for backend
- Ainulindale Client software for various platforms

## Requirements
- Docker
- Docker Compose
- A Web Browser + curl/postman

## Run everything locally with docker compose
```bash
docker-compose up --build
```

clean up
```bash
docker-compose down --remove-orphans
docker system prune -f
```
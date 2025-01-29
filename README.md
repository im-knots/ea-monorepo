# eru-labs-monorepo
A monorepo for all things eru labs

## Contents
- Eru Labs brand webpage front/backends 
- Ea platform front/backends
- Ainulindale Client software for various platforms
- Terraform to deploy infrastructure and Eru Labs services to GCP
- Documentation diagrams managed as code for the whole of Eru Labs


## Run everything locally with minikube
Some services (ea-job-engine) require kubernetes specifically for their operator patterns. Therefore we need a local kubernetes cluster for development. Minikube is the best bet. 

### Requirements
- minikube
- helm
- kubectl
- docker

### Start up Eru Labs components locally with helm and minikube
```bash
minikube delete # Clean up previous minikube setups
minikube start --drivver=docker
minikube addons enable registry
eval $(minikube docker-env) # Tell docker to use the minikube registry

./minikube.sh start # builds and runs all apps in local minikube, sets up portforwarding for local development

./minikube.sh stop # delete all services from the cluster and cleans up portforwarding processes

```
### Adding new services
To add a new service to the startup script simple create a new directory ea-platform/app or brand/app. Add a Dockerfile and `chart` directory that contains the standard helm chart. The minikube.sh script will pick up the new app automatically. 

Optionally, you can add a portforward line in the minikube.sh script's `k8s_port_forward()` function using existing as the example. 

### Smoke tests
All new API services should have associates `tests/smoke` directories and simple smoke tests to either populate test data or verify API handler functionality. 


## Ea Platform Architecture
![Ea Platform Architecture](docs/diagrams/eru_labs.png)

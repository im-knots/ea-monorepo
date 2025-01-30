#!/bin/bash

# set -e

REPO_DIR=$(pwd)

# Base directories containing the app folders
BASE_DIRS=("ea-platform" "brand")

# Local registry
LOCAL_REGISTRY="localhost:5000"
VERSION=${2:-"latest"}

# Namespaces for Helm deployments
EA_NAMESPACE="ea-platform"
BRAND_NAMESPACE="eru-labs-brand"

build_and_push() {
    local app_path=$1
    local app_name=$(basename "$app_path")
    
    echo "Processing app: $app_name"
    
    echo "Building Docker image for $app_name..."
    docker build -t "$LOCAL_REGISTRY/$app_name:$VERSION" "$app_path"

    echo "Pushing Docker image for $app_name to local registry..."
    docker push "$LOCAL_REGISTRY/$app_name:$VERSION"

    echo "Completed build and push for $app_name"
}

deploy_stack_terraform() {
    cd infra/environments/local
    terraform init
    terraform apply -auto-approve
    cd $REPO_DIR
}

k8s_port_forward() {
    echo "Starting port-forwarding..."

    # Port-forward brand-frontend
    nohup kubectl port-forward deployment/brand-frontend-eru-labs-brand-frontend 8080:8080 --namespace $BRAND_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for brand-frontend on port 8080 started."

    # Port-forward brand-backend
    nohup kubectl port-forward deployment/brand-backend-eru-labs-brand-backend 8081:8080 --namespace $BRAND_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for brand-backend on port 8081 started."

    # Port-forward ea-frontend
    nohup kubectl port-forward deployment/ea-frontend 8082:8080 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-frontend on port 8082 started."

    # Port-forward ea-agent-manager
    nohup kubectl port-forward deployment/ea-agent-manager 8083:8080 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-agent-manager on port 8083 started."

    # Port-forward ea-job-engine
    nohup kubectl port-forward deployment/ea-job-engine 8084:8080 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-job-engine on port 8084 started."

    # Port-forward ea-ainu-engine
    nohup kubectl port-forward deployment/ea-ainu-manager 8085:8080 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-ainu-manager on port 8085 started."

    # Port-forward ea-platform mongodb
    nohup kubectl port-forward deployment/mongodb 8086:27017 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-platform mongodb on port 8086 started."

    # Port-forward eru-labs-brand mongodb
    nohup kubectl port-forward deployment/mongodb 8087:27017 --namespace $BRAND_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for eru-labs-brand mongodb on port 8087 started."

    # Port-forward Grafana
    nohup kubectl port-forward deployment/kps-grafana 3000:3000 --namespace monitoring >/dev/null 2>&1 &
    echo "Port-forwarding for Grafana on port 3000 started."

    # Port-forward Prometheus
    nohup kubectl port-forward pod/prometheus-kps-kube-prometheus-stack-prometheus-0  9090:9090 --namespace monitoring >/dev/null 2>&1 &
    echo "Port-forwarding for Prometheus on port 9090 started."
}

seed_test_data() {
    echo "Seeding test data with smoke test scripts"
    cd ea-platform/ea-agent-manager/tests
    ./smoke/create-agent.sh
    ./smoke/create-node.sh
    cd ../../ea-ainu-manager/tests
    ./smoke/post-user.sh
}

cleanup() {
    echo "Cleaning up all kubectl port-forward processes..."
    # Find all kubectl port-forward processes and kill them
    pkill -f "kubectl port-forward"
    echo "Port-forwarding processes stopped."
    pwd
    cd infra/environments/local
    terraform init
    terraform destroy -auto-approve
    cd $REPO_DIR
}


# Main Script
case "$1" in
    start)
        eval $(minikube docker-env)
        helm repo add bitnami https://charts.bitnami.com/bitnami
        helm repo update

        for dir in "${BASE_DIRS[@]}"; do
            if [[ ! -d "$dir" ]]; then
                echo "Base directory $dir does not exist. Skipping."
                continue
            fi

            # Process each app in the directory
            echo "Iterating through apps in $dir..."
            for app_path in "$dir"/*; do
                # Skip if it's not a directory
                if [[ -d "$app_path" ]]; then
                    build_and_push "$app_path"
                fi
            done
        done

        deploy_stack_terraform

        echo "Waiting some time for services to be ready before starting portforwarding"
        sleep 10
        k8s_port_forward
        echo "Waiting some more time for port forwards to be set up before seeding data"
        sleep 5
        seed_test_data


        echo "All apps processed and deployed successfully."
        ;;
    stop)
        cleanup
        ;;
    *)
        echo "Usage: $0 {start|stop} [version]"
        exit 1
        ;;
esac

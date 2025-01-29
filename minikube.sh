#!/bin/bash

set -e

# Base directories containing the app folders
BASE_DIRS=("ea-platform" "brand")

# Local registry
LOCAL_REGISTRY="localhost:5000"
VERSION=${2:-"latest"}

# Namespaces for Helm deployments
EA_NAMESPACE="ea-platform"
BRAND_NAMESPACE="eru-labs-brand"

# Array to store PIDs of background port-forward processes
PORT_FORWARD_PIDS=()

create_namespace() {
    local namespace=$1
    echo "Ensuring namespace $namespace exists..."
    if ! kubectl get namespace "$namespace" &>/dev/null; then
        kubectl create namespace "$namespace"
        echo "Namespace $namespace created."
    else
        echo "Namespace $namespace already exists."
    fi
}

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

deploy_app_helm_charts() {
    local app_path=$1
    local app_name=$(basename "$app_path")
    local chart_path="$app_path/chart"
    local namespace=$2

    echo "Deploying Helm chart for $app_name in namespace $namespace..."

    if [[ -d "$chart_path" ]]; then
        helm upgrade --install "$app_name" "$chart_path" \
            --namespace "$namespace" \
            --set image.repository="$LOCAL_REGISTRY/$app_name" \
            --set image.tag="$VERSION" \
            --create-namespace
        echo "Helm deployment for $app_name in namespace $namespace completed."
    else
        echo "No chart directory found for $app_name. Skipping Helm deployment."
    fi
}

deploy_components_helm_charts() {
    local namespace=$1

    if helm upgrade --install "mongodb" "bitnami/mongodb" \
        --namespace "$namespace" \
        --set auth.enabled="false"; then
        echo "MongoDB component installed successfully in namespace $namespace."
    else
        echo "Failed to install MongoDB component in namespace $namespace."
    fi
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
    echo "Port-forwarding for ea-agent-manager on port 8084 started."

    # Port-forward ea-job-engine
    nohup kubectl port-forward deployment/ea-job-engine 8084:8080 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-job-engine on port 8085 started."
}

seed_test_data() {
    echo "Seeding test data with smoke test scripts"
    cd ea-platform/ea-agent-manager/tests
    pwd
    ./smoke/create-agent.sh
    ./smoke/create-node.sh
}

cleanup() {
    echo "Cleaning up all kubectl port-forward processes..."
    # Find all kubectl port-forward processes and kill them
    pkill -f "kubectl port-forward"
    echo "Port-forwarding processes stopped."
}


delete_namespaces() {
    echo "Deleting namespaces $EA_NAMESPACE and $BRAND_NAMESPACE..."
    kubectl delete namespace "$EA_NAMESPACE" --ignore-not-found
    kubectl delete namespace "$BRAND_NAMESPACE" --ignore-not-found
    echo "Namespaces deleted successfully."
}

# Main Script
case "$1" in
    start)
        helm repo add bitnami https://charts.bitnami.com/bitnami
        helm repo update

        for dir in "${BASE_DIRS[@]}"; do
            if [[ ! -d "$dir" ]]; then
                echo "Base directory $dir does not exist. Skipping."
                continue
            fi

            # Set namespace based on directory
            namespace=$([ "$dir" == "ea-platform" ] && echo "$EA_NAMESPACE" || echo "$BRAND_NAMESPACE")

            # Ensure namespace exists
            create_namespace "$namespace"

            echo "Deploying component Helm charts in $namespace..."
            deploy_components_helm_charts "$namespace"

            # Process each app in the directory
            echo "Iterating through apps in $dir..."
            for app_path in "$dir"/*; do
                # Skip if it's not a directory
                if [[ -d "$app_path" ]]; then
                    build_and_push "$app_path"
                    deploy_app_helm_charts "$app_path" "$namespace"
                fi
            done
        done

        echo "Waiting some time for services to be ready before starting portforwarding"
        sleep 30
        k8s_port_forward
        echo "Waiting some more time for port forwards to be set up before seeding data"
        sleep 5
        seed_test_data


        echo "All apps processed and deployed successfully."
        ;;
    stop)
        cleanup
        delete_namespaces
        ;;
    *)
        echo "Usage: $0 {start|stop} [version]"
        exit 1
        ;;
esac

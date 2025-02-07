#!/bin/bash

# set -e

REPO_DIR=$(pwd)

# Colors for terminal output
BOLD_YELLOW="\e[1;33m"
RESET="\e[0m"


# Base directories containing the app folders
BASE_DIRS=("ea-platform" "brand")

# Local registry
LOCAL_REGISTRY="localhost:5000"
VERSION=${2:-"latest"}

# Namespaces for Helm deployments
EA_NAMESPACE="ea-platform"
BRAND_NAMESPACE="eru-labs-brand"

build_and_push() {
    echo -e "${BOLD_YELLOW}BUILDING LOCAL IMAGES${RESET}"
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
    echo -e "${BOLD_YELLOW}DEPLOYING APPS AND DEPENDENCIES TO MINIKUBE${RESET}"
    cd infra/environments/local
    terraform init
    terraform apply -auto-approve
    cd $REPO_DIR
}

k8s_ingress_dns() {
    MINIKUBE_IP=$(minikube ip)
    echo "I need sudo permissions to update your /etc/hosts file"

    HOST_ENTRIES=(
        "agent-manager.ea.erulabs.local"
        "ainu-manager.ea.erulabs.local"
        "ea.erulabs.local"
        "job-api.ea.erulabs.local"
        "backend.erulabs.local"
        "erulabs.local"
        "ollama.ea.erulabs.local"
        "grafana.erulabs.local"
        "prometheus.erulabs.local"
    )

    for HOSTNAME in "${HOST_ENTRIES[@]}"; do
        # Check if the exact IP and hostname pair exists
        if ! grep -q "^$MINIKUBE_IP[[:space:]]\+$HOSTNAME\$" /etc/hosts; then
            echo "$MINIKUBE_IP $HOSTNAME" | sudo tee -a /etc/hosts > /dev/null
            echo "Added $HOSTNAME to /etc/hosts"
        else
            echo "$HOSTNAME already exists in /etc/hosts, skipping..."
        fi
    done
}

remove_k8s_ingress_dns() {
    MINIKUBE_IP=$(minikube ip)
    echo "I need sudo permissions to remove entries from your /etc/hosts file"

    HOST_ENTRIES=(
        "agent-manager.ea.erulabs.local"
        "ainu-manager.ea.erulabs.local"
        "ea.erulabs.local"
        "job-api.ea.erulabs.local"
        "backend.erulabs.local"
        "erulabs.local"
        "ollama.ea.erulabs.local"
        "grafana.erulabs.local"
        "prometheus.erulabs.local"
    )

    for HOSTNAME in "${HOST_ENTRIES[@]}"; do
        # Check if the exact IP and hostname pair exists
        if grep -q "^$MINIKUBE_IP[[:space:]]\+$HOSTNAME\$" /etc/hosts; then
            # Remove the matching line
            sudo sed -i.bak "/^$MINIKUBE_IP[[:space:]]\+$HOSTNAME$/d" /etc/hosts
            echo "Removed $HOSTNAME from /etc/hosts"
        else
            echo "$HOSTNAME not found in /etc/hosts, skipping..."
        fi
    done
}



k8s_port_forward() {
    echo -e "${BOLD_YELLOW}STARTING PORTFORWARDS${RESET}"
    # Port-forward ea-platform mongodb
    nohup kubectl port-forward deployment/mongodb 8086:27017 --namespace $EA_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for ea-platform mongodb on port 8086 started."

    # Port-forward eru-labs-brand mongodb
    nohup kubectl port-forward deployment/mongodb 8087:27017 --namespace $BRAND_NAMESPACE >/dev/null 2>&1 &
    echo "Port-forwarding for eru-labs-brand mongodb on port 8087 started."
}

seed_test_data() {
    echo -e "${BOLD_YELLOW}SEEDING TEST DATA WITH SMOKE TEST SCRIPTS${RESET}"
    cd ea-platform/ea-ainu-manager/tests
    ./smoke/post-user.sh
    cd ../../ea-agent-manager/tests
    ./smoke/post-agent.sh
    ./smoke/post-node.sh
    
    cd $REPO_DIR
}

run_tests() {
    echo "Running ea-ainu-manager smoke tests"
    cd ea-platform/ea-ainu-manager/tests
    ./smoke/get-all-users.sh
    ./smoke/get-user.sh
    ./smoke/post-user-device.sh
    ./smoke/post-user-job.sh
    ./smoke/delete-user-device.sh
    ./smoke/delete-user-job.sh
    ./smoke/put-user-compute-credits.sh

    echo "Running ea-agent-manager smoke tests"
    cd ../../ea-agent-manager/tests
    ./smoke/get-all-agents.sh
    ./smoke/get-all-nodes.sh
    ./smoke/get-agent.sh
    ./smoke/get-node.sh

    echo "Running ea-job-api smoke tests"
    cd ../../ea-job-api/tests
    ./smoke/post-job.sh
    
    cd $REPO_DIR
}

cleanup() {
    echo "Cleaning up all kubectl port-forward processes..."
    # Find all kubectl port-forward processes and kill them
    pkill -f "kubectl port-forward"
    remove_k8s_ingress_dns
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
        minikube addons enable ingress
        # minikube addons enable ingress-dns // we could do this but hosts file is more universal than resolvconf
        minikube addons enable registry
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

        k8s_ingress_dns
        k8s_port_forward
        seed_test_data
        run_tests


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

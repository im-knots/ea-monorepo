#!/bin/bash

set -e

# Configuration
APP_NAME=$1
VERSION=${2:-"latest"}
PROJECT_ID=$(gcloud config get-value project)
REGION="us-central1"
ARTIFACT_REGISTRY_DOCKER="gcr.io/$PROJECT_ID/$APP_NAME"
ARTIFACT_REGISTRY_HELM="helm://$REGION-docker.pkg.dev/$PROJECT_ID/$APP_NAME-helm"

# Functions
build_docker_image() {
    echo "Building Docker image for $APP_NAME..."
    docker build -t "$ARTIFACT_REGISTRY_DOCKER:$VERSION" "./$APP_NAME"
}

push_docker_image() {
    echo "Pushing Docker image to GCP Artifact Registry..."
    docker push "$ARTIFACT_REGISTRY_DOCKER:$VERSION"
}

# Main Script
if [[ -z "$APP_NAME" ]]; then
    echo "Usage: $0 <app-name> [version]"
    exit 1
fi

echo "Starting build for $APP_NAME with version $VERSION..."

build_docker_image
push_docker_image


echo "Build and deployment completed for $APP_NAME with version $VERSION."

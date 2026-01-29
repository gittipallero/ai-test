#!/bin/bash
set -e

# Variables
RESOURCE_GROUP="rg-pacman-dev"
LOCATION="westeurope"
DEPLOYMENT_NAME="pacman-deployment"

# Login to Azure if not logged in (uncomment if needed, but assuming user is logged in or will use 'az login')
# az login

# Create Resource Group
echo "Creating Resource Group $RESOURCE_GROUP in $LOCATION..."
az group create --name $RESOURCE_GROUP --location $LOCATION

# Deploy Infrastructure
echo "Deploying Infrastructure..."
az deployment group create \
  --name $DEPLOYMENT_NAME \
  --resource-group $RESOURCE_GROUP \
  --template-file infra/main.bicep

# Get Outputs
ACR_LOGIN_SERVER=$(az deployment group show --resource-group $RESOURCE_GROUP --name $DEPLOYMENT_NAME --query properties.outputs.acrLoginServer.value -o tsv)
WEB_APP_NAME=$(az deployment group show --resource-group $RESOURCE_GROUP --name $DEPLOYMENT_NAME --query properties.outputs.webAppUrl.value -o tsv | awk -F/ '{print $3}' | awk -F. '{print $1}')

echo "ACR Login Server: $ACR_LOGIN_SERVER"
echo "Web App Name: $WEB_APP_NAME"

# Login to ACR
echo "Logging into ACR..."
az acr login --name $(echo $ACR_LOGIN_SERVER | cut -d. -f1)

# Build and Push Docker Image
IMAGE_NAME="$ACR_LOGIN_SERVER/pacman:latest"
echo "Building Docker image $IMAGE_NAME..."
docker build -t $IMAGE_NAME .

echo "Pushing Docker image to ACR..."
docker push $IMAGE_NAME

# Restart Web App to pick up new image
echo "Restarting Web App..."
az webapp restart --name $WEB_APP_NAME --resource-group $RESOURCE_GROUP

echo "Deployment Complete! App URL: https://$WEB_APP_NAME.azurewebsites.net"

#!/bin/bash

# Get ECR URL
ECR_URL=$(aws ecr describe-repositories \
  --repository-names hello-service\
  --region us-west-2 \
  --query 'repositories[0].repositoryUri' \
  --output text)

echo "ECR URL: $ECR_URL"

# Login to ECR
ECR_BASE=$(echo $ECR_URL | cut -d'/' -f1)
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $ECR_BASE

# Build and push splitter
echo "Building and pushing splitter..."
cd splitter
docker buildx build \
  --platform linux/amd64 \
  --push \
  -t $ECR_URL:splitter .
cd ..

# Build and push mapper
echo "Building and pushing mapper..."
cd mapper
docker buildx build \
  --platform linux/amd64 \
  --push \
  -t $ECR_URL:mapper .
cd ..

# Build and push reducer
echo "Building and pushing reducer..."
cd reducer
docker buildx build \
  --platform linux/amd64 \
  --push \
  -t $ECR_URL:reducer .
cd ..

# Verify uploads
echo "Verifying uploads..."
aws ecr list-images --repository-name hello-service --region us-west-2 --query 'imageIds[*].imageTag' --output table

echo "Done! Images pushed to ECR with tags:"
echo "  - $ECR_URL:splitter"
echo "  - $ECR_URL:mapper"
echo "  - $ECR_URL:reducer"
echo ""
echo "Next steps:"
echo "1. Create ECS task definitions using these image URIs"
echo "2. Run 5 tasks total (1 splitter, 3 mappers, 1 reducer)"
echo "3. Get their public IPs and update test.sh"
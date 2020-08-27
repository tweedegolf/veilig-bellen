#!/usr/bin/env bash

REPO_NAME="tweedegolf/veilig-bellen-backend"
FULL_TAG="$REPO_NAME:$DOCKER_TAG"
LATEST_TAG="$REPO_NAME:latest"

docker load < ./build/backend/backend-image.tar.gz
echo "Tagging veilig-bellen-backend image: $FULL_TAG"
docker tag tweedegolf/veilig-bellen-backend $FULL_TAG
docker tag $FULL_TAG $LATEST_TAG
echo "Pushing $FULL_TAG"
docker push $FULL_TAG
docker push $LATEST_TAG

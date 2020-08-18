#!/usr/bin/env bash

FULL_TAG="tweedegolf/veilig-bellen-backend:$DOCKER_TAG"

docker load < ./build/backend/backend-image.tar.gz
echo "Tagging veilig-bellen-backend image: $FULL_TAG"
docker tag tweedegolf/veilig-bellen-backend $FULL_TAG
echo "Pushing $FULL_TAG"
docker push $FULL_TAGsudo apt-get update
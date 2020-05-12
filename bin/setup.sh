#!/usr/bin/env bash

export USER_ID="$(id -u)"
export GROUP_ID="$(id -g)"

docker-compose -f docker-compose.yml -f docker-compose.setup.yml run --no-deps --workdir /go/src/app  backend go mod vendor
docker-compose -f docker-compose.yml -f docker-compose.setup.yml run --no-deps frontend_agents yarn
docker-compose -f docker-compose.yml -f docker-compose.setup.yml run --no-deps frontend_public yarn
docker-compose -f docker-compose.yml -f docker-compose.setup.yml run --no-deps frontend_panel yarn

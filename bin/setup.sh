#!/bin/bash

docker-compose run --no-deps --workdir="/go/src/app" backend go mod vendor

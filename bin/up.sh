#!/bin/bash

HOST_LAN_IP=$(hostname -I | awk '{print $1}') docker-compose up
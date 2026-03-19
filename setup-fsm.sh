#!/bin/bash
set -e

mkdir -p fsm/data/mods
mkdir -p fsm/data/config

cd fsm

curl -fsSL https://raw.githubusercontent.com/S0CT/fsm/refs/heads/main/docker-compose.yml -o docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/S0CT/fsm/refs/heads/main/config/fsm.ini -o fsm.ini
curl -fsSL https://raw.githubusercontent.com/S0CT/fsm/refs/heads/main/config/mod-list.json -o data/mods/mod-list.json

echo "Setup complete."
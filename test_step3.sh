#!/bin/bash

# Opens 4 terminals and start the manager process
current_dir=$(pwd)

commands=(
    "./nodeManager"
    "./nodeManager -config ./manager_process/new-node-config2"
    "./nodeManager -config ./manager_process/new-node-config3"
    "./nodeManager -config ./manager_process/new-node-config4"
    "./nodeManager -config ./manager_process/new-node-config5"
)

for cmd in "${commands[@]}"; do
    osascript -e "tell application \"Terminal\" to do script \"cd $current_dir; $cmd\""
done

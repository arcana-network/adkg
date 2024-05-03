go build -tags test -o adkgNode main.go

go build -o nodeManager ./manager_process

mkdir ./manager_process/new-node-config
mkdir ./manager_process/new-node-config2
mkdir ./manager_process/new-node-config3
mkdir ./manager_process/new-node-config4

# if the folder already exist, then remove all the previous config
rm ./manager_process/new-node-config/*
rm ./manager_process/new-node-config2/*
rm ./manager_process/new-node-config3/*
rm ./manager_process/new-node-config4/*

cp test-node-config/config.test.1.json ./manager_process/new-node-config
cp test-node-config/config.test.2.json ./manager_process/new-node-config2
cp test-node-config/config.test.3.json ./manager_process/new-node-config3
cp test-node-config/config.test.4.json ./manager_process/new-node-config4

# # Create session 1
# tmux new-session -d -s node1 ./nodeManager

# # Create session 2
# tmux new-session -d -s node2 ./nodeManager -config ./manager_process/new-node-config2

# # Create session 3
# tmux new-session -d -s node3 ./nodeManager -config ./manager_process/new-node-config3

# # Create session 4
# tmux new-session -d -s node4 ./nodeManager -config ./manager_process/new-node-config4

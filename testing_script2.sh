
rm ./manager_process/new-node-config/*
rm ./manager_process/new-node-config2/*
rm ./manager_process/new-node-config3/*
rm ./manager_process/new-node-config4/*

cp test-node-config/config.test.5.json ./manager_process/new-node-config
cp test-node-config/config.test.6.json ./manager_process/new-node-config2
cp test-node-config/config.test.7.json ./manager_process/new-node-config3
cp test-node-config/config.test.8.json ./manager_process/new-node-config4


python nodelist-pyscript/node_list.py -pc 1 2 1
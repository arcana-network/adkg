
# Change PSS status from epoch1 to epoch2 to 0:
python nodelist-pyscript/node_list.py -pc 1 2 0

#Change epoch back to epoch 1: 
python nodelist-pyscript/node_list.py -ec 1

rm nodeManager
rm adkgNode
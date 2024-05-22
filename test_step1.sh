#!/bin/bash

# Create a Python virtual environment
python -m venv ./venv

# Activate the virtual environment
source ./venv/bin/activate

# Install the required packages
pip install -r ./nodelist-pyscript/requirements.txt

# Check current PSS status from epoch 1 to epoch 2
pss_status=$(python nodelist-pyscript/node_list.py -p 1 2)

# Check if PSS status is not 0
if [[ $pss_status -ne 0 ]]; then
    python nodelist-pyscript/node_list.py -pc 1 2 0
fi

# Check current epoch
current_epoch=$(python nodelist-pyscript/node_list.py -e)

# Check if current epoch is not 1
if [[ $current_epoch -ne 1 ]]; then
    python nodelist-pyscript/node_list.py -ec 1
fi

# Deactivate the virtual environment
# deactivate

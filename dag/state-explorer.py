#!/usr/bin/env python
'''
@FileName    : state-explorer.py
@Description :
@Date        : 2022/12/17 19:31:51
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
from web3 import Web3

# Connect to the Ethereum blockchain
w3 = Web3(Web3.HTTPProvider("http://localhost:8545"))

# Set the contract address and ABI (Application Binary Interface)
contract_address = "0x..."
abi = [{...}]

# Instantiate the contract
contract = w3.eth.contract(address=contract_address, abi=abi)

# Call the contract's functions
owner = contract.functions.owner().call()
contract_address = contract.functions.contractAddress().call()
caller = contract.functions.caller().call()

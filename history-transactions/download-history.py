#!/usr/bin/env python
'''
@FileName    : download-history.py
@Description : Download history transactions to database.
@Date        : 2022/12/14 17:42:02
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import pickle
from web3 import Web3, HTTPProvider

# Instantiate a web3 remote provider
w3 = Web3(HTTPProvider('https://eth-mainnet.g.alchemy.com/v2/QL0Uqa-Z_QgMgBBKDFOUB8zoYx7q5Lg-'))

# Request the latest block number
ending_blocknumber = w3.eth.blockNumber

# Set the starting block number
starting_blocknumber = ending_blocknumber - 100 

# Filter through blocks and look for transactions involving this address
blockchain_address = ""

# Create an empty dictionary we will add transaction data to
tx_dictionary = {}

def getTransactions(start, end, address=""):
    """
    @Description  : Doanload transactions involving this address from start block to end block.
    @Params       :
        start: Start block
        end: End block
        address: Filter address, 
    @Returns      : None
    """
    print(f"Started filtering through block number {start} to {end} for transactions involving the address: {address}...")
    for x in range(start, end):
        block = w3.eth.getBlock(x, True)
        for transaction in block.transactions:
            if address == "":
                with open("transactions.txt", "wb") as f:
                    hashStr = transaction['hash'].hex()
                    tx_dictionary[hashStr] = transaction
                    pickle.dump(tx_dictionary, f)
                f.close()
            elif transaction['to'] == address or transaction['from'] == address:
                with open("transactions.txt", "wb") as f:
                    hashStr = transaction['hash'].hex()
                    tx_dictionary[hashStr] = transaction
                    pickle.dump(tx_dictionary, f)
                f.close()
    print(f"Finished searching blocks {start} through {end} and found {len(tx_dictionary)} transactions")

getTransactions(starting_blocknumber, ending_blocknumber, blockchain_address)
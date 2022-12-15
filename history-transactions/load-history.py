#!/usr/bin/env python
'''
@FileName    : load-history.py
@Description : Load history transactions.
@Date        : 2022/12/14 17:41:21
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import pickle

# Open the file with read only permit
with open('transactions.txt', 'rb') as f:
    # read the file as a binary data stream
    data = pickle.load(f)

# Do something with the data
print(data)

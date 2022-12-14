<img width="671" alt="image" src="https://user-images.githubusercontent.com/120435282/207215816-7a295046-689a-464d-9e0c-06d38cd0ddf9.png">

## About
State verification machine is a middleware for EVM, which could pre-planning and divide transactions into no conflicting threads.

## Why
1. No rollback, no re-execute: Existing parallel projects use optimistic parallel execution where transactions are executed in parallel and then rolled back if conflicts are found, and re-executed.
2. Pre planning with off-chain: Build pre-order transactions using on-chain computing resources is a waste.
3. PBS (Proposer-Builder Separation): Just like PBS, we separate the nodes for pre-planning and actually execution to prevent malicious behavior.

## Documentation

## Directory Structure
```
~~ Production ~~
├── dag: DAG functions.
├── example: All examples of this project, like DAG construction, DAG patition.
├── history-transactions: Download, update, load history of transactions from block A to block B.
├── node: Basic nodes functions.
├── p2p: Make sequencer p2p function available.
├── transactions-x-x.pkl: Some data of transactions.
├── pre-order-go: Make all transactions ordered before really excuted in Go.
├── pre-order-python: Make all transactions ordered before really excuted in Python.
└── README.md: Now you are looking at me.
```

## Example
![Figure_1](https://user-images.githubusercontent.com/120435282/207559258-09c9d07e-d68a-49fb-b3d8-579beecf209f.png)
![Figure_2](https://user-images.githubusercontent.com/120435282/207560699-d06ff8df-d704-48da-8fae-9c905d248f60.png)

## Reference
- Optimism: https://github.com/ethereum-optimism/optimism
- Block-STM - Scaling Blockchain Execution by Turning Ordering Curse to a Performance Blessing: https://arxiv.org/abs/2203.06871

## License

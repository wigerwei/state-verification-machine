<img width="671" alt="image" src="https://user-images.githubusercontent.com/120435282/207215816-7a295046-689a-464d-9e0c-06d38cd0ddf9.png">

## About
State verification machine is a middleware for EVM, which could pre-planning and divide transactions into no conflicting threads.

## Why
1. No rollback, no re-execute: Existing parallel projects use optimistic parallel execution where transactions are executed in parallel and then rolled back if conflicts are found, and re-executed.
2. Pre planning with off-chain: Build pre-order transactions using on-chain computing resources is a waste.
3. PBS (Proposer-Builder Separation): Just like PBS, we separate the nodes for pre-planning and actually execution to prevent malicious behavior.

## Documentation

## Directory Structure

## License

import hashlib
import struct

def main():
    # example:
    block_headers = [
            {"prev_block_hash":"0000000000000000000000000000000000000000000000000000000000000000", "content":"genesis block:A pay C 12.3 BTC"},
            {"prev_block_hash":"to_be_hashed", "content":"2nd block:C pay B 2.0 BTC"},
            {"prev_block_hash":"to_be_hashed", "content":"3th block:transactions..."},
            {"prev_block_hash":"to_be_hashed", "content":"4th block:transactions...j"},
            {"prev_block_hash":"to_be_hashed", "content":"5th block:transactions..."}
            ]

    # hash prev block header
    index = 0
    for header in block_headers:
        # genesis block, ignore
        if index == 0:
            print(header)
            index = index + 1
            continue

        # generate hash chain
        prev_block_header = block_headers[index - 1]
        target_buffer = prev_block_header["content"] + prev_block_header["prev_block_hash"]
        header["prev_block_hash"] = hashlib.sha256(target_buffer).hexdigest()
        print(header)
        index = index + 1

if __name__ == '__main__':
    main()
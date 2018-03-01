package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) printChain(nodeID string) {
	bc := NewBlockchain(nodeID)
	defer bc.db.Close()

	bci := bc.Iterator()

	/*
		pow.block.PrevBlockHash,
		pow.block.HashTransactions(),
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce))
	*/
	for {
		block := bci.Next()
		fmt.Printf("============= Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		fmt.Printf("Merkle Root: %x\n", block.HashTransactions())
		fmt.Printf("Timestamp: %x\n", IntToHex(block.Timestamp))
		fmt.Printf("Nonce: %x\n", IntToHex(int64(block.Nonce)))
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}

	bb, _ := bc.GetBlock(bc.tip)
	test1 := bb.HashTransactions()
	test2 := DeserializeBlock(bb.Serialize()).HashTransactions()
	fmt.Printf("test1: %x\n", test1)
	fmt.Printf("test2: %x\n", test2)

}

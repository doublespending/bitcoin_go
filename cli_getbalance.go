package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	balance, _ := UTXOSet.FindUTXO(pubKeyHash)

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

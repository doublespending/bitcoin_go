package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func BlockEqual(b1 *Block, b2 *Block) bool {
	return bytes.Compare(b1.Serialize(), b2.Serialize()) == 0
}

func isGenesisBlock(b *Block) bool {
	if b.Height == 0 && len(b.Transactions) == 1 && b.Transactions[0].IsCoinbase() && len(b.PrevBlockHash) == 0 {
		// disable pow check first
		//pow := NewProofOfWork(b)
		//return pow.Validate()
		return true
	}
	return false
}

func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block
	var buff bytes.Buffer

	buff.Write(d)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

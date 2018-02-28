package main

import (
	"bytes"
	"encoding/hex"
	"log"
	"math"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accmulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

	Work:
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) {
					accmulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					if accmulated > amount {
						break Work
					}
				}
			}

		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accmulated, unspentOutputs
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) (int, map[string][]int) {
	amount := math.MaxInt64
	return u.FindSpendableOutputs(pubKeyHash, amount)
}

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}

func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updateOuts := TXOutputs{}
					outsByte := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsByte)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs[outIdx] = out
						}
					}

					if len(updateOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			for outIdx, out := range tx.Vout {
				newOutputs.Outputs[outIdx] = out
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// pay attention to coinbase transaction
func (u UTXOSet) VerifyTransaction(target *Transaction) bool {
	if target.IsCoinbase() {
		return true
	}

	db := u.Blockchain.db
	result := true
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		count := make(map[int]int)
		inSum := 0
		outSum := 0

		for _, vin := range target.Vin {
			txOutsByte := b.Get(vin.Txid)

			if txOutsByte == nil { // 这样用没有问题？
				result = false
				goto End
			}

			txOuts := DeserializeOutputs(txOutsByte)

			if _, ok := txOuts.Outputs[vin.Vout]; !ok {
				result = false
				goto End
			}
			count[vin.Vout]++
			if count[vin.Vout] > 1 {
				result = false
				goto End
			}

			if !vin.UseKey(txOuts.Outputs[vin.Vout].PubKeyHash) {
				result = false
				goto End
			}

			inSum += txOuts.Outputs[vin.Vout].Value
		}

		for _, vout := range target.Vout {
			outSum += vout.Value
		}

		if outSum > inSum {
			result = false
			goto End
		}

	End:
		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	return result && target.Verify()
}

// pay attention to coinbase transaction
func (u UTXOSet) VerifyBlock(b *Block, testPow bool) bool {
	bc := u.Blockchain
	// test height and hash
	desHeight, desHash := bc.GetBestHeight()
	if bytes.Compare(b.PrevBlockHash, desHash) != 0 || b.Height != desHeight+1 {
		return false
	}

	// test pow
	if testPow {
		pow := NewProofOfWork(b)
		if pow.Validate() == false {
			return false
		}
	}

	// test transactions
	type UTXOID struct {
		TxId []byte
		Vout int
	}

	count := make(map[string]map[int]int)

	for _, tx := range b.Transactions {
		// test a transaction
		if u.VerifyTransaction(tx) == false {
			return false
		}

		// pay attention to coinbase transaction
		// test if use the same utxo in a block
		for _, vin := range tx.Vin {
			count[hex.EncodeToString(vin.Txid)][vin.Vout]++
			if count[hex.EncodeToString(vin.Txid)][vin.Vout] > 1 {
				return false
			}
		}

	}
	return true
}

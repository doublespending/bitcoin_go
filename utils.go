package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func StrMap2Slice(strmap map[string]int) []string {
	var result []string
	for key, _ := range strmap {
		result = append(result, key)
	}
	return result
}

func AddSlice2StrMap(strmap map[string]int, slice []string) map[string]int {
	for _, e := range slice {
		strmap[e]++
	}

	return strmap
}

func ElementInStrSlice(slice []string, target string) bool {
	for _, e := range slice {
		if target == e {
			return true
		}
	}
	return false
}

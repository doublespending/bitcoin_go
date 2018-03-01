package main

func InitNewWork() {
	// init knownNodes
	for _, addr := range fullNodes {
		knownNodes[addr]++
	}

	delete(knownNodes, nodeAddress)
}

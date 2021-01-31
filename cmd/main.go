package main

import (
	"fmt"

	"github.com/hirasawayuki/block_chain/utils"
)

func main() {
	neighbors := utils.FindNeighbors("127.0.0.1", 5000, 0, 3, 5000, 5003)
	fmt.Println(neighbors)
}

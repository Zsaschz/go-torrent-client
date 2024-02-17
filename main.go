package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(file)
	bencodeTorrent, err := DecodeTorrent(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", bencodeTorrent)
}

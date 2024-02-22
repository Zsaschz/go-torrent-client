package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	filename := os.Args[1]
	port, _ := strconv.ParseUint(os.Args[2], 10, 64)
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(file)
	bto := BencodeTorrent{}
	err = bencode.Unmarshal(r, &bto)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%+v", bto)
	//bencodeTorrent, err := ParseTorrent(r)
	torrent := bto.toTorrentFile()
	peerId := [20]byte{}
	_, err = rand.Read(peerId[:])
	trackerURL, _ := torrent.buildTrackerURL(peerId, uint16(port))

	response, err := http.Get(trackerURL)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	var tracker Tracker

	responseData, err := ioutil.ReadAll(response.Body)

	parsed, _, err := ParseBencode(string(responseData), 0)
	parsedMap, _ := parsed.(map[string]interface{})
	err = MapToStruct(parsedMap, &tracker)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", tracker)

	peers, _ := tracker.getPeers()
	for _, peer := range peers {
		fmt.Printf("%+v\n", peer)
		conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
		if err != nil {
			panic(err)
		}
	}
}

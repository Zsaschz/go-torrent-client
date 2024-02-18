package main

import (
	"encoding/binary"
	"net"
	"net/url"
	"strconv"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type Tracker struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Peer struct {
	IP   net.IP
	Port uint16
}

func (t *TorrentFile) buildTrackerURL(peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *Tracker) getPeers() ([]Peer, error) {
	// Peers is made out of group of six bytes: 4 bytes represent ip and 2 bytes represent port (Big Endian)
	peerSize := 6
	numOfPeers := len(t.Peers) / peerSize
	peersRaw := []byte(t.Peers)
	peers := make([]Peer, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		offset := i * peerSize
		peers[i].IP = peersRaw[offset : offset+4]
		peers[i].Port = binary.BigEndian.Uint16(peersRaw[offset+4 : offset+6])
	}

	return peers, nil
}

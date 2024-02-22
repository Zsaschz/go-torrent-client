package main

import (
	"encoding/binary"
	"io"
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

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerId   [20]byte
}

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnhoke        messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message

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

func (p *Peer) String() string {
	return p.IP.String() + ":" + strconv.Itoa(int(p.Port))
}

func (h *Handshake) Serialize() []byte {
	// pstr = BitTorrent protocol
	// <len_of_pstr>BitTorrent protocol<8 reserved bit>><info_hash><peer_id>
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	ptr := 1
	ptr += copy(buf[ptr:], h.Pstr)
	ptr += copy(buf[ptr:], make([]byte, 8))
	ptr += copy(buf[ptr:], h.InfoHash[:])
	ptr += copy(buf[ptr:], h.PeerId[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	var err error
	pstr := make([]byte, 19)
	infoHash := make([]byte, 20)
	peerId := make([]byte, 20)
	_, err = r.Read(pstr)
	if err != nil {
		return nil, err
	}
	_, err = r.Read(infoHash)
	if err != nil {
		return nil, err
	}
	_, err = r.Read(peerId)
	if err != nil {
		return nil, err
	}
	return &Handshake{
		Pstr:     string(pstr),
		InfoHash: [20]byte{},
		PeerId:   [20]byte{},
	}, nil
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := len(m.Payload) + 1
	buf := make([]byte, length+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(length))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	return nil
}

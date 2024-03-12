package peer

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"
	"torrent-client/torrentfile"
)

type Tracker struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type Peer struct {
	IP   net.IP
	Port uint16
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type TorrentWork struct {
	torrentfile.TorrentFile
	Tracker
}

type BitField []byte

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

func (bf BitField) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	return bf[byteIndex]>>(7-offset)&1 != 0
}

func (bf BitField) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	bf[byteIndex] |= 1 << (7 - offset)
}

func (t *TorrentWork) calculatePieceBounds(index int) (start int, end int) {
	start = index * t.PieceLength
	end = start + t.PieceLength
	if end > t.Length-1 {
		end = t.Length - 1
	}
	return start, end
}

func (t *TorrentWork) calculatePieceSize(index int) int {
	start, end := t.calculatePieceBounds(index)
	return end - start
}

func (t *TorrentWork) startDownloadWorker(peer Peer, workQueue chan *pieceWork, result chan *pieceResult) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		panic(err)
	}
	hs := Handshake{
		Pstr:     ProtocolIdentifier,
		InfoHash: [20]byte{},
		PeerId:   [20]byte{},
	}
	conn.Write(hs)
	return
}

func (t *TorrentWork) Download() {
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{
			index:  index,
			hash:   hash,
			length: length,
		}
	}

	// Start workers
	peers, err := t.getPeers()
	if err != nil {
		panic(err)
	}
	for _, peer := range peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculatePieceBounds(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
	}
	close(workQueue)
}

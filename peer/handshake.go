package peer

import "io"

const ProtocolIdentifier string = "BitTorrent protocol"

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerId   [20]byte
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

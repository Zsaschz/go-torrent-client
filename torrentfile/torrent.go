package torrentfile

import (
	"bufio"
	"github.com/jackpal/bencode-go"
	"net/url"
	"os"
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

func Open(path string) *TorrentFile {
	file, err := os.Open(path)
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
	return bto.toTorrentFile()
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

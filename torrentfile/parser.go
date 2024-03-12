package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"github.com/jackpal/bencode-go"
	"io"
	"strconv"
	"strings"
	"torrent-client/utils"
)

type torrentInfo struct {
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
	PiecesLength int    `bencode:"piece length"`
	Pieces       string `bencode:"pieces"`
	//rawInfo      string
}

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     torrentInfo `bencode:"info"`
}

// TODO: fix hash calculation
//func (bto *BencodeTorrent) infoHash() [20]byte {
//	encodeByte := []byte(bto.Info.rawInfo)
//	result, _, _ := parse(bto.Info.rawInfo, 0)
//	fmt.Println(bto.Info.rawInfo)
//	fmt.Println(result)
//	return sha1.Sum(encodeByte)
//}

func (bto *BencodeTorrent) toTorrentFile() *TorrentFile {
	var pieceHashes [][20]byte
	for i := 0; i < len(bto.Info.Pieces); i = i + 20 {
		byteArray := []byte(bto.Info.Pieces[i : i+20])
		var arr [20]byte
		copy(arr[:], byteArray)
		pieceHashes = append(pieceHashes, arr)
	}

	var buff bytes.Buffer
	_ = bencode.Marshal(&buff, bto.Info)
	infoHash := sha1.Sum(buff.Bytes())

	return &TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PiecesLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
}

func parseInt(text string, idx int) (int64, int, error) {
	endIdx := strings.Index(text[idx:], "e")
	value, err := strconv.ParseInt(text[idx:idx+endIdx], 10, 64)
	if err != nil {
		return value, idx + endIdx, err
	}
	return value, idx + endIdx + 1, nil
}

func parseStr(text string, idx int) (string, int, error) {
	var value string
	delim := strings.Index(text[idx:], ":")
	length, err := strconv.ParseInt(text[idx:idx+delim], 10, 64)
	if err != nil {
		return value, idx + delim, err
	}
	endIdx := idx + delim + 1 + int(length)
	value = text[idx+delim+1 : endIdx]
	return value, endIdx, nil
}

func parseList(text string, idx int) ([]interface{}, int, error) {
	data := make([]interface{}, 0)
	for rune(text[idx]) != 'e' {
		value, endIdx, err := ParseBencode(text, idx)
		if err != nil {
			return nil, endIdx, err
		}
		data = append(data, value)
		idx = endIdx
	}
	return data, idx, nil
}

func parseDict(text string, idx int) (map[string]interface{}, int, error) {
	data := make(map[string]interface{})
	for rune(text[idx]) != 'e' {
		value, endIdx, err := ParseBencode(text, idx)
		if err != nil {
			return nil, endIdx, err
		}
		key, ok := value.(string)
		if !ok {
			return nil, endIdx, err
		}
		value, endIdx, err = ParseBencode(text, endIdx)
		if err != nil {
			return data, endIdx, err
		}
		data[key] = value
		idx = endIdx
	}
	return data, idx, nil
}

func ParseBencode(text string, idx int) (interface{}, int, error) {
	var value interface{}
	var err error
	length := len([]rune(text))
	if length < 3 {
		return nil, 0, errors.New("not enough value to unpack")
	}

	if rune(text[idx]) == 'i' {
		value, idx, err = parseInt(text, idx+1)
	} else if rune(text[idx]) == 'l' {
		value, idx, err = parseList(text, idx+1)
	} else if rune(text[idx]) == 'd' {
		value, idx, err = parseDict(text, idx+1)
	} else {
		value, idx, err = parseStr(text, idx)
	}
	return value, idx, err
}

func getInfoSubstring(text string) string {
	idx := strings.Index(text, "4:info") + 6
	return text[idx : len(text)-1]
}

func ParseTorrent(r io.Reader) (data *BencodeTorrent, err error) {
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	parsed, _, err := ParseBencode(buf.String(), 0)
	if err != nil {
		return nil, err
	}

	parsedMap, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to decode torrent file")
	}

	announce := parsedMap["announce"].(string)
	info, ok := parsedMap["info"].(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to decode torrent file")
	}
	bInfo := torrentInfo{}
	err = utils.MapToStruct(info, bInfo)
	if err != nil {
		return nil, err
	}
	return &BencodeTorrent{
		Announce: announce,
		Info:     bInfo,
	}, nil
}

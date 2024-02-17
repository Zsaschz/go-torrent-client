package main

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

type bencodeInfo struct {
	Pieces       string `bencode:"pieces"`
	PiecesLength int    `bencode:"pieces length"`
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
}

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func (bto *BencodeTorrent) toTorrentFile() *TorrentFile {
	return nil
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
		value, endIdx, err := parse(text, idx)
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
		value, endIdx, err := parse(text, idx)
		if err != nil {
			return nil, endIdx, err
		}
		key, ok := value.(string)
		if !ok {
			return nil, endIdx, err
		}
		value, endIdx, err = parse(text, endIdx)
		if err != nil {
			return data, endIdx, err
		}
		data[key] = value
		idx = endIdx
	}
	return data, idx, nil
}

func parse(text string, idx int) (interface{}, int, error) {
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

func DecodeTorrent(r io.Reader) (data *BencodeTorrent, err error) {
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	parsed, _, err := parse(buf.String(), 0)
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

	pieces := info["pieces"].(string)
	length := info["length"].(int64)
	name := info["name"].(string)
	piecesLength := info["piece length"].(int64)
	return &BencodeTorrent{
		Announce: announce,
		Info: bencodeInfo{
			Pieces:       pieces,
			PiecesLength: int(piecesLength),
			Length:       int(length),
			Name:         name,
		},
	}, nil
}

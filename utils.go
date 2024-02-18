package main

import "github.com/mitchellh/mapstructure"

func MapToStruct(m map[string]interface{}, s interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  s,
		TagName: "bencode",
	})
	if err != nil {
		return err
	}
	return decoder.Decode(m)
}

// Copyright (c) 2025 Nikolaos Grammatikos and Oglofus Ltd (Company No. 14840351)
// Distributed under the Boost Software License, Version 1.0.
// See accompanying file LICENSE or copy at https://www.boost.org/LICENSE_1_0.txt

package main

import (
	"bytes"
	"crypto/sha3"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"slices"
)

const IndexLength uint64 = 44

type Bang struct {
	T string `json:"t"`
	U string `json:"u"`
}

type HashedValue struct {
	Key   []byte
	Value []byte
}

func main() {
	var err error

	var bangsFile *os.File
	if bangsFile, err = os.Open("bangs.json"); err != nil {
		log.Fatalf("Error opening bangs.json: %v", err)

		return
	}

	defer bangsFile.Close()

	var decoder = json.NewDecoder(bangsFile)

	var bangs []Bang
	if err = decoder.Decode(&bangs); err != nil {
		log.Fatalf("Error decoding bangs.json: %v", err)

		return
	}

	var hashedValues = make([]HashedValue, 0, len(bangs))

	var totalLength = 0
	var bang Bang
	for _, bang = range bangs {
		var hash = sha3.Sum224([]byte(bang.T))
		var data = []byte(bang.U)

		data = bytes.ReplaceAll(data, []byte("<q>"), []byte{0xC0})

		hashedValues = append(
			hashedValues, HashedValue{
				Key:   hash[:],
				Value: data,
			},
		)

		totalLength += len(data)
	}

	slices.SortFunc(
		hashedValues, func(a, b HashedValue) int {
			return bytes.Compare(a.Key, b.Key)
		},
	)

	var keysBuf = make([]byte, 0, IndexLength*uint64(len(bangs)))
	var valuesBuf = make([]byte, 0, totalLength)

	var offset = uint64(0)

	var h HashedValue
	for _, h = range hashedValues {
		valuesBuf = append(valuesBuf, h.Value...)

		var valueLength = uint64(len(h.Value))

		var keySlice = make([]byte, 0, IndexLength)

		keySlice = append(keySlice, h.Key...)

		keySlice = binary.LittleEndian.AppendUint64(keySlice, offset)
		keySlice = binary.LittleEndian.AppendUint64(keySlice, valueLength)

		keysBuf = append(keysBuf, keySlice...)

		offset = offset + valueLength
	}

	var idxFile *os.File
	if idxFile, err = os.Create("bangs.idx"); err != nil {
		log.Fatalln(err)

		return
	}

	defer idxFile.Close()

	if _, err = idxFile.Write(keysBuf); err != nil {
		log.Fatalln(err)

		return
	}

	var dataFile *os.File
	if dataFile, err = os.Create("bangs.dat"); err != nil {
		log.Fatalln(err)

		return
	}

	defer dataFile.Close()

	if _, err = dataFile.Write(valuesBuf); err != nil {
		log.Fatalln(err)

		return
	}
}

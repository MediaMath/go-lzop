package lzop

// Copyright 2016 MediaMath <http://www.mediamath.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"encoding/binary"
	"hash/adler32"
)

const (
	version           = 0x1030
	libVersion        = 0x2080
	versionForExtract = 0x0940
	method            = 2
	level             = 1
	flags             = 0x3000001
	fileMode          = 0x000081B4
)

var lzopMagic = []byte{0x89, 0x4c, 0x5a, 0x4f, 0x00, 0x0d, 0x0a, 0x1a, 0x0a}
var endBytes = []byte{0x00, 0x00, 0x00, 0x00}

func writeData(buff *bytes.Buffer, fileTime int64, fileName string, data []byte,
	compressionFunction func([]byte) []byte) error {

	err := binary.Write(buff, binary.BigEndian, uint16(version))
	err = binary.Write(buff, binary.BigEndian, uint16(libVersion))
	err = binary.Write(buff, binary.BigEndian, uint16(versionForExtract))
	err = binary.Write(buff, binary.BigEndian, uint8(method))
	err = binary.Write(buff, binary.BigEndian, uint8(level))
	err = binary.Write(buff, binary.BigEndian, uint32(flags))
	err = binary.Write(buff, binary.BigEndian, uint32(fileMode))
	err = binary.Write(buff, binary.BigEndian, uint32(fileTime))
	err = binary.Write(buff, binary.BigEndian, uint32(0)) //timeHigh
	err = binary.Write(buff, binary.BigEndian, uint8(len(fileName)))
	_, err = buff.Write([]byte(fileName))
	err = binary.Write(buff, binary.BigEndian, uint32(adler32.Checksum(buff.Bytes())))

	if err != nil {
		return err
	}

	blockSize := 256 * 1024

	if len(data) < blockSize {
		blockSize = len(data)
	}

	iterations := len(data) / blockSize
	for i := 0; i < iterations+1; i++ {

		var compressed []byte
		var unCompressed []byte

		leftOver := len(data) - (i * blockSize)
		if leftOver < blockSize {
			unCompressed = data[(i * blockSize):]
		} else {
			unCompressed = data[(i * blockSize):((i + 1) * blockSize)]
		}

		if len(unCompressed) == 0 {
			continue
		}

		compressed = compressionFunction(unCompressed)

		// did you actually compress anything?
		//this is to stop the compression library from sticking in extra
		//characters which from what I can tell just messes with things
		//and you end up with a corrupted lzop file
		if len(compressed) > len(unCompressed) {
			compressed = unCompressed
		}

		err = binary.Write(buff, binary.BigEndian, uint32(len(unCompressed)))
		err = binary.Write(buff, binary.BigEndian, uint32(len(compressed)))
		err = binary.Write(buff, binary.BigEndian, uint32(adler32.Checksum(unCompressed)))

		if err != nil {
			return err
		}

		buff.Write(compressed)
	}

	return err
}

//CompressData Will Compress your data expecting an LZO 1X1 compression.
func CompressData(fileTime int64, fileName string, data []byte,
	compressionFunction func([]byte) []byte) ([]byte, error) {

	buff := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	err := writeData(buff, fileTime, fileName, data, compressionFunction)

	if err != nil {
		return nil, err
	}

	end := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	err = binary.Write(end, binary.BigEndian, lzopMagic)
	end.Write(buff.Bytes())
	end.Write(endBytes)

	return end.Bytes(), err
}

//CompressDataWithBuffers Allows you to specify the buffer to use for compression and ending output
//this allows re-use
func CompressDataWithBuffers(compressBuffer, endBuffer *bytes.Buffer, fileTime int64, fileName string, data []byte,
	compressionFunction func([]byte) []byte) ([]byte, error) {

	compressBuffer.Reset()
	endBuffer.Reset()

	err := writeData(compressBuffer, fileTime, fileName, data, compressionFunction)

	if err != nil {
		return nil, err
	}

	err = binary.Write(endBuffer, binary.BigEndian, lzopMagic)
	endBuffer.Write(compressBuffer.Bytes())
	endBuffer.Write(endBytes)

	return endBuffer.Bytes(), err
}

//GetAddedBufferLength Returns you the added byte size of characters
//this is so you can pre-size the end buffer coming in and it wont resize
func GetAddedBufferLength() int {
	return len(lzopMagic) + len(endBytes)
}

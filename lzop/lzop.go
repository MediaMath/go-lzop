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

//WriteHeader Writes only the header start of an lzop file
func WriteHeader(buff *bytes.Buffer, fileTime int64, fileName string) error {
	//this is started at the file but not included in the checksum
	//hence bytes[len(lzopMagic):]
	//if you include this you will get invalid checksum error on lzop files
	err := binary.Write(buff, binary.BigEndian, lzopMagic)

	err = binary.Write(buff, binary.BigEndian, uint16(version))
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
	bytes := buff.Bytes()
	err = binary.Write(buff, binary.BigEndian, uint32(adler32.Checksum(bytes[len(lzopMagic):])))

	if err != nil {
		return err
	}

	return nil
}

//WriteBytes Writes your bytes to a buffer via compression function
func WriteBytes(buff *bytes.Buffer, data []byte, compressionFunction func([]byte) []byte) error {
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

		err := binary.Write(buff, binary.BigEndian, uint32(len(unCompressed)))
		err = binary.Write(buff, binary.BigEndian, uint32(len(compressed)))
		err = binary.Write(buff, binary.BigEndian, uint32(adler32.Checksum(unCompressed)))

		if err != nil {
			return err
		}

		_, err = buff.Write(compressed)

		if err != nil {
			return err
		}
	}

	return nil
}

//WriteEnd writes the ending bytes of an lzop file
func WriteEnd(buff *bytes.Buffer) error {
	_, err := buff.Write(endBytes)
	return err
}

//write data is a helper to just do all three (header/body/end)
func writeData(buff *bytes.Buffer, fileTime int64, fileName string, data []byte,
	compressionFunction func([]byte) []byte) error {

	err := WriteHeader(buff, fileTime, fileName)

	if err != nil {
		return err
	}

	err = WriteBytes(buff, data, compressionFunction)

	if err != nil {
		return err
	}

	return WriteEnd(buff)
}

//CompressData Will Compress your data expecting an LZO 1X1 compression.  Creates buffer for you
func CompressData(fileTime int64, fileName string, data []byte, compressionFunction func([]byte) []byte) ([]byte, error) {

	buff := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	err := writeData(buff, fileTime, fileName, data, compressionFunction)

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), err
}

//CompressDataWithBuffer Allows you to specify the buffer to use for compression allowing re-use
func CompressDataWithBuffer(buff *bytes.Buffer, fileTime int64, fileName string, data []byte, compressionFunction func([]byte) []byte) ([]byte, error) {

	buff.Reset()
	err := writeData(buff, fileTime, fileName, data, compressionFunction)

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), err
}

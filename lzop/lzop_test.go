package lzop

// Copyright 2016 MediaMath <http://www.mediamath.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/rasky/go-lzo"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func getData(size int) []byte {
	return []byte(uniuri.NewLen(size))
}

func testData(input []byte, compressBuffer, endBuffer *bytes.Buffer) {

	filename := uniuri.NewLen(10) + ".txt"
	lzopFileName := filename + ".lzo"

	defer os.Remove(filename)
	defer os.Remove(lzopFileName)

	var data []byte
	var err error

	if compressBuffer == nil || endBuffer == nil {
		data, err = CompressData(time.Now().Unix(), filename,
			input,
			lzo.Compress1X)
	} else {
		data, err = CompressDataWithBuffers(compressBuffer, endBuffer, time.Now().Unix(), filename,
			input,
			lzo.Compress1X)
	}

	if err != nil {
		log.Fatal(err)
	}

	fo, err := os.Create(lzopFileName)

	if err != nil {
		log.Fatal(err)
	}

	fo.Write(data)

	cmd := exec.Command("lzop", "-d", "-f", lzopFileName)
	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	if len(out) != 0 {
		log.Fatal("Didn't get lzop success")
	}

	_, err = os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	out, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if string(out) != string(input) || len(out) != len(input) {
		log.Fatal("Files not the same")
	}

}

func TestRandomDataSizes(t *testing.T) {
	testData(getData(256), nil, nil)
	testData(getData(512), nil, nil)
	testData(getData(1024), nil, nil)
	testData(getData(256*1024), nil, nil)
	testData(getData(256*1024*2), nil, nil)
	testData(getData(256*1024*4), nil, nil)
	testData(getData(256*1024*8), nil, nil)
	testData(getData(256*1024*16), nil, nil)
	testData(getData(256*1024*32), nil, nil)
	testData(getData(256*1024*64), nil, nil)
	testData(getData(256*1024*128), nil, nil)
}

func TestRandomDataSizesWithPreAllocatedBuffers(t *testing.T) {
	b1 := bytes.NewBuffer(make([]byte, 0, 256*1024*128))
	b2 := bytes.NewBuffer(make([]byte, 0, 256*1024*128))

	testData(getData(256), b1, b2)
	testData(getData(512), b1, b2)
	testData(getData(1024), b1, b2)
	testData(getData(256*1024), b1, b2)
	testData(getData(256*1024*2), b1, b2)
	testData(getData(256*1024*4), b1, b2)
	testData(getData(256*1024*8), b1, b2)
	testData(getData(256*1024*16), b1, b2)
	testData(getData(256*1024*32), b1, b2)
	testData(getData(256*1024*64), b1, b2)
	testData(getData(256*1024*128), b1, b2)
}

func BenchmarkAlloc(b *testing.B) {
	d := getData(256 * 1024 * 128)
	now := time.Now().Unix()
	fn := "file"
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		CompressData(now, fn,
			d,
			lzo.Compress1X)
	}
}

func BenchmarkPreAlloc(b *testing.B) {
	b1 := bytes.NewBuffer(make([]byte, 0, 256*1024*128))
	b2 := bytes.NewBuffer(make([]byte, 0, 256*1024*128))
	d := getData(256 * 1024 * 128)
	now := time.Now().Unix()
	fn := "file"
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		CompressDataWithBuffers(b1, b2, now, fn,
			d,
			lzo.Compress1X)
	}
}

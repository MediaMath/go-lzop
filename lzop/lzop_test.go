package lzop

// Copyright 2016 MediaMath <http://www.mediamath.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
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

func getData(size int) string {
	return uniuri.NewLen(size)
}

func testData(str string) {

	log.Printf("Testing size :%d", len(str))

	filename := uniuri.NewLen(10) + ".txt"
	lzopFileName := filename + ".lzo"

	defer os.Remove(filename)
	defer os.Remove(lzopFileName)

	data, err := CompressData(time.Now().Unix(), filename,
		[]byte(str),
		lzo.Compress1X)

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

	if string(out) != string(str) || len(out) != len(str) {
		log.Fatal("Files not the same")
	}

}

func TestRandomDataSizes(t *testing.T) {
	testData(getData(256))
	testData(getData(512))
	testData(getData(1024))
	testData(getData(256 * 1024))
	testData(getData(256 * 1024 * 2))
	testData(getData(256 * 1024 * 4))
	testData(getData(256 * 1024 * 8))
	testData(getData(256 * 1024 * 16))
	testData(getData(256 * 1024 * 32))
	testData(getData(256 * 1024 * 64))
	testData(getData(256 * 1024 * 128))
}

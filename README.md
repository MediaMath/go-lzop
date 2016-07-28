### go-lzop: A wrapper for LZO compression that works with go-lzo libraries to create lzop compatible files.

#### Go library

To get:
````
go get github.com/MediaMath/go-lzop
````

#### Usage
You can use the library with either your own compiled LZO lib from source, or use https://github.com/rasky/go-lzo.
In this example I'm using the rasky lzoCompress1X.  It's important you only pass our lzop function some iteration of 1x1 compression.  If you want to work out the values for 1x999 or some other inbetween you will have to fork the repo, and use different constants for the binary values written to the header packer.  Writing a level 9 compression while header packing level 1 compression obviously isn't a great idea.
````
import "github.com/MediaMath/go-lzop"

fileCreationTime := time.Now().Unix()
filename := "file.txt"
someString := "Hello I am a string"
data, err := CompressData(fileCreationTime, filename, []byte(someString), lzo.Compress1X)
````
data in this example should be able to be directly written to disk, and decompressed via
````
lzop -d -f file.txt.lzo
````

#### Hexdump
On a xbG7k1TvFZ.txt which contains "data"
````
hexdump -C xbG7k1TvFZ.txt
00000000  64 61 74 61                                       |data|
00000004
````
 
````
lzop -1 -f xbG7k1TvFZ.txt
hexdump -C xbG7k1TvFZ.txt.lzo
00000000  89 4c 5a 4f 00 0d 0a 1a  0a 10 30 20 80 09 40 02  |.LZO......0 ..@.|
00000010  01 03 00 00 01 00 00 81  b4 57 9a 4a 84 00 00 00  |.........W.J....|
00000020  00 0e 78 62 47 37 6b 31  54 76 46 5a 2e 74 78 74  |..xbG7k1TvFZ.txt|
00000030  92 81 09 1f 00 00 00 04  00 00 00 04 04 00 01 9b  |................|
00000040  64 61 74 61 00 00 00 00                           |data....|
00000048
````
Hexdump result of the same file after running through go-lzop and compressed using the LZO 1x1 libs
`````
hexdump -C xbG7k1TvFZ.txt.lzo
00000000  89 4c 5a 4f 00 0d 0a 1a  0a 10 30 20 80 09 40 02  |.LZO......0 ..@.|
00000010  01 03 00 00 01 00 00 81  b4 57 9a 4a 84 00 00 00  |.........W.J....|
00000020  00 0e 78 62 47 37 6b 31  54 76 46 5a 2e 74 78 74  |..xbG7k1TvFZ.txt|
00000030  92 81 09 1f 00 00 00 04  00 00 00 04 04 00 01 9b  |................|
00000040  64 61 74 61 00 00 00 00                           |data....|
00000048
````
All the way back
`````
lzop -d -f xbG7k1TvFZ.txt.lzo
hexdump -C xbG7k1TvFZ.txt
00000000  64 61 74 61                                       |data|
00000004
`````

#### Warning
Data that's created with this packer seems to slightly differ as files grow in size.  Files compressed to 229392890 bytes (219MB) with lzop **might** be 229392**000** bytes with this lib.  LZOP is adding or moving bytes on large files that I haven't found, but it doesn't seem to impact the overall outcome of the data.  If you lzop with built LZO source, and LZO compress with this lib using the same built source, the .lzo produced might differ slightly in size, but contains the same data during decompression.  Diff will return no result on the files.

 
#### Go Test
Tests assume you have lzop installed if not
````
sudo apt-get install lzop
````
Then
````
go get github.com/dchest/uniuri
go get github.com/rasky/go-lzo
cd lzop
go test
````


#### Background information on LZOP header packing which go-lzop follows
````
//packingOrder Is the exact order and type length of everything needed to make a LZOP looking header.
//This struct defines the layout of the data and the types they represent, but isn't used directly
//since each value is constant or generated during compression
//All of these values must be written with some form of
//err = binary.Write(buff, binary.BigEndian (system dependant), <TYPE>(VALUE))
//After writing the complete header you will have to write, in a block size manner, the rest of the compressed data
//Then to mark the end of the file. LZOP was looking for \x00 4 times.
//So make sure you do a f.Write([]byte("\x00\x00\x00\x00"))
type packingOrder struct {
	LZOPMagic               []byte //The LZOP magic byte header []byte{0x89, 0x4c, 0x5a, 0x4f, 0x00, 0x0d, 0x0a, 0x1a, 0x0a}
	Version                 uint16 // Using 0x1030
	LibVersion              uint16 // Using 0x2080
	VersionNeededForExtract uint16 // Using 0x0940
	Method                  uint8  // using 2 because it gets us to -1 compression
	Level                   uint8  // 1 (-1 from command line)
	Flags                   uint32 //Flags is some magic number I dunno.  It looks like they are using it as a way to choose between adler32 and CRC checksum.  0x3000001 works for us
	FileMode                uint32 //0x000081B4 or 33204 (as int) this makes it to -rw-rw-r--
	// Filter                  uint32 //Not using filter at all, they would be written here, but they aren't written
	TimeLow        uint32 //This is the UTC stamp of file creation (time.Now().Unix())
	TimeHigh       uint32 // This is 0 on big endian systems.  It looks like it will give a different result on LittleEndian I think?
	FileNameLength uint8  //The length of the file name
	FileName       string //The file name you are packing

	//CheckSum Stuff
	//Because flags isn't using CRC, we default to adler32 checksum.
	//You need to write all values in this struct minus LZOPMagic and the checksum itself to a buffer.
	//This is done by calling adler32.Checksum() those []bytes and then write
	//Adler32CheckSum result to the buffer/file you were working on as uint32
	Adler32CheckSum uint32

	//After the adler32 checksum there is another checksum + data addition that occurs **EVERY** blocksize rotation

	/*As defined in LZOP
	  #if defined(ACC_OS_DOS16) && !defined(ACC_ARCH_I086PM)
	  #  define BLOCK_SIZE        (128*1024l)
	  #else
	  #  define BLOCK_SIZE        (256*1024l)
	  #endif
	  #define MAX_BLOCK_SIZE      (64*1024l*1024l)        
    */

	// You write the three values below (both sizes + checksum) every rotation

	//This is the uncompressed data length
	UncompressedDataLength uint32

	//Then the compressed data length
	CompressedDataLength uint32

	//This is a checksum where you have to write the uncompressed data to the adler32.Checksum() and it will yield this value
	Adler32UncompressedDataCheckSum uint32
}
````

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

	*/As defined in LZOP
	  #if defined(ACC_OS_DOS16) && !defined(ACC_ARCH_I086PM)
	  #  define BLOCK_SIZE        (128*1024l)
	  #else
	  #  define BLOCK_SIZE        (256*1024l)
	  #endif
	  #define MAX_BLOCK_SIZE      (64*1024l*1024l)        
    /*

	// You write the three values below (both sizes + checksum) every rotation

	//This is the uncompressed data length
	UncompressedDataLength uint32

	//Then the compressed data length
	CompressedDataLength uint32

	//This is a checksum where you have to write the uncompressed data to the adler32.Checksum() and it will yield this value
	Adler32UncompressedDataCheckSum uint32
}
````

package encoding

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

// GetFileEncoder creates encoder with file encoding type and encoding name according to its encoding type, currently supports UTF-8, UTF-8 with BOM, UTF-16LE, UTF-16BE.
// The returned encoder does not add BOM at the start of the transformed buffer.
//
// This function also tries to guess encoding type for some weird files with no BOM prepended(especially UTF-16LE and UTF-16BE files).
func GetFileEncoder(file *os.File) (encoder *encoding.Encoder, encName string, err error) {
	buf := make([]byte, 3) // BOM sequence is at maximum 3 bytes for utf-8
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		return
	}

	encNames := [...]string{"UTF-16LE", "UTF-16BE", "UTF-8", "UTF-8 with BOM", "Unknown"}
	encName = encNames[4] // "Unknown"

	var usedEnc encoding.Encoding

	// use IgnoreBOM to not insert BOM at start of buffer
	encList := [...]encoding.Encoding{unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), unicode.UTF8}

	if buf[0] == '\xff' && buf[1] == '\xfe' {
		usedEnc = encList[0]
		encName = encNames[0] // "UTF-16LE"
	} else if buf[0] == '\xfe' && buf[1] == '\xff' {
		usedEnc = encList[1]
		encName = encNames[1] // "UTF-16BE"
	} else if bytes.HasPrefix(buf, []byte("\ufeff")) {
		usedEnc = encList[2]  // use UTF-8 to not insert BOM at start of buffer
		encName = encNames[3] // "UTF-8 with BOM"
	} else {
		// try decoding with some decoders
		var decList []*encoding.Decoder
		for _, e := range encList {
			decList = append(decList, e.NewDecoder())
		}
		dBuf, resBuf := make([]byte, 4), make([]byte, 4) // utf-8 can use up to 4 bytes for currently used characters
		_, err = file.ReadAt(dBuf, 0)
		if err != nil {
			return nil, encName, err
		}
		for i, dec := range decList {
			_, _, err = dec.Transform(resBuf, dBuf, true)
			if bytes.Count(resBuf, []byte{0}) != 4 && err == nil {
				usedEnc = encList[i]
				encName = encNames[i]
				break
			}
		}

	}
	// TODO: add other encoding types..?

	if usedEnc == nil {
		err = fmt.Errorf("could not get encoding type of given file")
		return
	}

	encoder = usedEnc.NewEncoder()

	return
}

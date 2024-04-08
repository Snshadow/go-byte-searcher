package encoding

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

// get encoder and decoder based on its encoding type, currently supports UTF-8, UTF-16LE, UTF-16BE
// the returned encoder will not add BOM at the start of the transformed buffer
func GetFileEncoder(file *os.File) (encoder *encoding.Encoder, encName string, err error) {
	buf := make([]byte, 3) // BOM sequence is at maximum 3 bytes for utf-8
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		return
	}

    var usedEnc encoding.Encoding

    // use IgnoreBOM to not insert BOM at start of buffer
    if buf[0] == '\xff' && buf[1] == '\xfe' {
        usedEnc = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
        encName = "UTF-16LE"
    } else if buf[0] == '\xfe' && buf[1] == '\xff' {
        usedEnc = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
        encName = "UTF-16BE"
    } else if bytes.HasPrefix(buf, []byte("\ufeff")) {
        usedEnc = unicode.UTF8 // use UTF-8 to not insert BOM at start of buffer
        encName = "UTF-8 with BOM"
    } else {
        usedEnc = unicode.UTF8
        encName = "UTF-8"
    }
    // TODO: add other encoding types..?

    if usedEnc == nil {
        err = fmt.Errorf("could not get encoding type of given file")
        encName = "None"
        return
    }

    encoder = usedEnc.NewEncoder()

    return
}

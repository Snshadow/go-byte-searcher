package searcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"golang.org/x/text/encoding"

	enctool "github.com/Snshadow/go-byte-searcher/encoding"
)

type ByteSearcher struct {
	File    *os.File
	EncType string // UTF-8, UTF-8 with BOM, UTF-16LE, UTF-16BE

	fileSize int64

	encoder *encoding.Encoder

	isComplete *atomic.Bool

	result searchResult
}

type searchResult struct {
	mutex   *sync.Mutex
	offsets []int
}

func (r *searchResult) addResult(offset int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.offsets = append(r.offsets, offset)
}

// NewSearcher opens file descriptor and creates new byte searcher for a file in given path.
// If isText is set to true, the encoder for a file will also be assigned, return error if file is not a text file.
// The file descriptor can be accessed with File field. Needs to be closed after use.
func NewSearcher(path string, isText bool) (ByteSearcher, error) {
	newSearcher := ByteSearcher{
		isComplete: &atomic.Bool{},
		result: searchResult{
			mutex:   &sync.Mutex{},
			offsets: []int{},
		},
	}

	fd, err := os.Open(path)
	if err != nil {
		return newSearcher, err
	}
	newSearcher.File = fd

	fileStat, err := newSearcher.File.Stat()
	if err != nil {
		return newSearcher, err
	}

	if !fileStat.Mode().IsRegular() {
		return newSearcher, fmt.Errorf("the given file is not a regular file")
	}

	newSearcher.fileSize = fileStat.Size()

	if isText {
		newSearcher.encoder, newSearcher.EncType, err = enctool.GetFileEncoder(newSearcher.File)
		if err != nil {
			return newSearcher, err
		}
	}

	return newSearcher, nil
}

// Search finds offsets of a given query from searched file.
// If searchOne is set to true, searcher will search for one offset then return, it may be used if only one match exists in the file is guaranteed,
// runCount sets the number of concurrently run search sessions.
func (s *ByteSearcher) Search(query []byte, searchOne bool, runCount ...uint32) (offsets []int, err error) {
	s.isComplete.Store(false)
	s.result.offsets = make([]int, 0)

	var concur uint32 = 4 // default to 4 concurrent search

	querySize := len(query)

	wg := sync.WaitGroup{}

	if len(runCount) != 0 {
		concur = runCount[0]
	}
	runSize := s.fileSize / int64(concur)
	lastRem := s.fileSize % int64(concur)

	if runSize < int64(querySize) {
		err = fmt.Errorf("session count is too large for given query")
		return
	}

	for session := 0; session < int(concur); session++ {
		readBuf := make([]byte, querySize)
		readOffset := runSize * int64(session)
		session := session
		wg.Add(1)
		go func(f *os.File, b []byte) {
			sz := runSize
			if session == int(concur)-1 {
				sz += lastRem
			}
			for i := 0; i < int(sz); i++ {
				_, err := f.ReadAt(b, readOffset+int64(i))
				if err == io.EOF {
					break
				} else if err != nil {
					fmt.Printf("failed to read buffer from offset %d\n", i+int(readOffset))
					continue
				}

				if bytes.Equal(b, query) {
					s.result.addResult(int(readOffset) + i)
					if searchOne {
						s.isComplete.Store(true)
					}
				}

				if s.isComplete.Load() {
					break
				}
			}

			wg.Done()
		}(s.File, readBuf)
	}
	wg.Wait()

	sort.SliceStable(s.result.offsets, func(i, j int) bool {
		return s.result.offsets[i] < s.result.offsets[j]
	})
	offsets = s.result.offsets

	return
}

// SearchString searches for query string following the encoding type of file for searching.
// See Search for other details.
func (s *ByteSearcher) SearchString(query string, searchOne bool, runCount ...uint32) (offsets []int, err error) {
	if s.encoder == nil {
		err = fmt.Errorf("searcher was not initalized for text file")
		return
	}

	var cr uint32 = 4 // default to 4 concurrent search
	if len(runCount) != 0 {
		cr = runCount[0]
	}

	qBuf, err := s.encoder.Bytes([]byte(query))
	if err != nil {
		return
	}

	offsets, err = s.Search(qBuf, searchOne, cr)

	return
}

// Close closes the file descriptor of a searched file
func (s *ByteSearcher) Close() {
	s.File.Close()
}

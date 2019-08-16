package cbft

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/log"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// The limit size of a single journal file
	journalLimitSize = 100 * 1024 * 1024

	// A new Writer whose buffer has at least the specified size
	writeBufferLimitSize = 16 * 1024

	// A new Reader whose buffer has at least the specified size
	readBufferLimitSize = 16 * 1024

	// The setting of rotate timer ticker
	syncLoopDuration = 5 * time.Second
)

var crc32c = crc32.MakeTable(crc32.Castagnoli)

var (
	errNoActiveJournal = errors.New("no active journal")
	errOpenNewJournal  = errors.New("Failed to open new journal file")
	errWriteJournal    = errors.New("Failed to write journal")
	errLoadJournal     = errors.New("Failed to load journal")
)

type JournalMessage struct {
	Timestamp uint64
	Data      *MsgInfo
}

type sortFile struct {
	name string
	num  uint32
}

type sortFiles []sortFile

func (s sortFiles) Len() int {
	return len(s)
}

func (s sortFiles) Less(i, j int) bool {
	return s[i].num < s[j].num
}

func (s sortFiles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type journal struct {
	path   string         // Filesystem path to store the msgInfo at
	writer *WriterWrapper // Output stream to write new msgInfo into
	fileID uint32
	mu     sync.Mutex
	exitCh chan struct{}
}

func listJournalFiles(path string) sortFiles {
	files, err := ioutil.ReadDir(path)

	if err == nil && len(files) > 0 {
		var s []string
		for _, f := range files {
			s = append(s, f.Name())
		}
		log.Trace("The list of wal directory", "directory", path, "files", strings.Join(s, ","))
		reg := regexp.MustCompile("^wal.([1-9][0-9]*)$")
		regNum := regexp.MustCompile("([1-9][0-9]*)$")
		fs := make(sortFiles, 0)

		for _, f := range s {
			if reg.MatchString(f) {
				n, _ := strconv.Atoi(regNum.FindString(f))
				fs = append(fs, sortFile{
					name: f,
					num:  uint32(n),
				})
			}
		}
		sort.Sort(fs)
		return fs
	}
	return nil
}

// newTxJournal creates journal object
func NewJournal(path string) (*journal, error) {
	journal := &journal{
		path:   path,
		exitCh: make(chan struct{}),
		fileID: 1,
	}
	if files := listJournalFiles(path); files != nil && files.Len() > 0 {
		journal.fileID = files[len(files)-1].num
	}
	// open the corresponding journal file
	newFileID, newWriter, err := journal.newJournalFile(journal.fileID)
	if err != nil {
		return nil, err
	}
	// update field fileID and writer
	journal.fileID = newFileID
	journal.writer = newWriter

	go journal.mainLoop(syncLoopDuration)

	return journal, nil
}

func (journal *journal) mainLoop(syncLoopDuration time.Duration) {
	ticker := time.NewTicker(syncLoopDuration)
	<-ticker.C // discard the initial tick

	for {
		select {
		case <-ticker.C:
			if journal.writer != nil {
				log.Trace("Rotate timer trigger")
				journal.mu.Lock()
				if err := journal.rotate(journalLimitSize); err != nil {
					log.Error("Failed to rotate cbft journal", "err", err)
				}
				journal.mu.Unlock()
			}

		case <-journal.exitCh:
			return
		}
	}
}

// currentJournal retrieves the current fileID and fileSeq of the cbft journal.
func (journal *journal) CurrentJournal() (uint32, uint64, error) {
	journal.mu.Lock()
	defer journal.mu.Unlock()

	// Forced to flush
	journal.writer.Flush()
	fileSeq, err := journal.currentFileSize()
	if err != nil {
		return 0, 0, err
	}

	log.Trace("currentJournal", "fileID", journal.fileID, "fileSeq", fileSeq)
	return journal.fileID, fileSeq, nil
}

// insert adds the specified JournalMessage to the local disk journal.
func (journal *journal) Insert(msg *JournalMessage, sync bool) error {
	journal.mu.Lock()
	defer journal.mu.Unlock()

	if journal.writer == nil {
		return errNoActiveJournal
	}

	buf, err := encodeJournal(msg)
	if err != nil {
		return err
	}
	//
	if err := journal.rotate(journalLimitSize); err != nil {
		log.Error("Failed to rotate cbft journal", "err", err)
		return err
	}

	n := 0
	if n, err = journal.writer.Write(buf); err != nil || n <= 0 {
		log.Error("Write data error", "err", err)
		panic(errWriteJournal)
	}
	if sync {
		// Forced to flush
		if err = journal.writer.Flush(); err != nil {
			log.Error("Flush data error", "err", err)
			panic(err)
		}
	}

	log.Trace("Successful to insert journal message", "n", n)
	return nil
}

func encodeJournal(msg *JournalMessage) ([]byte, error) {
	data, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Error("Failed to encode journal message", "err", err)
		return nil, err
	}

	crc := crc32.Checksum(data, crc32c)
	length := uint32(len(data))
	totalLength := 10 + int(length)

	pack := make([]byte, totalLength)
	binary.BigEndian.PutUint32(pack[0:4], crc)                                   // 4 byte
	binary.BigEndian.PutUint32(pack[4:8], length)                                // 4 byte
	binary.BigEndian.PutUint16(pack[8:10], uint16(WalMessageType(msg.Data.Msg))) // 2 byte

	copy(pack[10:], data)
	return pack, nil
}

// close flushes the journal contents to disk and closes the file.
func (journal *journal) Close() {
	journal.mu.Lock()
	defer journal.mu.Unlock()

	if journal.writer != nil {
		log.Debug("Close journal,flush data")
		journal.writer.FlushAndClose()
		journal.writer = nil
	}
	close(journal.exitCh)
}

func (journal *journal) rotate(journalLimitSize uint64) error {
	//journal.mu.Lock()
	//defer journal.mu.Unlock()

	if journal.checkFileSize(journalLimitSize) {
		journalWriter := journal.writer
		if journalWriter == nil {
			return errNoActiveJournal
		}

		// Forced to flush
		journalWriter.FlushAndClose()
		journal.writer = nil

		// open another new journal file
		newFileID, newWriter, err := journal.newJournalFile(journal.fileID + 1)
		if err != nil {
			panic(err)
		}
		// update field fileID and writer
		journal.fileID = newFileID
		journal.writer = newWriter

		log.Debug("Successful to rotate journal file", "newFileID", newFileID)
	}
	return nil
}

func (journal *journal) checkFileSize(journalLimitSize uint64) bool {
	fileSize, err := journal.currentFileSize()
	return err == nil && fileSize >= journalLimitSize
}

func (journal *journal) currentFileSize() (uint64, error) {
	if fileInfo, err := journal.writer.file.Stat(); err != nil {
		log.Error("Get the current journal file size error", "err", err)
		return 0, err
	} else {
		return uint64(fileInfo.Size()), nil
	}
}

func (journal *journal) newJournalFile(fileID uint32) (uint32, *WriterWrapper, error) {
	newJournalFilePath := filepath.Join(journal.path, fmt.Sprintf("wal.%d", fileID))
	file, err := os.OpenFile(newJournalFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Error("Failed to open new journal file", "fileID", fileID, "filePath", newJournalFilePath, "err", err)
		return 0, nil, errOpenNewJournal
	}

	return fileID, NewWriterWrapper(file, writeBufferLimitSize), nil
}

func (journal *journal) ExpireJournalFile(fileID uint32) error {
	if files := listJournalFiles(journal.path); files != nil && files.Len() > 0 {
		for _, file := range files {
			if file.num != journal.fileID && file.num < fileID {
				os.Remove(filepath.Join(journal.path, fmt.Sprintf("wal.%d", file.num)))
			}
		}
	}
	return nil
}

func (journal *journal) LoadJournal(fromFileID uint32, fromSeq uint64, add func(info *MsgInfo)) error {
	journal.mu.Lock()
	defer journal.mu.Unlock()

	if files := listJournalFiles(journal.path); files != nil && files.Len() > 0 {
		log.Debug("begin to load journal", "fromFileID", fromFileID, "fromSeq", fromSeq)

		err, successBytes, totalBytes := error(nil), int64(0), int64(0)

		for index, file := range files {
			if file.num == fromFileID {
				err, successBytes, totalBytes = journal.loadJournal(file.num, fromSeq, add)
			} else if file.num > fromFileID {
				err, successBytes, totalBytes = journal.loadJournal(file.num, 0, add)
			}
			if err != nil {
				if err == errLoadJournal && index == files.Len()-1 && successBytes+writeBufferLimitSize > totalBytes {
					log.Warn("ignore this load journal error, ")
					journal.rotate(0)
					break
				}
				return err
			}
		}
	} else {
		log.Error("Failed to load journal", "fromFileID", fromFileID, "fromSeq", fromSeq)
		return errLoadJournal
	}
	return nil
}

func (journal *journal) loadJournal(fileID uint32, seq uint64, add func(info *MsgInfo)) (error, int64, int64) {
	file, err := os.Open(filepath.Join(journal.path, fmt.Sprintf("wal.%d", fileID)))
	if err != nil {
		return err, 0, 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err, 0, 0
	}
	defer file.Close()

	successBytes, totalBytes := int64(0), fileInfo.Size()
	bufReader := bufio.NewReaderSize(file, readBufferLimitSize)
	if seq > 0 {
		bufReader.Discard(int(seq))
		successBytes = successBytes + int64(seq)
	}

	for {
		index, _ := bufReader.Peek(10)
		crc := binary.BigEndian.Uint32(index[0:4])      // 4 byte
		length := binary.BigEndian.Uint32(index[4:8])   // 4 byte
		msgType := binary.BigEndian.Uint16(index[8:10]) // 2 byte

		pack := make([]byte, length+10)
		var (
			totalNum = 0
			readNum  = 0
		)
		for totalNum, err = 0, error(nil); err == nil && uint32(totalNum) < length+10; {
			readNum, err = bufReader.Read(pack[totalNum:])
			totalNum = totalNum + readNum
		}

		if 0 == readNum {
			log.Debug("load journal complete", "fileID", fileID, "fileSeq", seq, "successBytes", successBytes)
			break
		}

		// check crc
		_crc := crc32.Checksum(pack[10:], crc32c)
		if crc != _crc {
			log.Error("crc is invalid", "crc", crc, "_crc", _crc, "msgType", msgType)
			return errLoadJournal, successBytes, totalBytes
		}

		// decode journal message
		if msgInfo, err := WALDecode(pack[10:], msgType); err == nil {
			add(msgInfo)
			successBytes = successBytes + int64(totalNum)
		} else {
			log.Error("Failed to decode journal msg", "err", err)
			return errLoadJournal, successBytes, totalBytes
		}
	}
	return nil, successBytes, totalBytes
}

package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"sync"
)

const (
	OpGet    byte = 0
	OpSet    byte = 1
	OpDelete byte = 2
	buf_size      = 5
	Wal_path      = "/home/pianodessources/Projects/distributed-kv-store/log.wal"
)

type walRequest struct {
	data []byte
	errc chan error
}

type WAL struct {
	wg sync.WaitGroup
	ch chan walRequest
	f  *os.File
}

func NewWal(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	buf := make(chan walRequest, buf_size)

	if err != nil {
		return nil, err
	}
	w := &WAL{f: f, ch: buf}

	go w.StartWalWorker()

	return w, nil
}

func (w *WAL) StartWalWorker() {
	var batch []walRequest
	var buf []byte

	for {
		req, ok := <-w.ch
		if !ok {
			return
		}

		batch = append(batch, req)
		buf = append(buf, req.data...)

	drainLoop:
		for len(batch) < buf_size {
			select {
			case r, ok := <-w.ch:
				if !ok {
					break drainLoop
				}
				batch = append(batch, r)
				buf = append(buf, r.data...)
			default:
				break drainLoop
			}
		}

		_, err := w.f.Write(buf)
		if err == nil {
			err = w.f.Sync()
		}

		for _, r := range batch {
			r.errc <- err
		}

		batch = batch[:0]
		buf = buf[:0]

	}

}

func (w *WAL) Append(op byte, key, val []byte) error {
	//structure [OpCode 1B][KeyLen 2B][Key bytes][ValLen 4B][Val bytes]

	keyLen := len(key)
	valLen := len(val)
	size := 1 + 2 + keyLen + 4 + valLen

	//Allocate exactly once for the binary payload
	data := make([]byte, size)
	data[0] = op

	binary.LittleEndian.PutUint16(data[1:3], uint16(keyLen))
	copy(data[3:], key)

	binary.LittleEndian.PutUint32(data[3+keyLen:7+keyLen], uint32(valLen))
	copy(data[7+keyLen:], val)

	req := walRequest{
		data: data,
		errc: make(chan error, 1), //buffer of 1 prevents worker from blocking on notify
	}

	w.ch <- req

	return <-req.errc
}

func (w *WAL) Recover(s *Store) error {
	_, err := w.f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(w.f)

	len16 := make([]byte, 2)
	len32 := make([]byte, 4)

	for {
		op, err := reader.ReadByte()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if _, err := io.ReadFull(reader, len16); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break // corrupted tail end; stop recovering
			}
			return err
		}
		keyLen := binary.LittleEndian.Uint16(len16)

		// read teh key
		keyBuf := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, keyBuf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break // corrupted tail end; stop recovering
			}
			return err
		}
		key := string(keyBuf)

		// read value length
		if _, err := io.ReadFull(reader, len32); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break // corrupted tail end; stop recovering
			}
			return err
		}
		valLen := binary.LittleEndian.Uint32(len32)

		//  read the value
		valBuf := make([]byte, valLen)
		if _, err := io.ReadFull(reader, valBuf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
		//load into the shard thingy

		switch op {
		case OpSet:
			s.Set(key, valBuf)
		case OpDelete:
			s.Delete(key)
		}

	}

	_, err = w.f.Seek(0, io.SeekEnd)
	return err

}

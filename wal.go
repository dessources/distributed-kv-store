package main

import (
	"encoding/binary"
	"io"
	"os"
)

const (
	OpSet      = 1
	OpDelete   = 2
	BufferSize = 1024
)

type WalRequest struct {
	data []byte
	errc chan error
}

type WAL struct {
	f  *os.File
	ch chan WalRequest
}

func NewWal(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	buffer := make(chan WalRequest, BufferSize)
	w := &WAL{
		f:  file,
		ch: buffer,
	}
	go w.StartBackgroundWorker()

	return w, nil
}

func (w *WAL) StartBackgroundWorker() {
	request_buf := make([]WalRequest, BufferSize)
	data_buf := make([]byte, BufferSize)

	for {
		req, ok := <-w.ch
		if !ok {
			return
		}

		request_buf = append(request_buf, req)
		data_buf = append(data_buf, req.data...)

	drainLoop:
		for len(request_buf) < BufferSize {
			select {
			case req, ok = <-w.ch:
				if !ok {
					break drainLoop
				}
				request_buf = append(request_buf, req)
				data_buf = append(data_buf, req.data...)
			default:
				break drainLoop
			}
		}

		_, err := w.f.Write(data_buf)
		if err == nil {
			err = w.f.Sync()
		}

		for _, r := range request_buf {

			r.errc <- err
		}

		request_buf = request_buf[:0]
		data_buf = data_buf[:0]

	}

}

func (w *WAL) Append(op byte, key, val []byte) error {
	keyLen := len(key)
	valLen := len(val)
	size := 1 + 2 + keyLen + 4 + valLen

	data := make([]byte, size)
	data[0] = op

	binary.LittleEndian.PutUint16(data[1:3], uint16(keyLen))
	copy(data[3:], key)

	binary.LittleEndian.PutUint32(data[3+keyLen:7+keyLen], uint32(valLen))
	copy(data[7+keyLen:], val)

	req := WalRequest{
		data: data,
		errc: make(chan error, 1),
	}

	w.ch <- req
	return <-req.errc
}

func (w *WAL) Recover(s *Store) error {
	_, err := w.f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = RecoverFromReader(w.f, s)

	_, seekErr := w.f.Seek(0, io.SeekEnd)
	if err == nil {
		return seekErr
	}

	return err

}

func RecoverFromReader(r io.Reader, s *Store) error {
	return nil
}

package server

import (
	"os"
	"sync"
	"time"
)

type RotateWriter struct {
	lock     sync.Mutex
	filename string // should be set to the actual filename
	fp       *os.File
}

var GlobalWriter *RotateWriter = nil

func SetGlobalWriter(writer *RotateWriter) {
	GlobalWriter = writer
}

// Make a new RotateWriter. Return nil if error occurs during setup.
func NewRotateWriter(filename string) (*RotateWriter, error) {
	w := &RotateWriter{filename: filename}
	err := w.Rotate()
	if err != nil {
		return nil, err
	}
	return w, nil
}

// Start RotateScheduler with duration in seconds
func RotateScheduler(writer *RotateWriter, duration int) {
	for {
		time.Sleep(time.Duration(duration * int(time.Second)))
		writer.Rotate()
	}
}

// Write satisfies the io.Writer interface.
func (w *RotateWriter) Write(output []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.fp.Write(output)
}

// Perform the actual act of rotating and reopening file.
func (w *RotateWriter) Rotate() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	// Close existing file if open
	if w.fp != nil {
		err = w.fp.Close()
		w.fp = nil
		if err != nil {
			return
		}
	}
	// Rename dest file if it already exists
	_, err = os.Stat(w.filename)
	if err == nil {
		err = os.Rename(w.filename, w.filename+"."+time.Now().Format(time.RFC3339))
		if err != nil {
			return
		}
	}

	// Create a file.
	w.fp, err = os.Create(w.filename)
	return
}

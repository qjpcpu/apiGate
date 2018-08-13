package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileLogWriter struct {
	filename     string
	file         *os.File
	writeMtx     *sync.Mutex
	rt           RotateType
	realFilename string
}

type RotateType int

const (
	RotateDaily RotateType = iota
	RotateHourly
	RotateWeekly
)

func logFilename(filename string, rt RotateType) string {
	now := time.Now()
	switch rt {
	case RotateHourly:
		return fmt.Sprintf("%s.%s.%02d", filename, now.Format("2006-01-02"), now.Hour())
	case RotateWeekly:
		offset := int(now.Weekday()) - 1
		if offset < 0 {
			// sunday
			offset = 6
		}
		return fmt.Sprintf("%s.%s", filename, now.AddDate(0, 0, -offset).Format("2006-01-02"))
	default:
		// default rotate daily
		return fmt.Sprintf("%s.%s", filename, now.Format("2006-01-02"))
	}
}

func NewFileLogWriter(filename string, rt RotateType) (*FileLogWriter, error) {
	f, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	filename = f
	w := &FileLogWriter{
		filename:     filename,
		writeMtx:     &sync.Mutex{},
		rt:           rt,
		realFilename: logFilename(filename, rt),
	}
	if err := w.openFile(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *FileLogWriter) openFile() error {
	// Open the log file
	w.realFilename = logFilename(w.filename, w.rt)
	fd, err := os.OpenFile(w.realFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w.file = fd
	linkto, _ := os.Readlink(w.filename)
	if linkto == "" || filepath.Base(linkto) != filepath.Base(w.realFilename) {
		os.Remove(w.filename)
		os.Symlink(filepath.Base(w.realFilename), w.filename)
	}
	return nil
}

// If this is called in a threaded context, it MUST be synchronized
func (w *FileLogWriter) Rotate() error {
	// Close any log file that may be open
	if w.file != nil {
		// fmt.Fprint(w.file, fmt.Sprintf("file logger closed at %s", time.Now().String()))
		w.file.Close()
	}
	// Open the log file
	return w.openFile()
}

func (w *FileLogWriter) needRotate() bool {
	return w.realFilename != logFilename(w.filename, w.rt)
}

func (w *FileLogWriter) Write(p []byte) (int, error) {
	if w.needRotate() {
		w.writeMtx.Lock()
		if w.needRotate() {
			if err := w.Rotate(); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
				w.writeMtx.Unlock()
				return 0, err
			}
		}
		w.writeMtx.Unlock()
	}

	// Perform the write
	n, err := w.file.Write(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return n, err
	}
	return n, err
}

func (w *FileLogWriter) Close() {
	w.file.Close()
}

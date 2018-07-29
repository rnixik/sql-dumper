package main

import (
	"fmt"
	"os"
)

// File is interface which can writes strings to file
type File interface {
	WriteString(s string) (n int, err error)
	Close() error
}

// FileWriter is interface which can instantiate File
type FileWriter interface {
	getFileHandler(filename string) (File, error)
}

// FileWriter writes files using filesystem and methods from OS
type OsFileWriter struct {
	openFiles map[string]bool
}

// NewOsFileWriter builds new OsFileWriter
func NewOsFileWriter() *OsFileWriter {
	return &OsFileWriter{
		make(map[string]bool, 0),
	}
}

func (fw *OsFileWriter) getFileHandler(filename string) (f File, err error) {
	if _, ok := fw.openFiles[filename]; !ok {
		if _, err = os.Stat(filename); !os.IsNotExist(err) {
			// File exists
			return nil, fmt.Errorf("File '%s' already exists", filename)
		} else {
			f, err = os.Create(filename)
			if err != nil {
				return nil, err
			}
			fw.openFiles[filename] = true
		}
	} else {
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	}

	return
}

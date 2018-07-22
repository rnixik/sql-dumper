package main

import (
	"fmt"
	"os"
)

type File interface {
	WriteString(s string) (n int, err error)
	Close() error
}

type FileWriter interface {
	getFileHandler(filename string) (File, error)
}

type OsFileWriter struct {
	openFiles map[string]bool
}

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

package main

import (
	"fmt"
	//"testing"
)

type TestFile struct {
	contents string
}

func (f *TestFile) WriteString(s string) (n int, err error) {
	f.contents += s
	return 1, nil
}
func (f *TestFile) Close() error {
	return nil
}

type TestFileWriter struct {
	files map[string]*TestFile
}

func NewTestFileWriter() *TestFileWriter {
	return &TestFileWriter{
		make(map[string]*TestFile, 0),
	}
}

func (fw *TestFileWriter) getFileHandler(filename string) (f File, err error) {
	testFile := &TestFile{""}
	fw.files[filename] = testFile
	return testFile, nil
}

func (fw *TestFileWriter) getContents(filename string) string {
	testFile := fw.files[filename]
	return testFile.contents
}

// TestErrorFile

type TestErrorFile struct {
}

func (f *TestErrorFile) WriteString(s string) (n int, err error) {
	return 0, fmt.Errorf("Some testing error at WriteString")
}
func (f *TestErrorFile) Close() error {
	return nil
}

type TestFileErrorWriter struct {
}

func (fw *TestFileErrorWriter) getFileHandler(filename string) (f File, err error) {
	testFile := &TestErrorFile{}
	return testFile, nil
}

type TestFileHandlerErrorWriter struct {
}

func (fw *TestFileHandlerErrorWriter) getFileHandler(filename string) (f File, err error) {
	return nil, fmt.Errorf("Some testing error at getFileHandler")
}

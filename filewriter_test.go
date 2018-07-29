package main

import (
	"fmt"
	"os"
	"testing"
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

// TestFileWhichFailsAtSecondAttempt

type TestFileWhichFailsAtSecondAttempt struct {
	attempt int
}

func (f *TestFileWhichFailsAtSecondAttempt) WriteString(s string) (n int, err error) {
	f.attempt += 1
	if f.attempt > 1 {
		return 0, fmt.Errorf("Error at second attempt to write file")
	}
	return 1, nil
}
func (f *TestFileWhichFailsAtSecondAttempt) Close() error {
	return nil
}

type TestFileWhichFailsAtSecondAttemptWriter struct {
}

func (fw *TestFileWhichFailsAtSecondAttemptWriter) getFileHandler(filename string) (f File, err error) {
	testFile := &TestFileWhichFailsAtSecondAttempt{}
	return testFile, nil
}

// Tests on NewOsFileWriter

func TestGetFileHandlerFileExists(t *testing.T) {
	fw := NewOsFileWriter()
	_, err := fw.getFileHandler(".env.example")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestGetFileHandlerDoubleAccess(t *testing.T) {
	os.Remove("test_file_handler")
	fw := NewOsFileWriter()
	_, err := fw.getFileHandler("test_file_handler")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	_, err = fw.getFileHandler("test_file_handler")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	os.Remove("test_file_handler")
}

func TestGetFileHandlerCreatingError(t *testing.T) {
	fw := NewOsFileWriter()
	_, err := fw.getFileHandler("/not_writtable/file")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

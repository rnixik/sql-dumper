package main

import "testing"

func TestCsvWriterWriteRows(t *testing.T) {
	fw := NewTestFileWriter()
	writer := NewCsvWriter(fw, "result.csv", "", ",")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{
		"name":    "one",
		"title":   "t\"wo",
		"id":      int(123),
		"value":   int64(456),
		"amount":  1.23,
		"chars":   []uint8{0x26, 0x23, 0x29},
		"nulled":  nil,
		"strange": uintptr(1),
	})
	rows = append(rows, &map[string]interface{}{
		"name":    "four",
		"title":   "five",
		"id":      int(789),
		"value":   int64(345),
		"amount":  2.23,
		"chars":   []uint8{0x23, 0x23, 0x29},
		"nulled":  nil,
		"strange": uintptr(1),
	})

	writer.WriteRows("some_table", []string{"name", "title", "id", "value", "amount", "chars", "nulled", "strange"}, rows)
	result := fw.getContents("result.csv")
	expected := "\"name\",\"title\",\"id\",\"value\",\"amount\",\"chars\",\"nulled\",\"strange\"\r\n"
	expected += "\"one\",\"t\"\"wo\",123,456,1.230000,\"&#)\",NULL,UNDEFINED\r\n"
	expected += "\"four\",\"five\",789,345,2.230000,\"##)\",NULL,UNDEFINED\r\n"
	if expected != result {
		t.Errorf("Expected:\n%sGot:\n%s", expected, result)
	}
}

func TestCsvWriterWriteDDL(t *testing.T) {
	fw := NewTestFileWriter()
	writer := NewCsvWriter(fw, "result.csv", "", ",")
	err := writer.WriteDDL("some_table", "")
	if err != nil {
		t.Errorf("Unexpected error: ", err)
	}
}

func TestCsvWriterWriteRowsFileWriteError(t *testing.T) {
	fw := &TestFileErrorWriter{}
	writer := NewCsvWriter(fw, "result.csv", "", ",")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{"name": "one"})
	err := writer.WriteRows("some_table", []string{"name"}, rows)
	if err == nil {
		t.Errorf("Expected file writer error, but got nil ")
	}
}

func TestCsvWriterWriteRowsFileHandlerError(t *testing.T) {
	fw := &TestFileHandlerErrorWriter{}
	writer := NewCsvWriter(fw, "result.csv", "", ",")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{"name": "one"})
	err := writer.WriteRows("some_table", []string{"name"}, rows)
	if err == nil {
		t.Errorf("Expected write error, but got nil ")
	}
}

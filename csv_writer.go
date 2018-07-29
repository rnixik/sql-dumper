package main

import (
	"fmt"
	"strings"
)

// CsvWriter writes data in csv format using FileWriter
type CsvWriter struct {
	fw        FileWriter
	dstFile   string
	dstDir    string
	delimiter string
}

// NewCsvWriter builds new CsvWriter
func NewCsvWriter(fw FileWriter, dstFile, dstDir, delimiter string) *CsvWriter {
	return &CsvWriter{
		fw,
		dstFile,
		dstDir,
		delimiter,
	}
}

// WriteDDL is part of interface. It is not usefull for csv. 
func (w *CsvWriter) WriteDDL(tableName string, ddl string) (err error) {
	return
}

// WriteRows writes result rows in csv format
func (w *CsvWriter) WriteRows(tableName string, columns []string, rows []*map[string]interface{}) (err error) {
	f, err := w.fw.getFileHandler(w.getFilename(tableName))
	if err != nil {
		return err
	}
	defer f.Close()
	columnsNames := make([]string, 0)
	for _, column := range columns {
		columnsNames = append(columnsNames, escapeCsvString(column))
	}
	_, err = f.WriteString(strings.Join(columnsNames, w.delimiter) + "\r\n")
	if err != nil {
		return fmt.Errorf("Error at writing header to file: %s", err)
	}
	for _, row := range rows {
		values := make([]string, 0)
		for _, field := range columns {
			v := (*row)[field]
			value := ""
			switch typedValue := v.(type) {
			case int:
				value = fmt.Sprintf("%d", typedValue)
				break
			case int64:
				value = fmt.Sprintf("%d", typedValue)
				break
			case float64:
				value = fmt.Sprintf("%f", typedValue)
				break
			case string:
				value = escapeCsvString(typedValue)
				break
			case []uint8:
				value = escapeCsvString(fmt.Sprintf("%s", typedValue))
			case nil:
				value = "NULL"
			default:
				value = "UNDEFINED"
			}
			values = append(values, value)
		}
		_, err = f.WriteString(strings.Join(values, w.delimiter) + "\r\n")
		if err != nil {
			return fmt.Errorf("Error at writing rows to file: %s", err)
		}
	}
	return
}

func (w *CsvWriter) getFilename(tableName string) (filename string) {
	if w.dstDir != "" {
		return w.dstDir + "/" + tableName + ".csv"
	}
	return w.dstFile
}

func escapeCsvString(str string) string {
	str = strings.Replace(str, "\"", "\"\"", -1)
	return "\"" + str + "\""
}

package main

import (
	"fmt"
	"strings"
)

// SqlWriter writes data in sql format using FileWriter
type SqlWriter struct {
	fw      FileWriter
	dstFile string
	dstDir  string
}

// NewSqlWriter builds new SqlWriter
func NewSqlWriter(fw FileWriter, dstFile, dstDir string) *SqlWriter {
	return &SqlWriter{
		fw,
		dstFile,
		dstDir,
	}
}

// WriteDDL writes SQL-query which creates tables - DDL
func (w *SqlWriter) WriteDDL(tableName string, ddl string) (err error) {
	contents := "SET FOREIGN_KEY_CHECKS=0;\n"
	contents += ddl + "\n"
	contents += "SET FOREIGN_KEY_CHECKS=1;\n"
	f, err := w.fw.getFileHandler(w.getFilename(tableName))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		return fmt.Errorf("Error at writing DDL to file: %s", err)
	}
	return
}

// WriteRows writes result rows in sql-insert format
func (w *SqlWriter) WriteRows(tableName string, columns []string, rows []*map[string]interface{}) (err error) {
	f, err := w.fw.getFileHandler(w.getFilename(tableName))
	if err != nil {
		return err
	}
	defer f.Close()
	columnsNames := make([]string, 0)
	for _, column := range columns {
		columnsNames = append(columnsNames, "`"+column+"`")
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
				value = escapeString(typedValue)
				break
			case []uint8:
				value = escapeString(fmt.Sprintf("%s", typedValue))
			case nil:
				value = "NULL"
			default:
				value = "UNDEFINED"
			}
			values = append(values, value)
		}
		insert := "INSERT INTO `" + tableName + "` (" + strings.Join(columnsNames, ", ") + ") " +
			"VALUES (" + strings.Join(values, ", ") + ");\n"
		_, err = f.WriteString(insert)
		if err != nil {
			return fmt.Errorf("Error at writing rows to file: %s", err)
		}
	}
	return
}

func (w *SqlWriter) getFilename(tableName string) (filename string) {
	if w.dstDir != "" {
		return w.dstDir + "/" + tableName + ".sql"
	}
	return w.dstFile
}

func escapeString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "'", "\\'", -1)
	return "'" + str + "'"
}

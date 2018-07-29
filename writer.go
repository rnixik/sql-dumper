package main

import (
	"fmt"
)

// DataWriter is interface which can write result somewhere
type DataWriter interface {
	WriteRows(tableName string, columns []string, results []*map[string]interface{}) (err error)
	WriteDDL(tableName string, ddl string) (err error)
}

// SimpleWriter writes result into stdout using simple format (concatinated values)
type SimpleWriter struct {
}

// WriteRows prints result rows in simple format (concatinated values) into stdout
func (w *SimpleWriter) WriteRows(tableName string, _ []string, rows []*map[string]interface{}) (err error) {
	fmt.Println(tableName)
	for _, row := range rows {
		for field, v := range *row {
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
				value = typedValue
				break
			case []uint8:
				value = fmt.Sprintf("%s", typedValue)
			case nil:
				value = "NULL"
			default:
				value = "UNDEFINED"
			}
			fmt.Printf("%s = %s;||", field, value)
		}
		fmt.Println("")
	}
	return nil
}

// WriteDDL prints DDL as is into stdout
func (w *SimpleWriter) WriteDDL(tableName string, ddl string) (err error) {
	fmt.Println(ddl)
	return nil
}

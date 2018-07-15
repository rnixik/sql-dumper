package main

import (
	"fmt"
)

type DataWriter interface {
	Write(results []*map[string]interface{}) (err error)
}

type SimpleWriter struct {
}

func (w *SimpleWriter) Write(results []*map[string]interface{}) (err error) {
	for _, row := range results {
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
			fmt.Printf("%s = %s; ", field, value)
		}
		fmt.Println("")
	}
	return nil
}

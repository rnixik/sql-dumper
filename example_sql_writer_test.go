package main

import "fmt"

func ExampleSqlWriter_WriteRows() {
	fw := NewTestFileWriter()
	writer := NewSqlWriter(fw, "result.sql", "")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{
		"name":    "one",
		"title":   "two",
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
	fmt.Printf(fw.getContents("result.sql"))

	// Output:
	// INSERT INTO `some_table` (`name`, `title`, `id`, `value`, `amount`, `chars`, `nulled`, `strange`) VALUES ('one', 'two', 123, 456, 1.230000, '&#)', NULL, UNDEFINED);
	// INSERT INTO `some_table` (`name`, `title`, `id`, `value`, `amount`, `chars`, `nulled`, `strange`) VALUES ('four', 'five', 789, 345, 2.230000, '##)', NULL, UNDEFINED);
}

func ExampleSqlWriter_WriteDDL() {
	fw := NewTestFileWriter()
	writer := NewSqlWriter(fw, "result.sql", "")
	ddl := "CREATE TABLE `some_table` (\n"
	ddl += "    `id` bigint(20) NOT NULL,\n"
	ddl += "    `name` varchar(100) NOT NULL,\n"
	ddl += "    PRIMARY KEY (`id`),\n"
	ddl += "    UNIQUE INDEX `name` (`name`)\n"
	ddl += ");"
	writer.WriteDDL("some_table", ddl)
	fmt.Printf(fw.getContents("result.sql"))

	// Output:
	// SET FOREIGN_KEY_CHECKS=0;
	// CREATE TABLE `some_table` (
	//     `id` bigint(20) NOT NULL,
	//     `name` varchar(100) NOT NULL,
	//     PRIMARY KEY (`id`),
	//     UNIQUE INDEX `name` (`name`)
	// );
	// SET FOREIGN_KEY_CHECKS=1;
}

func ExampleSqlWriter_WriteDDL_toDir() {
	fw := NewTestFileWriter()
	writer := NewSqlWriter(fw, "", "/tmp/some_dir")
	ddl1 := "CREATE TABLE `some_table1` (\n"
	ddl1 += "    `id` bigint(20) NOT NULL\n"
	ddl1 += ");"
	writer.WriteDDL("some_table1", ddl1)
	ddl2 := "CREATE TABLE `some_table2` (\n"
	ddl2 += "    `id` bigint(20) NOT NULL\n"
	ddl2 += ");"
	writer.WriteDDL("some_table2", ddl2)
	fmt.Printf(fw.getContents("/tmp/some_dir/some_table1.sql"))
	fmt.Printf(fw.getContents("/tmp/some_dir/some_table2.sql"))

	// Output:
	// SET FOREIGN_KEY_CHECKS=0;
	// CREATE TABLE `some_table1` (
	//     `id` bigint(20) NOT NULL
	// );
	// SET FOREIGN_KEY_CHECKS=1;
	// SET FOREIGN_KEY_CHECKS=0;
	// CREATE TABLE `some_table2` (
	//     `id` bigint(20) NOT NULL
	// );
	// SET FOREIGN_KEY_CHECKS=1;
}

func ExampleSqlWriter_WriteRows_fileWriteError() {
	fw := &TestFileErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{"name": "one"})
	err := writer.WriteRows("some_table", []string{"name"}, rows)
	fmt.Print(err)

	// Output:
	// Error at writing rows to file: Some testing error at WriteString
}

func ExampleSqlWriter_WriteRows_fileHandlerError() {
	fw := &TestFileHandlerErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{"name": "one"})
	err := writer.WriteRows("some_table", []string{"name"}, rows)
	fmt.Print(err)

	// Output:
	// Some testing error at getFileHandler
}

func ExampleSqlWriter_WriteDDL_fileWriteError() {
	fw := &TestFileErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")
	ddl := "CREATE TABLE `some_table` (\n"
	ddl += "    `id` bigint(20) NOT NULL\n"
	ddl += ");"
	err := writer.WriteDDL("some_table", ddl)
	fmt.Print(err)

	// Output:
	// Error at writing DDL to file: Some testing error at WriteString
}

func ExampleSqlWriter_WriteDDL_fileHandlerError() {
	fw := &TestFileHandlerErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")
	ddl := "CREATE TABLE `some_table` (\n"
	ddl += "    `id` bigint(20) NOT NULL\n"
	ddl += ");"
	err := writer.WriteDDL("some_table", ddl)
	fmt.Print(err)

	// Output:
	// Some testing error at getFileHandler
}

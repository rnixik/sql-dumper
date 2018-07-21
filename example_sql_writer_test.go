package main

func Example_SqlWriterWriteRows() {
	writer := &SqlWriter{}
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

	// Output:
	// INSERT INTO `some_table` (`name`, `title`, `id`, `value`, `amount`, `chars`, `nulled`, `strange`) VALUES ('one', 'two', 123, 456, 1.230000, '&#)', NULL, UNDEFINED);
	// INSERT INTO `some_table` (`name`, `title`, `id`, `value`, `amount`, `chars`, `nulled`, `strange`) VALUES ('four', 'five', 789, 345, 2.230000, '##)', NULL, UNDEFINED);
}

func Example_SqlWriterWriteDDL() {
	writer := &SqlWriter{}
	ddl := "CREATE TABLE `some_table` (\n"
	ddl += "    `id` bigint(20) NOT NULL,\n"
	ddl += "    `name` varchar(100) NOT NULL,\n"
	ddl += "    PRIMARY KEY (`id`),\n"
	ddl += "    UNIQUE INDEX `name` (`name`)\n"
	ddl += ");"
	writer.WriteDDL("some_table", ddl)

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

package main

func Example_WriteRows() {
	writer := &SimpleWriter{}
	rows := make([]*map[string]interface{}, 0)
	rows = append(rows, &map[string]interface{}{"name": "one", "title": "two"})
	rows = append(rows, &map[string]interface{}{"id": int(123)})
	rows = append(rows, &map[string]interface{}{"value": int64(456)})
	rows = append(rows, &map[string]interface{}{"amount": 1.23})
	rows = append(rows, &map[string]interface{}{"chars": []uint8{0x26, 0x23, 0x29}})
	rows = append(rows, &map[string]interface{}{"nulled": nil})
	rows = append(rows, &map[string]interface{}{"strange": uintptr(1)})

	writer.WriteRows("some_table", []string{}, rows)

	// Output:
	// some_table
	// name = one;||title = two;||
	// id = 123;||
	// value = 456;||
	// amount = 1.230000;||
	// chars = &#);||
	// nulled = NULL;||
	// strange = UNDEFINED;||
}

func Example_WriteDDL() {
	writer := &SimpleWriter{}
	ddl := "CREATE TABLE `some_table` (\n"
	ddl += "    `id` bigint(20) NOT NULL,\n"
	ddl += "    `name` varchar(100) NOT NULL,\n"
	ddl += "    PRIMARY KEY (`id`),\n"
	ddl += "    UNIQUE INDEX `name` (`name`)\n"
	ddl += ");"
	writer.WriteDDL("some_table", ddl)

	// Output:
	// CREATE TABLE `some_table` (
	//     `id` bigint(20) NOT NULL,
	//     `name` varchar(100) NOT NULL,
	//     PRIMARY KEY (`id`),
	//     UNIQUE INDEX `name` (`name`)
	// );
}

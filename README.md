# sql-dumper - command-line tool to dump a portion of data from DB

[![Build Status](https://travis-ci.org/rnixik/sql-dumper.svg?branch=master)](https://travis-ci.org/rnixik/sql-dumper) [![Coverage Status](https://coveralls.io/repos/github/rnixik/sql-dumper/badge.svg?branch=master)](https://coveralls.io/github/rnixik/sql-dumper?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/rnixik/sql-dumper)](https://goreportcard.com/report/github.com/rnixik/sql-dumper)

## Usage

```
Usage: sql-dumper [OPTIONS] <tables> <interval> [relations]

Options:
  --config <filename>        File with settings of connection to DB.
                             It will be used if environment variable DB_NAME is not defined (default .env)
  --format {sql|csv|simple}  Format of output format (default sql)
  --csv-delimiter            Sets delimiter of values in CSV (default ,)
  --file <filename>          Specify file to save combined result from all tables. Can't be used with --dir (default result.sql)
  --dir <directory>          Specify directory to save the result in a separate file for every table

Arguments:

  tables     List of tables and columns to dump: table1:column11,column12,...,column1N;table2:column21;...
  interval   Interval of values for the first column in the first table to select from DB: int-int
  relations  List of relations between chosen tables and columns:
             table1.column11=table2.column21;table2.column22=table3.column31

Example:

  sql-dumper "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
     2000-2200 \
     "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"

```

By default, the tool reads connection settings from environment variables:

```
DB_USER
DB_PASSWORD
DB_NAME
DB_HOST
```

If it can't read values, it reads from file `.env`. Filename with config can be specified with option `--config <filename>`.
Example if `.env` can be found in file `.env.example`.

## Examples

For example, you have tables with DDL:

```
CREATE TABLE `routes` (
    `id` bigint(20) NOT NULL,
    `name` varchar(100) NOT NULL,
    `unused` varchar(100) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `name` (`name`)
);
CREATE TABLE `stations` (
    `id` bigint(20) NOT NULL,
    `name` varchar(150) NOT NULL,
    `unused` varchar(100) NOT NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `stations_for_routes` (
    `station_id` bigint(20) NOT NULL,
    `route_id` bigint(20) NOT NULL,
    `ord` int(11) NOT NULL DEFAULT '0',
    `unused` varchar(100) NOT NULL,
    PRIMARY KEY (`station_id`, `route_id`, `ord`),
    CONSTRAINT `fk_station_id` FOREIGN KEY (`station_id`) REFERENCES `stations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_route_id` FOREIGN KEY (`route_id`) REFERENCES `routes` (`id`) ON DELETE CASCADE
);
```

To extract information about `routes` with `id` in interval BETWEEN 100 AND 200 with relative information from
`stations` and `stations_for_routes` run:


```
sql-dumper --config stations.ini \
    "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
    100-200 \
    "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"
```

It will save DDL for mentioned tables and data in SQL-insert format.


### Combined result in one SQL-file
```
sql-dumper --config stations.ini --file result.sql \
    "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
    100-102 \
    "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"
```

Output in result.sql:

```
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `routes` (
    `id` bigint(20) NOT NULL,
    `name` varchar(100) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `name` (`name`)
);
SET FOREIGN_KEY_CHECKS=1;
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `stations` (
    `id` bigint(20) NOT NULL,
    `name` varchar(150) NOT NULL,
    PRIMARY KEY (`id`)
);
SET FOREIGN_KEY_CHECKS=1;
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `stations_for_routes` (
    `station_id` bigint(20) NOT NULL,
    `route_id` bigint(20) NOT NULL,
    `ord` int(11) NOT NULL DEFAULT '0',
    PRIMARY KEY (`station_id`, `route_id`, `ord`),
    CONSTRAINT `fk_station_id` FOREIGN KEY (`station_id`) REFERENCES `stations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_route_id` FOREIGN KEY (`route_id`) REFERENCES `routes` (`id`) ON DELETE CASCADE
);
SET FOREIGN_KEY_CHECKS=1;
INSERT INTO `routes` (`id`, `name`) VALUES (100, 'Route 1');
INSERT INTO `routes` (`id`, `name`) VALUES (101, 'Route 2');
INSERT INTO `routes` (`id`, `name`) VALUES (102, 'Route 3');
INSERT INTO `stations` (`id`, `name`) VALUES (1, 'Station 1');
INSERT INTO `stations` (`id`, `name`) VALUES (2, 'Station 2');
INSERT INTO `stations` (`id`, `name`) VALUES (3, 'Station 3');
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (1, 100, 0);
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (2, 101, 0);
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (2, 102, 1);
```


### Separated result in SQL-files
```
sql-dumper --config stations.ini --dir . \
    "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
    100-102 \
    "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"
```

Output in routes.sql:

```
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `routes` (
    `id` bigint(20) NOT NULL,
    `name` varchar(100) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `name` (`name`)
);
SET FOREIGN_KEY_CHECKS=1;
INSERT INTO `routes` (`id`, `name`) VALUES (100, 'Route 1');
INSERT INTO `routes` (`id`, `name`) VALUES (101, 'Route 2');
INSERT INTO `routes` (`id`, `name`) VALUES (102, 'Route 3');
```

Output in stations.sql:

```
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `stations` (
    `id` bigint(20) NOT NULL,
    `name` varchar(150) NOT NULL,
    PRIMARY KEY (`id`)
);
SET FOREIGN_KEY_CHECKS=1;
INSERT INTO `stations` (`id`, `name`) VALUES (1, 'Station 1');
INSERT INTO `stations` (`id`, `name`) VALUES (2, 'Station 2');
INSERT INTO `stations` (`id`, `name`) VALUES (3, 'Station 3');
```

Output in stations_for_routes.sql:

```
SET FOREIGN_KEY_CHECKS=0;
CREATE TABLE `stations_for_routes` (
    `station_id` bigint(20) NOT NULL,
    `route_id` bigint(20) NOT NULL,
    `ord` int(11) NOT NULL DEFAULT '0',
    PRIMARY KEY (`station_id`, `route_id`, `ord`),
    CONSTRAINT `fk_station_id` FOREIGN KEY (`station_id`) REFERENCES `stations` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_route_id` FOREIGN KEY (`route_id`) REFERENCES `routes` (`id`) ON DELETE CASCADE
);
SET FOREIGN_KEY_CHECKS=1;
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (1, 100, 0);
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (2, 101, 0);
INSERT INTO `stations_for_routes` (`station_id`, `route_id`, `ord`) VALUES (2, 102, 1);
```


### Combined result in one CSV-file (INNER-JOIN)
```
sql-dumper --config stations.ini --format csv --csv-delimiter "," --file result.csv \
    "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
    100-102 \
    "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"
```

Output in result.csv:

```
"routes.id","routes.name","stations.id","stations.name","stations_for_routes.station_id","stations_for_routes.route_id","stations_for_routes.ord"
100,"Route 1",1,"Station 1",1,100,0
101,"Route 2",1,"Station 2",2,101,0
102,"Route 3",1,"Station 2",2,102,1

```


### Separated result in CSV-files
```
sql-dumper --config stations.ini --format csv --csv-delimiter "," --dir . \
    "routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord" \
    100-102 \
    "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id"
```

Output in routes.csv:

```
"id","name"
100,"Route 1"
101,"Route 2"
102,"Route 3"
```

Output in stations.csv:

```
"id","name"
1,"Station 1"
1,"Station 2"
1,"Station 3"
```

Output in stations_for_routes.csv:

```
"station_id","route_id","ord"
1,100,0
2,101,0
2,102,1
```

## Limitations

* It supports only MySQL
* Not full range of column types is supported
* It does not support composite index except PK
* It writes DDL with FK by specified relations in arguments
* Combined result for one CSV made by INNER JOIN
* Escaping output values can go wrong

## License

The MIT License

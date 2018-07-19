# sql-dumper - command-line tool to dump a portion of data from DB

[![Build Status](https://travis-ci.org/rnixik/sql-dumper.svg?branch=master)](https://travis-ci.org/rnixik/sql-dumper) [![Coverage Status](https://coveralls.io/repos/github/rnixik/sql-dumper/badge.svg?branch=master)](https://coveralls.io/github/rnixik/sql-dumper?branch=master)

## Work in progress

## Usage

```
sql-dumper [--config .env] "tables definition" "interval definition" ["relations definition"]
```

For example, you have tables with DDL:

```
CREATE TABLE `routes` (
    `id` bigint(20) NOT NULL,
    `name` varchar(100) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `name` (`name`)
);
CREATE TABLE `stations` (
    `id` bigint(20) NOT NULL,
    `name` varchar(150) NOT NULL,
    PRIMARY KEY (`id`)
);
CREATE TABLE `stations_for_routes` (
    `station_id` bigint(20) NOT NULL,
    `route_id` bigint(20) NOT NULL,
    `ord` int(11) NOT NULL DEFAULT '0',
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

It will output DDL for mentioned tables and data in SQL-insert format.

## License

The MIT License

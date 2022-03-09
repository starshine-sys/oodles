#!/bin/sh
pggen gen go \
--query-glob 'star/queries/queries.sql' \
--output-dir 'star/queries' \
--postgres-connection 'postgresql://oodles:password@localhost/oodles' \
--go-type 'int8=int64' \
--go-type 'text=string' \
--go-type 'int4=int' \
--go-type 'boolean=bool'
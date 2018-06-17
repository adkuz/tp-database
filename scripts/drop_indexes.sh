#!/bin/bash



for idx in $(for line in $(cat src/sql/dbscheme.sql | grep -E '[a-z_]+_idx'); do echo $line | grep -E '.*_idx'; done;); 
do 
    echo "DROP INDEX IF EXISTS $idx;"; 
done;
#!/bin/bash

# This script adds empty down migrations for any migration that misses them.
# Adding them is necessary because if the down migration is missing, the
# migration will only be applied once, even if the database is completely
# rolled back.
# In newer versions of ory/x/popx, the migration box enforces that all up
# migrations have a down migration. Use this script to add them.

set -Eeuo pipefail

for f in $(find . -name "*.up.sql"); do
	base=$(basename $f)
	dir=$(dirname $f)
	migra_name=$(echo $base | sed -e "s/\..*\.up\.sql//" | sed -e "s/\.up\.sql//")
	if ! compgen -G "$dir/$migra_name*.down.sql" > /dev/null; then
		echo "Adding empty down migration for $f"
		touch $dir/$migra_name.down.sql
	fi
done
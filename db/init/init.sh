#!/bin/bash
set -e

mysql -u root -p${MYSQL_ROOT_PASSWORD} <<-EOSQL
  CREATE DATABASE IF NOT EXISTS alice;
  CREATE DATABASE IF NOT EXISTS bob;
EOSQL
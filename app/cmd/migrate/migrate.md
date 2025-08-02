# Migrating to MySQL

## Initialisation

    docker run --detach --name mariadb-dogs -p 3306:3306 --volume migrate.sql:/docker-entrypoint-initdb.d/migrate.sql --volume ./data:/var/lib/mysql:Z --env MARIADB_ROOT_PASSWORD=xIbu0Qm7lWR4TqmM --env MARIADB_USER=dogs --env MARIADB_PASSWORD=byBpn9Znss8Onl3C --env MARIADB_DATABASE=dogs mariadb:11.8

## Running it next time

    docker start --detach --name mariadb-dogs -p 3306:3306 --volume ./data:/var/lib/mysql:Z --env MARIADB_USER=dogs --env MARIADB_PASSWORD=byBpn9Znss8Onl3C --env MARIADB_DATABASE=dogs mariadb:11.8

## Run CLI/client

    docker run -it --network host --volume .:/project:Z --rm mariadb mariadb -h 127.0.0.1 -u dogs --password --database dogs --ssl-verify-server-cert=OFF


    show DATABASES;
    show TABLES;

Note the mapping of pwd to /project. So you can source sql files with:

    \. /project/initialise.sql

try `\s`

    exit

## Generating stringer code

    go generate ./...

See https://last9.io/blog/golang-stringer-tool/

## Inspecting with jq

man page for jq is great. Just read it top to bottom.

This outputs the number of records in each tx

    jq -c '.[]|.Records|length' dogs.json

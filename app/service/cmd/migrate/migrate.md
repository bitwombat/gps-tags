# Migrating to MySQL

## Initialisation

    docker run --detach --name mariadb-dogs -p 3306:3306 --volume setup_new_tables.sql:/docker-entrypoint-initdb.d/migrate.sql --volume ./data:/var/lib/mysql:Z --env MARIADB_ROOT_PASSWORD=xIbu0Qm7lWR4TqmM --env MARIADB_USER=dogs --env MARIADB_PASSWORD=byBpn9Znss8Onl3C --env MARIADB_DATABASE=dogs mariadb:11.8

    If you watch the log, it exits with an error about readline. That's OK.

## Running it next time

    docker start mariadb-dogs

## Run CLI/client

    docker run -it --network host --volume .:/project:Z --rm mariadb:11.8 mariadb -h 127.0.0.1 -u dogs --password --database dogs --ssl-verify-server-cert=OFF


    show DATABASES;
    show TABLES;

Note the mapping of pwd to /project. So you can source sql files with:

    \. /project/setup_new_tables.sql   # Though, this should have run when the container was created above.

try `\s`

    exit

## Running sql

    docker exec -i mariadb-dogs mariadb -udogs -pbyBpn9Znss8Onl3C dogs < setup_new_tables.sql
    docker exec -i mariadb-dogs mariadb -udogs -pbyBpn9Znss8Onl3C dogs < dogs_pre_20250711.sql

## Generating stringer code

    go generate ./...

See https://last9.io/blog/golang-stringer-tool/

## Inspecting with jq

man page for jq is great. Just read it top to bottom.

This outputs the number of records in each tx

    jq -c '.[]|.Records|length' dogs.json

Migrating to MySQL

Analysing the JSON output from MongoDB

These are the reasons: 1 2 3 6 36 37


    docker run --detach --name mariadb-dogs -p 3306:3306 --volume migrate.sql:/docker-entrypoint-initdb.d/migrate.sql --volume ./data:/var/lib/mysql:Z --env MARIADB_ROOT_PASSWORD=xIbu0Qm7lWR4TqmM --env MARIADB_USER=dogs --env MARIADB_PASSWORD=byBpn9Znss8Onl3C --env MARIADB_DATABASE=dogs mariadb:11.8

Later, you start the stopped container:

    docker start --detach --name mariadb-dogs -p 3306:3306 --volume ./data:/var/lib/mysql:Z --env MARIADB_USER=dogs --env MARIADB_PASSWORD=byBpn9Znss8Onl3C --env MARIADB_DATABASE=dogs mariadb:11.8


# Run CLI

    docker run -it --network host --rm mariadb mariadb -h 127.0.0.1 -u dogs --password --database dogs --ssl-verify-server-cert=OFF

Try `\s`
and `exit`


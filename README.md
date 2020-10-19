# Scan for MySQL

Just a very basic example of connecting and getting version and protocol string at the moment.

TODO: Finish me

## Run MySQL

    docker run --name mysql -e MYSQL_ROOT_PASSWORD=mysecret -d -p 3306:3306 mysql:latest

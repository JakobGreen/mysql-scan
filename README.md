# Simple Scanning for MySQL

This tool will connect over tcp to given host and port to determine whether or not it is running MySQL. If it is indeed running MySQL the tool will output the information received from MySQL. If it is not a MySQL server then a non-zero exit code will be returned and error message with be place in stderr. A limitation of this tool is that it can only detect MySQL using handshake version 10.

Comments within the code got a bit verbose, but should explain things well enough.

## Building

There are no external dependencies. Build using the standard go build command.

    go build

## Testing

There are some units tests within sql_test.go which can be run using the go unit testing tool

    go test

To test the full solution we can run MySQL within Docker and try to connect.

Run MySQL:

    docker run --name mysql -e MYSQL_ROOT_PASSWORD=mysecret -d -p 3306:3306 mysql:latest

Run Scanner:

    ./mysql-scan -host 127.0.0.1:3306

package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	scanHost    string
	scanTimeout int
)

func parseCommandLine() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Tool for checking a given host and port for running MySQL\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&scanHost, "host", "127.0.0.1:3306", "Host and port to test for running MySQL server")
	flag.IntVar(&scanTimeout, "t", 1, "Dial timeout in seconds")
	flag.Parse()
}

func main() {
	parseCommandLine()

	if sql, err := DetectMySQL(scanHost, scanTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Detected MySQL:\n%s\n", sql.String())
	}
}

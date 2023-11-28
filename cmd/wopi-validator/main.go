package main

import (
	"flag"
	"fmt"
	"github.com/wkloucek/cs3-wopi-server/pkg/wopivalidator"
	"os"
)

func main() {
	var username = flag.String("u", "", "Username - mandatory")
	var password = flag.String("p", "", "Password - mandatory")
	var testGroup = flag.String("g", "", "Run only the tests in the specified group (cannot be used with testname)")
	var testName = flag.String("n", "", "Run only the test specified (cannot be used with testgroup)")
	var help = flag.Bool("help", false, "Show usage")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}
	if *username == "" || *password == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *testGroup != "" && *testName != "" {
		flag.Usage()
		os.Exit(1)
	}

	err := wopivalidator.Run(*username, *password, *testGroup, *testName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

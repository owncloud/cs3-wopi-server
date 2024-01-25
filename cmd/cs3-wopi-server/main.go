package main

import (
	"fmt"

	"github.com/owncloud/cs3-wopi-server/pkg/cs3wopiserver"
)

func main() {
	err := cs3wopiserver.Start()
	if err != nil {
		fmt.Println(err)
	}
}

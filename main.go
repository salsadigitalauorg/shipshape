package main

import (
	"fmt"
	"log"
	"os"
	"salsadigitalauorg/shipshape/pkg/shipshape"
)

func main() {

	c, err := shipshape.ReadAndParseConfig("shipshape.yml")
	if err != nil {
		log.Fatal(err)
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Printf("c:\n%+v\n\n", c)
}

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

	fmt.Printf("c: %+v\n\n", c)
	r := c.RunChecks()
	fmt.Printf("\nresults: %+v\n", r)
}

package main

import (
	"encoding/json"
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

	r := c.RunChecks()
	data, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(string(data))
}

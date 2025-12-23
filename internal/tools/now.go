package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Print(time.Now().UTC().Format(time.RFC3339))
}

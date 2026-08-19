package main

import (
	"fmt"
	"github.com/tmoson/go-gator/internal/config"
)

func main() {
	configuration := config.Read()
	configuration.SetUser("tyler")
	config2 := config.Read()
	fmt.Printf("%v", config2)
}

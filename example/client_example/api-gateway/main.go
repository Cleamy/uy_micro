package main

import (
	"fmt"

	"github.com/Cleamy/uy_micro"
)



func main() {
	// hook trigger
	uy_micro.Server.OnBootstrap(func () {
		fmt.Println("start-getway")
	})	
	
	uy_micro.Server.Run()
}

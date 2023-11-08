package main

import (
	"fmt"

	"gitlab.playlinks.co/micro/sloco.common.server-proxy/server"
)

func main() {
	if err := server.Start(":3250"); err != nil {
		fmt.Println(err.Error())
	}
}
package main

import (
	"github.com/solsteace/kochira/account"
	"github.com/solsteace/kochira/link"
)

func main() {
	done := make(chan struct{})

	link.LoadEnv()
	go link.RunApp()
	account.LoadEnv()
	go account.RunApp()

	<-done
}

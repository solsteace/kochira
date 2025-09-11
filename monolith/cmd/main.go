package main

import (
	"github.com/solsteace/kochira/account"
	"github.com/solsteace/kochira/link"
	"github.com/solsteace/kochira/subscription"
)

func main() {
	done := make(chan struct{})

	link.LoadEnv()
	go link.RunApp()
	account.LoadEnv()
	go account.RunApp()
	subscription.LoadEnv()
	go subscription.RunApp()

	<-done
}

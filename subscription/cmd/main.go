package main

import "github.com/solsteace/kochira/subscription"

func main() {
	subscription.LoadEnv()
	subscription.RunApp()
}

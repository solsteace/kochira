package main

import "github.com/solsteace/kochira/account"

func main() {
	account.LoadEnv()
	account.RunApp()
}

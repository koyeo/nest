package main

import "github.com/koyeo/yo/config"

func main() {

	err := config.Load("./nest.toml")
	if err != nil {
		panic(err)
	}

}

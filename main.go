package main

import (
	"fmt"
)

func main() {
	current_config, err := read()
	if err != nil {
		println(err)
		return
	}

	current_config.set_user("astor")
	updated_config, err := read()
	if err != nil {
		println(err)
		return
	}

	fmt.Println(updated_config)
}

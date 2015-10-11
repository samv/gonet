package main

import (
	"network/http"
	"path/filepath"
	"fmt"
)

func main() {
	path, err := filepath.Abs(".")
	if err != nil {
		fmt.Println("filepath abs:", err)
		return
	}
	http.SetDir(path)
	http.Run()
}

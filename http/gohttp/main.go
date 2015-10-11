package main

import (
	"fmt"
	"network/http"
	"path/filepath"
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

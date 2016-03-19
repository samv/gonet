package main

import (
	"fmt"
	"path/filepath"

	"github.com/hsheth2/gonet/http"
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

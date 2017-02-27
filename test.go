package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func list(){
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	})
}

func main(){
	list()
}

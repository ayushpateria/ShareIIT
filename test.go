package main

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
)
type File_s struct {
	name string
	hash string
	size int
}

files := []File_s

func list(){
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)
			file, err := os.Open(path)
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println(fi.Size())
			f := File_s{}
			f.name = fileInfo.Name()
			f.size = fileInfo.Size()
			
			fmt.Println("File name:", fileInfo.Name())
			fmt.Println("Size in bytes:", fileInfo.Size())

		return nil
	})
}

func main(){
	list()
}

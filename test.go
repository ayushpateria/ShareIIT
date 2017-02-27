package main

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
)

func list(){
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)
			file, err := os.Open(path)
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println(fi.Size())
			fmt.Println("File name:", fileInfo.Name())
			fmt.Println("Size in bytes:", fileInfo.Size())
			fmt.Println("Permissions:", fileInfo.Mode())
			//fmt.Println("Last modified:", fileInfo.ModTime())
			//fmt.Println("Is Directory: ", fileInfo.IsDir())
			//fmt.Printf("System interface type: %T\n", fileInfo.Sys())
			//fmt.Printf("System info: %+v\n\n", fileInfo.Sys())
		return nil
	})
}

func main(){
	list()
}

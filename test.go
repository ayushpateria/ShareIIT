package main

import (
	"fmt"
	"os"
	"path/filepath"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"io"
) 
type File_s struct {
	name string
	hash string
	size int64	
}

var files []File_s 

func hash_file_sha1(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnSHA1String string
	
	//Open the filepath passed by the argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}
	
	//Tell the program to call the following function when the current function returns
	defer file.Close()
	
	//Open a new SHA1 hash interface to write to
	hash := sha1.New()
	
	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}
	
	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]
	
	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)
	
	return returnSHA1String, nil
 
}
func list(){
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)
			file, err := os.Open(path)
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			if(fileInfo.Size() != 0){
				f := File_s{}
				f.name = fileInfo.Name()
				f.size = fileInfo.Size()
			
				hash,_ := hash_file_sha1(path)
				f.hash = hash
				files = append(files, f)
				fmt.Println("File name:", fileInfo.Name())
				fmt.Println("Size in bytes:", fileInfo.Size())
				fmt.Println("File Hash:", hash)
			}
 		return nil
	})
}

func main(){
	files = make([]File_s, 10)
	list()
}

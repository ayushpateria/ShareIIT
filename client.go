
package main

import (
	"fmt"
	"net"
	"os"
    "bufio"
    "io"
    "strconv"
	"strings"
	"encoding/json"
	)

type File_s struct {
	Name string
	Hash string
	Size float64	
  path string
}

var files []File_s 


//Define that the binairy data of the file will be sent 1024 bytes at a time
const BUFFERSIZE = 1024

func main() {
	connection, err := net.Dial("tcp", "localhost:3333")
	fmt.Println("Connected to server, start receiving the file name and file size")
	if err != nil {
		panic(err)
	}
	for {
	defer connection.Close()
	fmt.Println("Hi !! Enter your choice ----")
	fmt.Println("->>>>>>>>>>>Enter 1 for PRINTING all available files ")
	fmt.Println("->>>>>>>>>>>Enter 2 for DOWNLOADING a particular file ")
	fmt.Println("->>>>>>>>>>>Enter 3 for SEARCHING for a particular file ")
	
	// read in input from stdin
    reader := bufio.NewReader(os.Stdin)
    option, _ := reader.ReadString('\n')
    if option[:1] == "1" {
	    // send to socket
	    fmt.Fprintf(connection, option + "\n")
	    message, _ := bufio.NewReader(connection).ReadString('\n')
	    err := json.Unmarshal([]byte(message), &files)
	    if err != nil { } 
	    fmt.Println("FILEID              FILENAME                   FILESIZE (In MB) \n")

	    for  i, value := range files {
	    	fmt.Print((i+1))
	    	fmt.Print("."+value.Name+"           ")
	    	fmt.Print((value.Size)/1024.0)
	    	fmt.Println("  MB")
	    	}

	}else if option[:1] == "2"{
				fmt.Fprintf(connection, option + "\n")
				fmt.Println("Enter the FILEID for the file you want to download  ")
				reader := bufio.NewReader(os.Stdin)
    			option, _ := reader.ReadString('\n')
    			i, _:= strconv.Atoi(option[:1])
    			hash := files[i-1].Hash
    			fmt.Fprintf(connection, hash + "\n")
    			go recivefile(connection)
    		}
    	}
	}
	
func recivefile(connection net.Conn) {
			//Create buffer to read in the name and size of the file
			bufferFileName := make([]byte, 128)
			bufferFileSize := make([]byte, 20)
			//Get the filesize
			connection.Read(bufferFileSize)
			//Strip the ':' from the received size, convert it to a int64
			fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 20, 64)
			//Get the filename
			connection.Read(bufferFileName)
			//Strip the ':' once again but from the received file name now
			fileName := strings.Trim(string(bufferFileName), ":")
			//Create a new file to write in
			newFile, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer newFile.Close()
			//Create a variable to store in the total amount of data that we received already
			var receivedBytes int64
			//Start writing in the file
			for {
				if (fileSize - receivedBytes) < BUFFERSIZE {
					io.CopyN(newFile, connection, (fileSize - receivedBytes))
					//Empty the remaining bytes that we don't need from the network buffer
					connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
					//We are done writing the file, break out of the loop
					break
				}
				io.CopyN(newFile, connection, BUFFERSIZE)
				//Increment the counter
				receivedBytes += BUFFERSIZE
			}
			fmt.Println("Received file completely!")
	}

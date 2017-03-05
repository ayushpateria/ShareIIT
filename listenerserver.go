package main

import (
    "fmt"
    "net"
    "strconv"
    "os"
    "bufio"
    "encoding/json"
	  "path/filepath"
	  "crypto/sha1"
	  "encoding/hex"
	  "log"
	  "io"
)
//Define the size of how big the chunks of data will be send each time
const BUFFERSIZE = 1024

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

type File_s struct {
	Name string
	Hash string
	Size int64	
  path string
}

var files []File_s 

func main() {
    // Listen for incoming connections.
    l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes.
    defer l.Close()
    fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
    for {
        // Listen for an incoming connection.
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
        // Handle connections in a new goroutine.
        go handleRequest(conn)
    }
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
  // Make a buffer to hold incoming data.

  // Read the incoming connection into the buffer.
  choice, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  if choice[:1] == "1" {
            //conn.Write([]byte("im in 1 \n"))
            b := list()
            conn.Write([]byte(b+"\n"))
            conn.Close()
            //conn.Write([]byte(list()))
             
  }
  if choice[:1] == "2" { 
            hash ,err := bufio.NewReader(conn).ReadString('\n')
            if err != nil {
                  fmt.Println("Error reading:", err.Error())
                  }
            //conn.Write([]byte(hash))
            for _, f := range files {
              if (hash == f.Hash+"\n") {
                go sendFileToClient(conn, f.path)
              }
            } 
  }
   
}

//This function is to 'fill'
func fillString(retunString string, toLength int) string {
  for {
    lengtString := len(retunString)
    if lengtString < toLength {
      retunString = retunString + ":"
      continue
    }
    break
  }
  return retunString
}

func list() string{
  files = nil
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)
			file, err := os.Open(path)
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			if(fileInfo.Size() != 0 && !fileInfo.IsDir()){
				f := File_s{}
				f.Name = fileInfo.Name()
				f.Size = fileInfo.Size()
        f.path = path
				hash,_ := hash_file_sha1(path)
				f.Hash = hash
				files = append(files, f)
			}
 		return nil
	})
  
	b, err := json.Marshal(files)
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}


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

func sendFileToClient(connection net.Conn, path string) {
  fmt.Println("A client has connected!")
  //defer connection.Close()
  //Open the file that needs to be send to the client
  file, err := os.Open(path)                      //pass a hash ka return string i.e -filename in that sharedfolder
  if err != nil {
    fmt.Println(err)
    return
  }
  //Get the filename and filesize
  fileInfo, err := file.Stat()
  if err != nil {
    fmt.Println(err)
    return
  }
  fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 20), 20)
  fileName := fillString(fileInfo.Name(), 128)
  fmt.Println(string(fileName))
      fmt.Println(string(fileSize))
                //Send the file header first so the client knows the filename and how long it has to read the incomming file
  fmt.Println("Sending filename and filesize!")
                                                    //Write first 10 bytes to client telling them the filesize
  connection.Write([]byte(fileSize))
                                                    //Write 64 bytes to client containing the filename
  connection.Write([]byte(fileName))
                                                  //Initialize a buffer for reading parts of the file in
  sendBuffer := make([]byte, BUFFERSIZE)
                                                  //Start sending the file to the client
  fmt.Println("Start sending file!")
  for {
    _, err = file.Read(sendBuffer)
    if err == io.EOF {
      //End of file reached, break out of for loop
      break
    }
    connection.Write(sendBuffer)
  }
  fmt.Println("File has been sent, closing connection!")
  return
}
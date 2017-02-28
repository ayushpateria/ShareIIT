package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
)
//Define the size of how big the chunks of data will be send each time
const BUFFERSIZE = 1024

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

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
  for {
  // Read the incoming connection into the buffer.
  choice, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }
  
  if choice[:1] == "1" {
            conn.Write([]byte(" The files availabe are -----\n"))
            // list()
  }
  if choice[:1] == "2" { // input format : 2 [hash of the file]
            conn.Write([]byte(" im in choice 2-----\n"))
            hash := choice[2:len(choice)-1]
            conn.Write([]byte(hash))
            // Open the file and write its bytes to the connection.
             sendFileToClient(conn)

  }
   }
  //  conn.Close()
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


func sendFileToClient(connection net.Conn) {
  fmt.Println("A client has connected!")
  defer connection.Close()
  //Open the file that needs to be send to the client
  file, err := os.Open("aimssem4.png")                      //pass a hash ka return string i.e -filename in that sharedfolder
  if err != nil {
    fmt.Println(err)
    return
  }
  defer file.Close()
  //Get the filename and filesize
  fileInfo, err := file.Stat()
  if err != nil {
    fmt.Println(err)
    return
  }
  fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
  fileName := fillString(fileInfo.Name(), 64)
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
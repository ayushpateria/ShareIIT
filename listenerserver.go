package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
)

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
  }
   }
  //  conn.Close()
}
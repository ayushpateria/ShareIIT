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
  
  if choice == "1\n" {
            conn.Write([]byte(" The files availabe are -----\n"))
  }
  if choice == "2\n" {
            conn.Write([]byte(" im in choice 2-----\n"))
            filename , _ := bufio.NewReader(conn).ReadString('\n')
            conn.Write([]byte(filename[2:len(filename)]))
  }
  // Send a response back to person contacting us.
  }
  // Close the connection when you're done with it.
 //  conn.Close()
}
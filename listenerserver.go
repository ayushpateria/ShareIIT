package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Define the size of how big the chunks of data will be send each time
var BUFFERSIZE int64 = 81920

const (
	CONN_HOST = "0.0.0.0"
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
var INSERT_URL = "http://ayushpateria.com/ShareIIT/insert.php"
var myIP string

// Get preferred outbound ip of this machine. Taken from http://stackoverflow.com/a/37382208/921872
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func sendIP() {
	response, err := http.Get(INSERT_URL + "?IP=" + GetOutboundIP())
	if err != nil {
		fmt.Println(err)
	} else {
		defer response.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {

	sendIP()

	quit := make(chan struct{})

	ticker := time.NewTicker(300 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				sendIP()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
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
		conn.Write([]byte(b + "\n"))
		conn.Close()
		//conn.Write([]byte(list()))

	}
	if choice[:1] == "2" {
		if len(files) == 0 {
			list()
		}
		parms := strings.Fields(choice)
		hash := parms[1]

		for _, f := range files {
			if f.Hash == hash {
				start, _ := strconv.Atoi(parms[2])
				end, _ := strconv.Atoi(parms[3])
				sendFileToClient(conn, f.path, int64(start), int64(end))
				conn.Close()
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

func list() string {
	files = nil
	filepath.Walk("Shared", func(path string, info os.FileInfo, err error) error {
		//fmt.Println(path)
		file, err := os.Open(path)
		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		if fileInfo.Size() != 0 && !fileInfo.IsDir() {
			f := File_s{}
			f.Name = fileInfo.Name()
			f.Size = fileInfo.Size()
			f.path = path
			hash, _ := hash_file_sha1(f)
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

// Not hashing the complete file as it's slow.
func hash_file_sha1(f File_s) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnSHA1String string

	//Open the filepath passed by the argument and check for any error
	file, err := os.Open(f.path)
	if err != nil {
		return returnSHA1String, err
	}
	// Read first 20 bytes.
	b1 := make([]byte, 20)
	file.Read(b1)
	ss := f.Name + string(f.Size) + string(b1)
	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new SHA1 hash interface to write to
	hash := sha1.New()

	hash.Write([]byte(ss))
	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil

}

func sendFileToClient(connection net.Conn, path string, start int64, end int64) {
	fmt.Println("A client has connected!")
	//defer connection.Close()
	//Open the file that needs to be send to the client
	file, err := os.Open(path) //pass a hash ka return string i.e -filename in that sharedfolder
	if err != nil {
		fmt.Println(err)
		return
	}

	file.Seek(start, 0)

	r := bufio.NewReader(file)
	// make a buffer to keep chunks that are read
	buf := make([]byte, BUFFERSIZE)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}
		if n > int(end-start) {
			n = int(end - start)
		}
		connection.Write(buf[:n])
	}

	fmt.Println("File has been sent, closing connection!")
	return
}

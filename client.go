package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type File_s struct {
	Name string
	Hash string
	Size float64
	path string
	ip   string
}

var files []File_s
var ips []string
var LIST_URL = "http://ayushpateria.com/ShareIIT/list.php"

//Define that the binairy data of the file will be sent 1024 bytes at a time
const BUFFERSIZE = 81920
const NUM_THREADS = 4

/*
To know the servers which are active in the network we store it in a list on a website. This function fetches that list.
*/
func fetchIPS() {
	ips = nil
	response, err := http.Get(LIST_URL)
	if err != nil {
		fmt.Println(err)
	} else {
		defer response.Body.Close()
		r, _ := ioutil.ReadAll(response.Body)
		//to decode JSON unmarshal is used
		json.Unmarshal(r, &ips)
		if err != nil {
			fmt.Println(err)
		}
	}

}

/*
This function connects to a server and asks it for all the files it has.
*/
func updateList(ip string) {
	var lFiles []File_s
	connection, err := net.DialTimeout("tcp", ip+":3333", time.Duration(5)*time.Second)
	if err != nil {
		return
		//	panic(err)
	}
	defer connection.Close()
	fmt.Fprintf(connection, "1\n")
	message, _ := bufio.NewReader(connection).ReadString('\n')

	json.Unmarshal([]byte(message), &lFiles)
	for i, _ := range lFiles {
		lFiles[i].ip = ip
	}
	files = append(files, lFiles...)
}

/*
This function calls updateList on all the ips concurrently and creates one big list containing all the files.
*/
func createList() {
	files = nil
	fetchIPS()
	var wg sync.WaitGroup
	for _, ip := range ips {
		// Increment the WaitGroup counter.
		wg.Add(1)
		go func(ip string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			updateList(ip)
		}(ip)
	}
	// Wait for all Lists fetches to complete.
	wg.Wait()
}

func main() {

	fmt.Println("Welcome to ShareIIT! Your intra college file sharing hub.")

	fmt.Println("1. List all available files.")
	fmt.Println("2. Download a file,")
	fmt.Println("3. Search a file.")
	fmt.Println("0. Exit.")
	for {

		fmt.Print(">> ")

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		option, _ := reader.ReadString('\n')
		if option[:1] == "1" {
			createList()
			os.Mkdir("Shared", 0755)

			for i, value := range files {
				fmt.Print((i + 1))
				fmt.Print(". " + value.Name + "		")
				fmt.Print(value.Size / 1024)
				fmt.Println(" kb")
			}

		} else if option[:1] == "2" {

			if len(files) == 0 {
				createList()
			}
			fmt.Println("Enter the ID of the file from the list : ")
			var id int
			fmt.Scanf("%d", &id)
			if id > len(files) {
				fmt.Println("Please enter a valid ID.")
			} else {
				receivefile(id - 1)
			}
		} else if option[:1] == "3" {

			createList()

			fmt.Print("Enter file name : ")
			reader := bufio.NewReader(os.Stdin)
			filename, _ := reader.ReadString('\n')
			flag := 0
			for i, value := range files {
				if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
					fmt.Print((i + 1))
					fmt.Print(". " + value.Name + "		")
					fmt.Print(value.Size)
					fmt.Println(" kb")
					flag = 1
				}
			}
			if flag == 0 {
				fmt.Print("No items match your search.\n")
			}
		} else if option[:1] == "0" {
			break
		} else {
			fmt.Println("Please enter a valid choice.")
		}
	}

}

var totalBytesTransferred int64
var percentage int
var count int

/*
This function calculates the percentage of file which has been downloaded and prints it. It runs every 2 or 1 second.
*/
func DownloadProgress(size float64, i int) {
	t := time.Now()
	count = 0
	var temp int64
	for totalBytesTransferred < int64(size) {

		speed := ((totalBytesTransferred) / 1024) / int64(math.Ceil(time.Since(t).Seconds()))
		percentage = int((totalBytesTransferred) * 100 / int64(size))
		if temp == totalBytesTransferred {
			count++
		} else {
			count = 0
		}
		if count == 5 { // If even after 5 (or 10) seconds the download percentage stays same, we change the server.
			fmt.Println("\nSeems like there is a problem with the connection. Please wait while we see if we can connect you to some other server.")

			getBackConnection(i)
			break
		}
		out := "\rFinished : " + strconv.Itoa(percentage) + "%, Avg. Speed : " + strconv.Itoa(int(speed)) + " kb/s"
		os.Stdout.Write([]byte(out))

		temp = totalBytesTransferred

		if size > 200*1024*1024 {
			time.Sleep(time.Second * 2) //For file size greater than 200MB
		} else {
			time.Sleep(time.Second * 1)
		}

	}
}

type indexrange struct {
	startindex int64
}

var pointer [NUM_THREADS]indexrange //Initialise to the number of threads pointer which points to point where it lost connection

/*
This function downloads the file by breaking in NUM_THREADS independent pieces. Each piece is downloaded concurrently and is written onto the output file.
*/
func receivefile(i int) {

	file := files[i]
	fmt.Println("Downloading " + file.Name + ", this may take a while.")
	hash := files[i].Hash

	size := file.Size

	// Waitgroup waits for a collection of goroutines to complete
	var wg sync.WaitGroup
	inc := int64(math.Ceil(size / float64(NUM_THREADS)))
	var j int64

	var mutex = &sync.Mutex{}
	pointer[0].startindex = 0
	go DownloadProgress(size, i)
	threadindex := 0
	for j = 0; j < int64(math.Ceil(size)); j = j + inc {
		start := j
		end := start + inc
		start = pointer[threadindex].startindex

		if end > int64(math.Ceil(size)) {
			end = int64(math.Ceil(size))
		}

		wg.Add(1)
		go func(start int64, end int64, threadindex int) {

			defer wg.Done()
			var receivedBytes int64
			newFile, err := os.Create(file.Name)
			if err != nil {
				panic(err)
			}
			defer newFile.Close()
			connection, err := net.Dial("tcp", file.ip+":3333")
			if err != nil {
				panic(err)
			}
			defer connection.Close()

			newFile.Seek(start, 0)
			fmt.Fprintf(connection, "2 %s %d %d\n", hash, start, end)

			//Start writing in the file
			for {
				n, err := io.CopyN(newFile, connection, BUFFERSIZE)

				mutex.Lock()
				receivedBytes += n
				pointer[threadindex].startindex = receivedBytes
				totalBytesTransferred += n
				mutex.Unlock()

				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if n == 0 {
					break
				}
			}

		}(start, end, threadindex)
		threadindex++
	}

	// Wait until all parts have been finished downloading.
	wg.Wait()

	fmt.Println("Received file completely!")

}

/*
In the case there was a connection drop in between the file transfer, this function tries to find the find the same file in other active servers.
If it successfully finds a match, it will reestablish the connection with the new server and resume the download from the same point.
*/
func getBackConnection(i int) {

	file := files[i]
	filename := file.Name
	createList()
	var flag int
	flag = 0
	for i, value := range files {

		if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
			fmt.Println("Resuming your download...")
			receivefile(i)
			flag++
		}
		i = i + 1
	}
	if flag == 0 {
		fmt.Println("Sorry, it's not possible to continue the download at the moment.")
	}
}

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

func fetchIPS() {
	ips = nil
	response, err := http.Get(LIST_URL)
	if err != nil {
		fmt.Println(err)
	} else {
		defer response.Body.Close()
		r, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(r, &ips)
		if err != nil {
			fmt.Println(err)
		}
	}

}

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

			for i, value := range files {
				fmt.Print((i + 1))
				fmt.Print(". " + value.Name + "		")
				fmt.Print(int(math.Ceil(value.Size / 1024)))
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
			//fmt.Print(filename)
			for i, value := range files {
				//fmt.Println(value.Name)
				if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
					fmt.Print((i + 1))
					fmt.Print(". " + value.Name + "		")
					fmt.Print(int(value.Size))
					fmt.Println(" kb")
					flag = 1
				}
			}
			if flag == 0 {
				fmt.Print("No items match your search.\n")
			}
		} else if option[:1] == "0" {
			break
		}
	}

}

func receivefile(i int) {

	NUM_THREADS := 5

	file := files[i]
	fmt.Println("Downloading " + file.Name + ", this may take a while.")

	hash := files[i].Hash

	size := file.Size

	var wg sync.WaitGroup
	inc := int64(math.Ceil(size / float64(NUM_THREADS)))
	var j int64
	t := time.Now()
	threadsRemaining := NUM_THREADS
	var mutex = &sync.Mutex{}
	var bytesTransferred int64

	// We divivde each file in pieces and give them to each thread.
	for j = 0; j < int64(math.Ceil(size)); j = j + inc {
		start := j
		end := start + inc
		if end > int64(math.Ceil(size)) {
			end = int64(math.Ceil(size))
		}

		wg.Add(1)
		go func(start int64, end int64) {

			defer wg.Done()

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

			newFile.Seek(start, 0) // Since this thread is only downloading after 'start' bytes, we seek the file to that position.
			fmt.Fprintf(connection, "2 %s %d %d\n", hash, start, end)

			//Start writing in the file
			for {

				n, err := io.CopyN(newFile, connection, BUFFERSIZE)
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if n == 0 {
					break
				}
			}
			mutex.Lock() // To calculate progress.
			bytesTransferred += end - start
			threadsRemaining--
			speed := (bytesTransferred / 1024) / int64(math.Ceil(time.Since(t).Seconds()))
			out := "Finished : " + strconv.Itoa((NUM_THREADS-threadsRemaining)*100/NUM_THREADS) + "%, Avg. Speed : " + strconv.Itoa(int(speed)) + " kb/s\n"
			os.Stdout.Write([]byte(out))
			os.Stdout.Sync()
			mutex.Unlock()

		}(start, end)
	}

	// Wait until all parts have been finished downloading.
	wg.Wait()
	fmt.Println("Received file completely!")
}

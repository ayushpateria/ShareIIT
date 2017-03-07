package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"sync"
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
	connection, err := net.Dial("tcp", ip+":3333")
	if err != nil {
		panic(err)
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
				fmt.Print(value.Size)
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
				recivefile(id - 1)
			}
		} else if option[:1] == "0" {
			break
		}
	}

}

func recivefile(i int) {

	NUM_THREADS := 5

	file := files[i]
	fmt.Println("Downloading " + file.Name + ", this may take a while.")

	hash := files[i].Hash

	size := file.Size

	newFile, err := os.Create(file.Name)
	if err != nil {
		panic(err)
	}
	defer newFile.Close()

	var wg sync.WaitGroup
	inc := int64(math.Ceil(size / float64(NUM_THREADS)))
	var j int64
	for j = 0; j < int64(math.Ceil(size)); j = j + inc {
		start := j
		end := start + inc
		if end > int64(math.Ceil(size)) {
			end = int64(math.Ceil(size))
		}

		go func() {

			wg.Add(1)
			defer wg.Done()

			connection, err := net.Dial("tcp", file.ip+":3333")
			if err != nil {
				panic(err)
			}
			defer connection.Close()

			newFile.Seek(start, 0)
			fmt.Fprintf(connection, "2 %s %d %d\n", hash, start, end)
			fmt.Printf("2 %s %d %d\n", hash, start, end)

			var receivedBytes int64
			//Start writing in the file
			for {
				if (end - start - receivedBytes) < BUFFERSIZE {
					io.CopyN(newFile, connection, (end - start - receivedBytes))
					//We are done writing the file, break out of the loop
					break
				}
				io.CopyN(newFile, connection, BUFFERSIZE)
				//Increment the counter
				receivedBytes += BUFFERSIZE
			}
			fmt.Printf("Done thread " + string(j))
		}()
	}

	// Wait until all parts have been finished downloading.
	wg.Wait()
	fmt.Println("Received file completely!")
}

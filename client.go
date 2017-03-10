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
		//to decode JSON unmarshal is used
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
			os.Mkdir("Shared", 0755)

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
			fmt.Println("INVALID")
			fmt.Println("1. List all available files.")
			fmt.Println("2. Download a file,")
			fmt.Println("3. Search a file.")
			fmt.Println("0. Exit.")
		}
	}

}
	var TotalbytesTransferred int64
	var percentage int 
	var count int

	func DownloadProgress(size float64 ,i int ){								//Function which shows download progress
		t := time.Now()
		//count=0
		for(TotalbytesTransferred < int64(size) && size > 200000 ){				//For file size greater than 200MB
			time.Sleep( time.Second * 2)
			temp := percentage
			speed := ((TotalbytesTransferred) / 1024) / int64(math.Ceil(time.Since(t).Seconds()))
				if	(TotalbytesTransferred < int64(size) ){
					percentage = int((TotalbytesTransferred)*100/int64(size))
					if temp ==percentage {
						count++
					}
					if count ==5{
						fmt.Println("Connection lost .....Retrieving wait a min")
						getBackConnection(i)
						break
					}
					out := "Finished : " + strconv.Itoa(percentage) + "%, Avg. Speed : " + strconv.Itoa(int(speed)) + " kb/s\n"
					os.Stdout.Write([]byte(out))
					}
			}
		for(TotalbytesTransferred < int64(size) && size <= 200000 ){				//For filze size less than 200MB

			time.Sleep(  time.Second * 1)
			temp := percentage
			speed := ((TotalbytesTransferred) / 1024) / int64(math.Ceil(time.Since(t).Seconds()))
				if	(TotalbytesTransferred < int64(size) ){
					percentage = int((TotalbytesTransferred)*100/int64(size))
					if temp ==percentage {
						count++
					}
					if count ==5{													// which checks if connection is lost
						fmt.Println("Connection lost .....Retrieving wait a min")
						getBackConnection(i)
						break
					}

					out := "Finished : " + strconv.Itoa(int((TotalbytesTransferred)*100/int64(size))) + "%, Avg. Speed : " + strconv.Itoa(int(speed)) + " kb/s\n"
					os.Stdout.Write([]byte(out))
					}
			}

	}

func getBackConnection(i int) {											//Function which checks for same file in another 
																		// server and get backs connection

				file := files[i]
				filename := file.Name
				createList()
				var flag int
				flag=0
				for i, value := range files {
					
					if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
						fmt.Println("Connection Found....")
						receivefile(i)
						flag++
					}
					i=i+1
				}
					if flag == 0 {
						fmt.Println("Sorry ....Connection permanantely lost.Download Failed :(")
						fmt.Println()
						main()
					}
		}

type indexrange struct{
	 startindex int64
}

var pointer [4]indexrange												//Initialise to the number of threads pointer
																		//Which points to point where it lost connection

func receivefile(i int) {

	//os.Chdir("Shared")

	NUM_THREADS := 4

	file := files[i]
	fmt.Println("Downloading "  +file.Name + ", this may take a while.")
	hash := files[i].Hash

	size := file.Size
	//waitgroup waits for a collection of goroutines to complete
	var wg sync.WaitGroup
	inc := int64(math.Ceil(size / float64(NUM_THREADS)))
	var j int64
	
	//threadsRemaining := NUM_THREADS
	//var mutex = &sync.Mutex{}
	pointer[0].startindex=0
	go DownloadProgress(size, i)
	threadindex :=0
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
				if (end-start-receivedBytes) < BUFFERSIZE {
						n, err := io.CopyN(newFile, connection, (end-start-receivedBytes))
															//Empty the remaining bytes that we don't need from the network buffer
						if err != nil && err != io.EOF {
							log.Fatal(err)
						}
						//mutex.Lock()
						TotalbytesTransferred += n
						//mutex.Unlock()
						pointer[threadindex].startindex += n
						break
					}
				n, err := io.CopyN(newFile, connection, BUFFERSIZE)
		        receivedBytes += n
				pointer[threadindex].startindex=receivedBytes
				//fmt.Println("threadindex  pointer ",threadindex,receivedBytes)				
				//mutex.Lock()
				TotalbytesTransferred += n
				//mutex.Unlock()
				
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if n == 0 {
					break
				}
			}
			
		}(start, end,threadindex)
		threadindex++
	}

	// Wait until all parts have been finished downloading.
	wg.Wait()
	fmt.Println("Received file completely!")
	//os.Chdir("..")

	main()
}

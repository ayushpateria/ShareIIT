package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	//"log"
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

func main() {																		//Main Function

	 f, _ := os.OpenFile("output.txt",os.O_RDWR |os.O_CREATE, 0755)					//Create a outputfile
	 flag1:=0
	 err1 := os.Remove("output.txt")												//Empty the output file everytime program runs
	 
	 channel_name := make(chan int)													//Channel for synchronising Downloads
	 
      if err1 != nil {
          fmt.Println(err1)
          return
      }
/*
The function of output file is to write the download progress and write back to terminal if user needs
*/
	f, err := os.OpenFile("output.txt",os.O_RDWR |os.O_CREATE, 0755)				//After Deleting output file , create a fresh file
	if err != nil {
		        fmt.Print(err)
		    }

	fmt.Println("Welcome to ShareIIT! Your intra college file sharing hub.")

	fmt.Println("1. List all available files.")
	fmt.Println("2. Download a file,")
	fmt.Println("3. Search a file.")
	fmt.Println("5.Show Download status")
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
				fmt.Println("Downloading " + files[id-1].Name + ", this may take a while.")
				fmt.Println("MeanWhile relax or check fow another download")
				go receivefile(id - 1,f,channel_name)
				if flag1==0 {
					channel_name <- 1											// pass message to channel only at start of program
					flag1 =1
				}
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
		} else if option[:1] == "5" {
			DisplayDownloadList()												 
		} else {
			fmt.Println("Please enter a valid choice.")
		}
	}

}
/*
Function which Shows output to console periodically , the status of files downloading
*/

func DisplayDownloadList() {											

    b, err := ioutil.ReadFile("output.txt") 							// just pass the file name
    if err != nil {
		        fmt.Print(err)
		    }
	str := string(b) 													// convert content to a 'string'
    	if len(str)== 0 {
    		fmt.Println("No Downloads Until Now")						
    		return
    	}
	zero := "?"
	go func() {
		fmt.Scanln(&zero)
	}()
    for  {
		b, err := ioutil.ReadFile("output.txt") 						// just pass the file name
    	if err != nil {
		        fmt.Print(err)
		    }  
    	str := string(b) 												// convert content to a 'string'
    	fmt.Println("--------------------")
    	fmt.Print("The list of Downloads in this Session are.....")
    	fmt.Println(str)												// print the content as a 'string'
    	fmt.Println("--------------------")
 	    fmt.Println("Press any symbol followed by ENTER to return to main menu")
 	    if zero != "?"{
 	    	break
 	    }
 	    time.Sleep(time.Second * 2)
    }
    fmt.Println()
    return
}
																																																																																																														
var totalBytesTransferred int64
var percentage int
var count int
var linenumber int
var work_mux sync.Mutex 												//Mutex locks for writing to output file 

/*
This function calculates the percentage of file which has been downloaded and prints it. It runs every 2 or 1 second.
*/
func DownloadProgress(size float64, i int, f *os.File, channel_name chan int) {
	t := time.Now()
	var dummy string
	count = 0
	var temp int64
	for i := 0; i <= linenumber	; i++ {									//Move towards the end of the file
		fmt.Fscanln(f,&dummy)
		}
		fmt.Fprintln(f,"\n")
		pos, _ := f.Seek(0, os.SEEK_CUR)								//Get the position of file pointer 
		linenumber += 2
	for totalBytesTransferred < int64(size)*4 {

		speed := ((totalBytesTransferred) / 1024) / int64(math.Ceil(time.Since(t).Seconds()))
		percentage = int(((totalBytesTransferred) * 100) / (int64(size)*4))
		if temp == totalBytesTransferred {
			count++
		} else {
			count = 0
		}
		if count == 5 { // If even after 5 (or 10) seconds the download percentage stays same, we change the server.
			fmt.Println("\nSeems like there is a problem with the connection. Please wait while we see if we can connect you to some other server.")
			work_mux.Lock()
			f.Seek(pos,0)
			fmt.Fprint(f,files[i].Name+"\n")
			out := "Connection Problem............File Downloaded incompletely"
			fmt.Fprint(f,out)
			work_mux.Unlock()
			getBackConnection(i,f, channel_name)
			return
		}

		work_mux.Lock()														//Write the status of Download to Output file
		f.Seek(pos,0)
		fmt.Fprint(f,files[i].Name+"\n")
		out := "\rFinished : " + strconv.Itoa(percentage) + "%, Avg. Speed : " + strconv.Itoa(int(speed)) + " kb/s"
		fmt.Fprint(f,out)
		work_mux.Unlock()

		temp = totalBytesTransferred

		if size > 200*1024*1024 {
			time.Sleep(time.Second * 2) //For file size greater than 200MB
		} else {
			time.Sleep(time.Second * 1)
		}

	}
	work_mux.Lock()
	f.Seek(pos,0)
	fmt.Fprint(f,files[i].Name+"\n")
	str:= "Received file completely!  Check In your Working Files Directory"	
	fmt.Fprint(f,str)
	work_mux.Unlock()
	channel_name <- 1														//If one thread completes,then start next thread by passing
																			//message.This is to establish a fast connection ,rather than 
																			// dividing the band width making it slow connection
	return
	
}

type indexrange struct {
	startindex int64
}

var pointer [NUM_THREADS]indexrange                        //Initialise to the number of threads pointer which points to point where it lost connection

/*
This function downloads the file by breaking in NUM_THREADS independent pieces. Each piece is downloaded concurrently and is written onto the output file.
*/
func receivefile(i int, f *os.File, channel_name chan int) {

	 
	 _ = <-channel_name										//If the thread recieves message then Go and Download, otherwise wait here....

	file := files[i]
	hash := files[i].Hash

	size := file.Size

	// Waitgroup waits for a collection of goroutines to complete
	var wg sync.WaitGroup
	inc := int64(math.Ceil(size / float64(NUM_THREADS)))
	var j int64

	var mutex = &sync.Mutex{}
	totalBytesTransferred=totalBytesTransferred-pointer[0].startindex
	pointer[0].startindex = 0

	go DownloadProgress(size, i, f, channel_name )

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

			// Set a deadline for reading. Read operation will fail if data is received after deadline.
			//This is for detecting poor connections or Server TimeOuts and disconnecting them ,
			// or connect to high speed server
			for {
				timeoutDuration := 10* time.Second
				connection.SetReadDeadline(time.Now().Add(timeoutDuration))
																				//	Start writing in the file
				n, err := io.CopyN(newFile, connection, BUFFERSIZE)
				mutex.Lock()
				receivedBytes += n
				pointer[threadindex].startindex = receivedBytes
				totalBytesTransferred += n
				mutex.Unlock()

				if err != nil && err != io.EOF {
					fmt.Println()
					fmt.Println("Poor Connection Detected....Exiting")				//Exit if Poor connection
					os.Exit(3)
				}
				if n == 0 {
					break
				}
			}
			return

		}(start, end, threadindex)
		threadindex++
	}
	// Wait until all parts have been finished downloading.
	wg.Wait()
	
}

/*
In the case there was a connection drop in between the file transfer, this function tries to find the find the same file in other active servers.
If it successfully finds a match, it will reestablish the connection with the new server and resume the download from the same point.
*/
func getBackConnection(i int, f *os.File, channel_name chan int) {

	file := files[i]
	filename := file.Name
	createList()
	var flag int
	flag = 0
	for i, value := range files {

		if strings.Contains(value.Name, strings.Trim(filename, "\n")) {
			fmt.Println("Resuming your download...")
			fmt.Println(">>")
			go receivefile(i , f, channel_name)
			channel_name <- 1										//Pass message to resume Download
			flag++
		}
		i = i + 1
	}
	if flag == 0 {
		fmt.Println("Sorry, it's not possible to continue the download at the moment.")
		fmt.Print(">> ")
	}
}

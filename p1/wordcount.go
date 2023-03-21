package main

/*references used, other than official go documentation
-- https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
-- https://gosamples.dev/remove-non-alphanumeric/
-- https://stackoverflow.com/questions/24073697/how-to-find-out-the-number-of-cpus-in-go-lang
-- https://codewithyury.com/golang-wait-for-all-goroutines-to-finish/
-- https://gist.github.com/mattes/d13e273314c3b3ade33f
-- https://www.includehelp.com/golang/how-to-find-the-number-of-cpu-cores-used-by-the-current-process.aspx
-- https://socketloop.com/tutorials/golang-how-to-split-or-chunking-a-file-to-smaller-pieces 
*/

import (
	"bufio"
	"fmt"
	"bytes"
	"log"
	"os"
	"regexp"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)
var MAXWORDSIZE int = 30
var READSIZE int = 10000
var NONLETTER = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var M_OUTPUT = "/multi.txt"
var S_OUTPUT = "/single.txt"

func readFiles(directory string)([]string, int) {
	filepaths, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal("Error: could not read files.")
	}
	for _, file := range(filepaths) {
		if file.IsDir() {
			log.Fatal("One of the files is a directory")
		}
	}
	files := make([]string, len(filepaths))
	for i, fname := range(filepaths) {
		files[i] = directory + "/" + fname.Name()
	}
	return files, 0
}


/**
Reads each file in a dictionary
**/
func singleThreaded(files []string) {
	counts := countSomeFiles(files)
	err := writeMapToFile(os.Args[1] + S_OUTPUT, counts)
	if err != nil {
		log.Fatal(err)
	}
}

type SafeMap struct {
	wordmap map[string] int
	lock sync.Mutex
} 



func grabSomeText(file *os.File, readSize int) ([]byte, error) {
	buff := make([]byte, readSize, readSize + MAXWORDSIZE) //avoid reallocation
	_, err := file.Read(buff) // read the length of buffer from file
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF //reached end of file
		} else {
			log.Fatal(err)
		}
	}
	extra, err2 := getStr(file) // read till next space if present
	buff = append(buff, extra...)
	if err2 == io.EOF {
		return buff, io.EOF
	}
	return buff, nil
}
// Reads from file up to encountering whitespace. Used to account for words
// that may be cut off when reading by amount of bytes. Returns string
func getStr(file *os.File) (string, error) {
	extra := strings.Builder{}
	b := make([]byte, 1)
	for {
		_, err := file.Read(b)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		} else if err == io.EOF {
			return extra.String(), io.EOF
		} else if b[0] == ' ' {
			return extra.String(), nil
		} else {
			extra.Grow(1)
			extra.WriteByte(b[0])
		}
	}
}


func multiThreaded2(files []string) {
	globalMap := SafeMap{wordmap:make(map[string]int)}
	var waitgroup sync.WaitGroup
	for _, filename := range(files) {
		file, err := os.Open(filename) //file pointer
		if err != nil {
			file.Close()
			log.Fatal(err)
		}
		txt, err2 := grabSomeText(file, READSIZE) //read from this file
		for text, err3 := txt, err2; err3 == nil; text, err3 = grabSomeText(file, READSIZE){
			waitgroup.Add(1)
			go wordRoutine(text, &globalMap, &waitgroup) //routine for this chunk.
			/*I am curious about the "lifespan" of this byte array (text).
			//Does go compiler's garbage collector free the memory after the routine completes?
			//It could technically still be used in this function
			but since it's not, I am wondering if garbage collector takes care of it*/
		}
		file.Close()
	}
	waitgroup.Wait()
	globalMap.lock.Lock() //grab the global map's lock.
	err := writeMapToFile(os.Args[1] + M_OUTPUT, &globalMap.wordmap) //write safe map to file
	if err != nil {
		log.Fatal(err)
	}	
	globalMap.lock.Unlock() //not really necessary since the safemap's lifespan ends with return
	//but I include it anyway because each lock should come with an unlock
}

func wordRoutine(text []byte, globalMap *SafeMap, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()
	localMap := make(map[string]int)
	reader := bytes.NewReader(text)
	scanner := bufio.NewScanner(reader) //buffered i/o: creates a pipe for reading
	scanner.Split(bufio.ScanWords) //break reading pattern into words
	for scanner.Scan() { //reads until EOF OR until the limit
		word := scanner.Text()
		word = strings.ToLower(word) //lowercase-ify
		word = NONLETTER.ReplaceAllString(word, " ") //get rid of extra characters
		words := strings.Split(word, " ") //split words by char
		for _, wd := range(words) {
			wd2 := NONLETTER.ReplaceAllString(wd, "") //get rid of spaces
			localMap[wd2] = localMap[wd2] + 1 //increment word count in the dictionary
		}
	}
	globalMap.lock.Lock()
	for wrd, count := range(localMap) {
		globalMap.wordmap[wrd] = globalMap.wordmap[wrd] + count
	}
	globalMap.lock.Unlock()
}

func countSomeFiles(files []string) *map[string]int {
	counts := make(map[string] int) //store word counts by key which is the word itself
	var nonLetter = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	for _, filename := range(files) {
		file, err := os.Open(filename) //file pointer
		if err != nil {
			log.Println(err)
			fmt.Println("error opening file:", filename)
			return nil
		}
		defer file.Close() //make sure this is closed before return
		scanner := bufio.NewScanner(file) //buffered i/o: creates a pipe for reading
		scanner.Split(bufio.ScanWords) //break reading pattern into words
		for scanner.Scan() { //reads until EOF OR until the limit
			word := scanner.Text()
			word = strings.ToLower(word) //lowercase-ify
			word = nonLetter.ReplaceAllString(word, " ") //get rid of extra characters
			words := strings.Split(word, " ") //split words by char
			for _, wd := range(words) {
				wd2 := nonLetter.ReplaceAllString(wd, "") //get rid of spaces
				counts[wd2] = counts[wd2] + 1 //increment word count in the dictionary
			}
		}
	}
	return &counts
}


/**
Writes a Map of string:int to file, handles open and close
**/
func writeMapToFile(filename string, counts *map[string]int) error {
	singleOut, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file")
	}
	defer singleOut.Close() //make sure file closes before return.
	writer := bufio.NewWriter(singleOut)
	total := 0
	for key, count := range(*counts) {
		str := key + " " + strconv.Itoa(count) + "\n"
		_, err := writer.WriteString(str)
		if err != nil {
			return fmt.Errorf("error writing to output file")
		}
		writer.Flush()
		total += 1
	}
	fmt.Println("total is", total)
	return nil
}

func countRoutine(file *os.File, c chan *map[string]int, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()
	//fmt.Println("thread is going!")
	var nonLetter = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	scanner := bufio.NewScanner(file) //buffered i/o: creates a pipe for reading
	counts := make(map[string]int) //store word counts by key which is the word itself
		scanner.Split(bufio.ScanWords) //break reading pattern into words
		for scanner.Scan() { //reads until EOF
			word := scanner.Text() //get a word
			word = strings.ToLower(word) //lowercase-ify the word
			word = nonLetter.ReplaceAllString(word, " ") //replace extra characters with space
			words := strings.Split(word, " ") //split words by s
			for _, wd := range(words) {
				wd2 := nonLetter.ReplaceAllString(wd, "") //get rid of spaces
				counts[wd2] = counts[wd2] + 1 //increment word count in the dictionary
			}
		}
	c <- &counts //once you're done counting, add the completed map to the channel.
}

/**
determine the size of each file and break large ones into two.
**/
func multiThreaded(files []string) {
	//divide up the files here.
	var waitgroup sync.WaitGroup //waiting for completion of routines
	l := len(files)
	c := make(chan *map[string]int, l) //store results here
	for _, filename := range(files) {
		file, err := os.Open(filename) //file pointer
		if err != nil {
			log.Fatal(err)
		} else {
			waitgroup.Add(1) //add this routine to wait on.
			go countRoutine(file, c, &waitgroup) //start a
		}
	}
	waitgroup.Wait() //wait for all threads to complete.
	close(c) //done with channel	

	totals := make(map[string]int)
	for countMap := range(c) { //for each routine's individual map...
		for word, count := range(*countMap) { //dereferenced
			totals[word] = totals[word] + count
		}
	}
	err := writeMapToFile(os.Args[1] + M_OUTPUT, &totals)
	if err != nil {
		log.Fatal(err)
	}	
}

func main() {
	
	if len(os.Args) < 2 {
		log.Fatal("please specify an output directory")
	}
	output_dir := os.Args[1]
	if _, err := os.Stat(output_dir); os.IsNotExist(err) {
		err := os.Mkdir(output_dir, os.ModePerm)
		if err != nil {
        	log.Fatal(err)
    	}
	}
	files, error := readFiles("./input")
	switch error {
	case 1:
		log.Fatal("Error reading file names from directory") //calls to panic
	case 2:
		log.Fatal("Please store only text files in your directory")
	default:
		start0 := time.Now()
		singleThreaded(files)
		fmt.Println(time.Since(start0))

		start1 := time.Now()
		multiThreaded(files)
		fmt.Println(time.Since(start1))

		start2 := time.Now()
		multiThreaded2(files)
		fmt.Println(time.Since(start2))
	}
	
	// TODO: add argument processing and run both single-threaded and multi-threaded functions
}
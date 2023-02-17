package main

/*references used, other than official go documentation
-- https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
-- https://gosamples.dev/remove-non-alphanumeric/
-- https://stackoverflow.com/questions/24073697/how-to-find-out-the-number-of-cpus-in-go-lang
-- https://codewithyury.com/golang-wait-for-all-goroutines-to-finish/
-- https://gist.github.com/mattes/d13e273314c3b3ade33f
-- https://www.includehelp.com/golang/how-to-find-the-number-of-cpu-cores-used-by-the-current-process.aspx
*/

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"runtime"
)


func readFiles(directory string)([]string, int) {
	filepaths, err := os.ReadDir(directory)
	if err != nil {
		log.Println("Error: could not read files.")
		return nil, 1
	}
	for _, file := range(filepaths) {
		if file.IsDir() {
			log.Println("One of the files is a directory")
			return nil, 2
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
	err := writeMapToFile(os.Args[1] + "/single.txt", counts)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
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
		for scanner.Scan() { //reads until EOF
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
	fmt.Println("thread is going!")
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

func multi2(files []string) {
	numRoutines := runtime.NumCPU() //base the number of threads on the number of CPUs you have.
	l := len(files)
	perEach := l / numRoutines
	/**
	essentially, we predetermine the number of threads, and assign some files per thread.
	we sort the files by size such that the smaller files are first
	this way, the routine which handles the smallest chunk is handling bigger files.
	**/
	if (l % numRoutines) > 0 {
		perEach += 1 //round up.
	}
	if perEach <= 1 { //more routines than files, just assign one per each.
		multiThreaded(files)
		return
	}
	var waitgroup sync.WaitGroup //waiting for completion of routines
	c := make(chan *map[string]int, l) //store results here
	current := 0
	for i := 0; i < numRoutines; i ++ { //send out a routine which acts on each set of files.
		waitgroup.Add(1)
		if current + perEach >= l {
			go manage(c, files[current:], &waitgroup)
		} else { //don't exceed file array size.
			go manage(c, files[current:current+perEach], &waitgroup)
		}
		current += perEach
	}
	waitgroup.Wait()
	close(c)
	totals := make(map[string]int)
	for countMap := range(c) { //for each routine's individual map...
		for word, count := range(*countMap) { //dereferenced
			totals[word] = totals[word] + count
		}
	}
	err := writeMapToFile(os.Args[1] + "/multi2.txt", &totals)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}	
}

func manage(c chan *map[string]int, files []string, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()
	counts := countSomeFiles(files) //returns a pointer to the array
	c <- counts
} 

func multiThreaded(files []string) {
	//divide up the files here.
	var waitgroup sync.WaitGroup //waiting for completion of routines
	l := len(files)
	c := make(chan *map[string]int, l) //store results here
	for _, filename := range(files) {
		file, err := os.Open(filename) //file pointer
		if err != nil {
			log.Println(err)
			fmt.Println("error opening file:", filename)
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
	err := writeMapToFile(os.Args[1] + "/multi.txt", &totals)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}	
}


func main() {
	if len(os.Args) < 2 {
		log.Fatal("please specify an output directory")
	}
	output_dir := os.Args[1]
	if _, err := os.Stat(output_dir); os.IsNotExist(err) { //doe
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
		singleThreaded(files)
		multiThreaded(files)
		multi2(files)
	}
	
	// TODO: add argument processing and run both single-threaded and multi-threaded functions
}
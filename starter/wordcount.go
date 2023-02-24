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
	"log"
	"os"
	"regexp"
	"io"
	"strconv"
	"strings"
	"sync"
	"runtime"
	"time"
)

/**
Wrapper for the readerAt so that bufio will start from an offset.
**/
type OffsetReader struct {
	r io.ReaderAt
	offset int64
}
func (readerfrom OffsetReader) Read(p []byte) (bytesRead int, err error) {
	read, err := readerfrom.r.ReadAt(p, readerfrom.offset)
	if err != nil {
		return read, err
	}
	readerfrom.offset = readerfrom.offset + int64(read) //move the offset.
	return read, err

}


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

/**reads from a chunk of a file
args:
-- filename: the filepath as a string
-- start: the start of the chunk to read
-- end: the end of a chunk to read
**/

func countDivided(filename string, start int64, end int64, c chan *map[string]int, waitgroup *sync.WaitGroup) {
	defer waitgroup.Done()
	counts := make(map[string] int) //store word counts by key which is the word itself
	var nonLetter = regexp.MustCompile(`[^a-zA-Z0-9]+`)
		
	file, err := os.Open(filename) //file pointer
	if err != nil {
		log.Fatal(err) //this wasnt fatal, then I got a segfault, so here
	}
	defer file.Close() //make sure this is closed before return
	reader := OffsetReader{io.ReaderAt(file), start} //start at desired start point.
	scanner := bufio.NewScanner(reader) //buffered i/o: creates a pipe for reading
	scanner.Split(bufio.ScanWords) //break reading pattern into words
	bytesRead := 0
	for scanner.Scan() && reader.offset < end { //reads until EOF OR until the limit
		word := scanner.Text()
		bytesRead += len(word) + 1 //read this many bytes
		word = strings.ToLower(word) //lowercase-ify
		word = nonLetter.ReplaceAllString(word, " ") //get rid of extra characters
		words := strings.Split(word, " ") //split words by char
		for _, wd := range(words) {
			wd2 := nonLetter.ReplaceAllString(wd, "") //get rid of spaces
			counts[wd2] = counts[wd2] + 1 //increment word count in the dictionary
		}
	}
	c <- &counts
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
An attempt to divvy up the work based on the number of available CPUs.
Oddly enough, it does not perform very well.
**/
func multi2(files []string) {
	numRoutines := runtime.NumCPU() //base the number of threads on the number of CPUs you have.
	//fmt.Println("number of threads:", numRoutines)
	l := len(files)
	perEach := l / numRoutines
	leftOver := l % numRoutines
	/**
	essentially, we predetermine the number of threads, and assign some files per thread
	**/
	if perEach < 1 { //more routines than files, just assign one per each.
		multiThreaded(files)
		return
	}
	var waitgroup sync.WaitGroup //waiting for completion of routines
	c := make(chan *map[string]int, l) //store results here
	current := 0 //divide up the files
	for i := 0; i < numRoutines; i ++ { //send out a routine which acts on each set of files.
		//fmt.Println("current is", current)
		waitgroup.Add(1)
		if leftOver > 0 {
			go manage(c, files[current:current+perEach+1], &waitgroup)
		} else {
			go manage(c, files[current:current+perEach], &waitgroup)
		}
		current += perEach
		leftOver -= 1
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

//used for splitting large files
type fpOffsets struct {
	filenames []string
	offsets [][]int64
}

func newFpOffsets (filenames []string, offsets [][]int64) *fpOffsets {
	p := fpOffsets{filenames: filenames, offsets: offsets}
	return &p
}
/**
determine the size of each file and break large ones into two.
**/
func divideFiles(fps []string, threshold int64) *fpOffsets {
	fileOffsets := make([][]int64, 0, len(fps) * 2) 
	fileNames := make([]string, 0, len(fps) * 2)
	for _, f := range(fps) {
		descript, err := os.Stat(f)
		if os.IsNotExist(err) {
			log.Fatal("error: file not found")
		}
		fileNames = append(fileNames, descript.Name())
		s := descript.Size()
		if s > threshold {
			fileNames = append(fileNames, descript.Name()) //append a second one.
			fileOffsets = append(fileOffsets, ([]int64 {0, s / 2}))
			fileOffsets = append(fileOffsets,([]int64 {s/2, s}) )
		} else {
			fileOffsets = append(fileOffsets, ([]int64 {0, s})) //whole file.
		}
	}
	fpOffs := newFpOffsets(fileNames, fileOffsets)
	return fpOffs
}


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
	err := writeMapToFile(os.Args[1] + "/multi.txt", &totals)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}	
}

func multiDivided(files []string) {

	filesOffsets := divideFiles(files, 5000000) //list of filenames (names are repeated) and the offsets
	names := filesOffsets.filenames
	offsets := filesOffsets.offsets

	c := make(chan *map[string]int, len(names)) //store results here
	var waitgroup sync.WaitGroup //waiting for completion of routines

	for i, name := range(names) {
		start := offsets[i][0]
		end := offsets[i][1]
		waitgroup.Add(1)
		go countDivided(name, start, end, c, &waitgroup)	
	}

	waitgroup.Wait()
	close(c)

	result := make(map[string]int)
	for countMap := range(c) { //for each routine's individual map...
		for word, count := range(*countMap) { //dereferenced
			result[word] = result[word] + count
		}
	}

	err := writeMapToFile(os.Args[1] + "/multiDivided.txt", &result)
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
		start := time.Now()
		multiThreaded(files)
		fmt.Println(time.Since(start))
		start2 := time.Now()
		multi2(files)
		fmt.Println(time.Since(start2))
		start3 := time.Now()
		multiDivided(files)
		fmt.Println(time.Since(start3))
	}
	
	// TODO: add argument processing and run both single-threaded and multi-threaded functions
}
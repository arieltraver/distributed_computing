package main
/*references used
https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
*/


import(
	"fmt"
	"os"
	"log"
	"bufio"
)

func readFiles(directory string)([]string, int) {
	filepaths, err := os.ReadDir(directory)
	if err != nil {
		log.Println("Error: could not read files.")
		return nil, 1
	}
	for _, file := range(filepaths) {
		if file.IsDir() {
			log.Println("One of the files is a directory.\nPlease place all text files in a flat directory.")
			return nil, 2
		}
	}
	files := make([]string, len(filepaths))
	for i, fname := range(filepaths) {
		files[i] = fname.Name()
	}
	return files, 0
}

/*type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}*/

func single_threaded(files []string) {
	counts := make(map[string] int) //store word counts by key which is the word itself
	for _, filename := range(files) {
		file, err := os.Open(filename) //file pointer
		if err != nil {
			log.Println(err)
			fmt.Println("error opening file:", filename)
		}
		scanner := bufio.NewScanner(file) //buffered i/o: creates a pipe for reading
		scanner.Split(bufio.ScanWords) //break reading pattern into words
		for scanner.Scan() { //reads until EOF
			word := scanner.Text()
			counts[word] = counts[word] + 1 //increment word count in the dictionary
			fmt.Println(word)
		}
	}
	for key, count := range(counts) {
		fmt.Println(key, ": ", count)
	}
	
}

func multi_threaded(files []string) {
	// TODO: Your multi-threaded implementation
}


func main() {
	files, error := readFiles("./input")
	switch error {
	case 1:
		fmt.Println("Error reading file names from directory")
		return
	case 2:
		fmt.Println("One of the files in that directory is also a directory")
		fmt.Println("Please store only text files in your directory")
	default:
		single_threaded(files)
	}
	
	// TODO: add argument processing and run both single-threaded and multi-threaded functions
}
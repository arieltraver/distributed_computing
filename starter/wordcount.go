package main
/*references used
https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
*/


import(
	"fmt"
	"os"
	"log"
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



func single_threaded(files []string) {

	
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
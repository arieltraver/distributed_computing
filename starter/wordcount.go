package main
/*references used
https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
*/


import(
	"fmt"
	"io/iutil"
	"os"
	"log"
)

func readFiles(directory string)([]os.FileInfo, int) {
	filepaths, err := iutil.ReadDir(directory)
	if err {
		log.Println("Error: could not read files.")
		return nil, 1
	}
	for _, file := range(filepaths) {
		if file.IsDir() {
			log.Println("One of the files is a directory.\nPlease place all text files in a flat directory.")
			return nil, 2
		}
	}
	return filepaths, 0
}

func single_threaded(files []string) {
	// TODO: Your single-threaded implementation
}

func multi_threaded(files []string) {
	// TODO: Your multi-threaded implementation
}


func main() {
	// TODO: add argument processing and run both single-threaded and multi-threaded functions
}
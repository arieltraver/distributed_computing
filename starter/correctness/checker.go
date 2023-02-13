package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
)

func main() {
	single_map := file_to_map("output/single.txt")
	multi_map := file_to_map("output/multi.txt")

	for sk, sv := range single_map {
		count, ok := multi_map[sk]
		if !ok {
			fmt.Println("FAIL: single.txt contains " + sk + ", but multi.txt doesn't")
			os.Exit(1)
		}
		if count != sv {
			fmt.Println("FAIL: single.txt has count " + fmt.Sprint(sk) + " for " + sk + ", but multi.txt has " + fmt.Sprint(count))
			os.Exit(1)
		}
	}

	fmt.Println("OK")
}

func file_to_map(file_name string) map[string]int{
	result_map := make(map[string]int)
	dat, err := os.ReadFile(file_name)
	if err != nil {
		fmt.Println("FAIL: " + file_name + " not found")
		os.Exit(1)
	}

	arr := strings.Fields(string(dat))
	
	for i := 0; i < len(arr); i += 2 {
		counter, err := strconv.Atoi(arr[i + 1])
		if err != nil {
			fmt.Println("FAIL: Formatting error - " + arr[i + 1] + " is not an integer")
			os.Exit(1)
		}
		result_map[arr[i]] = counter
	}
	return result_map
}
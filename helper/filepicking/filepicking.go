package filepicking

import (
	"fmt"
	"bufio"
	"os"
)

func PickFile() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter file path: ")
	if scanner.Scan() {
		filePath := scanner.Text()
		fmt.Println("You entered:", filePath)
		fmt.Println("Sending filepath to handler...")
	} else {
		fmt.Println("Error reading file path")
	}
}
package root

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	flagW bool
	flagC bool
	flagL bool
)

var wordCountCmd = &cobra.Command{
	Use:   "word-count",
	Short: "Count the number of words in a string",
	Run: func(cmd *cobra.Command, args []string) {
		//check if file path is valid or is it directory
		if len(args) == 0 {
			fmt.Println("Please provide file names as arguments")
			return
		}
		maxFilesLimitBuffer := make(chan int, 1)
		var wg sync.WaitGroup
		for _, file := range args {
			go processFile(file, &wg, maxFilesLimitBuffer)
			wg.Add(1)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagW, "w", "w", false, "Enable flag -w")
	rootCmd.PersistentFlags().BoolVarP(&flagC, "c", "c", false, "Enable flag -c")
	rootCmd.PersistentFlags().BoolVarP(&flagL, "l", "l", false, "Enable flag -l")
	rootCmd.AddCommand(wordCountCmd)
}

type result struct {
	charCount int
	wordCount int
	lineCount int
	err       error
	file      string
}

func processFile(file string, wg *sync.WaitGroup, maxFilesLimitBuffer chan int) {
	maxFilesLimitBuffer <- 1
	lines := make(chan string)
	errChan := make(chan error)

	defer func() {
		<-maxFilesLimitBuffer
		wg.Done()
	}()

	go readLinesInFile(file, lines, errChan)

	result := count(lines, errChan)
	if result.err != nil {
		fmt.Println(result.err)
		return
	}
	if flagW {
		fmt.Printf("%d ", result.wordCount)
	}
	if flagC {
		fmt.Printf("%d ", result.charCount)
	}
	if flagL {
		fmt.Printf("%d ", result.lineCount)
	}
	fmt.Printf("%s\n", file)

}

func readLinesInFile(file string, lines chan<- string, errChan chan<- error) {
	var scanner *bufio.Scanner
	const chunkSize = 1024 * 1024
	defer close(lines)
	defer close(errChan)

	file_, err := os.Open(file)
	if err != nil {
		errChan <- err
		return
	}
	defer file_.Close()
	scanner = bufio.NewScanner(file_)
	scanner.Buffer(make([]byte, chunkSize), chunkSize)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		errChan <- err
		return
	}
}

func count(lines <-chan string, errChan <-chan error) result {
	var r result
	for {
		select {
		case err := <-errChan:
			r.err = err
			return r
		case line, ok := <-lines:
			if !ok {
				return r
			}
			r.charCount += len(line) + 1
			r.wordCount += len(strings.Fields(line))
			r.lineCount++
		}
	}
}

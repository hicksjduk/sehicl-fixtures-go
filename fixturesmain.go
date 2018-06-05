package main

import (
	"bytes"
	"fixtures/fixtures"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"log"
	"os/exec"
)

const breakpointFile = "breakpoint"
const bestFile = "best"

var bestScore = -1
const messageFrequency = 100000
const commitFrequency = 1000000

func main() {
	bestScore = readBestScore()
	resultChan := make(chan EvaluationResult, 10)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGKILL)
	stoppingChan := make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go waitForSignal(sigChan, stoppingChan)
	go processResults(resultChan, wg)
	go processCombinations(resultChan, stoppingChan, wg)
	wg.Wait()
}

func waitForSignal(sigChan chan os.Signal, stoppingChan chan struct{}) {
	for sig := range sigChan {
		log.Printf("Signal %v received, stopping", sig)
		stoppingChan <- struct{}{}
		break
	}
}

func processCombinations(resultChan chan EvaluationResult, stoppingChan chan struct{}, wg sync.WaitGroup) {
	defer wg.Done()
	defer close(resultChan)
	list := fixtures.BuildFixtureList()
	it := list.Iterator(readBreakpoints()...)
	for {
		if checkForStop(stoppingChan) {
			break
		}
		sch, ok := it.Next()
		if !ok {
			log.Print("All combinations processed!")
			break
		}
		result := EvaluationResult{
			indices:  it.NextIndices(),
			schedule: sch,
			score:    sch.Evaluate(),
		}
		resultChan <- result
	}
}

func checkForStop(stoppingChan chan struct{}) bool {
	select {
	case <-stoppingChan:
		return true
	default:
		return false
	}
}

func processResults(resultChan chan EvaluationResult, wg sync.WaitGroup) {
	defer wg.Done()
	committer := intervalProcessor(commitFrequency, func(indices []int) {
		log.Printf("Committing after %d combinations", commitFrequency)
		writeBreakpoints(indices)
		if cmdout, err := exec.Command("git", "commit", "-m", "Latest status", bestFile, breakpointFile).Output(); err != nil {
			log.Printf("Commit failed", cmdout, err)
		}
	})
	logger := intervalProcessor(messageFrequency, func(indices []int) {
		log.Printf("Processed another batch of %d combinations: latest one was %v", messageFrequency, indices)
	})
	for result := range resultChan {
		if bestScore == -1 || bestScore > result.score {
			writeBest(result.schedule, result.score)
			log.Printf("Found a better score: %d (was %d)", result.score, bestScore)
			bestScore = result.score
		}
		committer(result.indices)
		logger(result.indices)
	}
}

func intervalProcessor(interval int, f func([]int)) func([]int) {
	count := 0
	return func(indices []int) {
		count++
		if count >= interval {
			count = 0
			f(indices)
		}
	}
}

func readBreakpoints() []int {
	data, read := readFile(breakpointFile)
	if !read {
		log.Printf("File %s not found", breakpointFile)
		return []int{}
	}
	breakpoints := parseBreakpoints(data)
	if breakpoints == nil {
		log.Fatalf("File %s found but is not valid in format", breakpointFile)
	}
	log.Printf("Found breakpoints in file %s: %v", breakpointFile, breakpoints)
	return breakpoints
}

func parseBreakpoints(data []byte) []int {
	if r, _ := regexp.Compile("^\\s*(\\d+\\s*)+$"); !r.Match(data) {
		return nil
	}
	r, _ := regexp.Compile("\\d+")
	numbers := r.FindAll(data, -1)
	islice := make([]int, len(numbers))
	for i, v := range numbers {
		islice[i], _ = strconv.Atoi(string(v))
	}
	return islice
}

func writeBreakpoints(data []int) {
	var buffer bytes.Buffer
	for _, v := range data {
		buffer.WriteString(fmt.Sprintf("%d ", v))
	}
	ioutil.WriteFile(breakpointFile, buffer.Bytes(), 0644)
}

func readFile(name string) ([]byte, bool) {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		log.Fatalf("File %s found but could not be read: %v", name, err)
	}
	defer f.Close()
	answer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("File %s found but could not be read: %v", name, err)
	}
	return answer, true
}

func readBestScore() int {
	data, read := readFile(bestFile)
	if !read {
		log.Printf("File %s not found", bestFile)
		return -1
	}
	best, err := parseBestScore(data)
	if err != nil {
		log.Fatalf("File %s found but is not valid in format: %v", bestFile, err)
		os.Exit(1)
	}
	log.Printf("Found best score in file %s: %d", bestFile, best)
	return best
}

func parseBestScore(data []byte) (int, error) {
	r, _ := regexp.Compile("\\d+")
	best, err := strconv.Atoi(string(r.Find(data)))
	if err != nil {
		return 0, err
	}
	return best, nil
}

func writeBest(schedule fixtures.Schedule, score int) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%d\n%v", score, schedule.String()))
	ioutil.WriteFile(bestFile, buffer.Bytes(), 644)
}

type EvaluationResult struct {
	indices  []int
	schedule fixtures.Schedule
	score    int
}

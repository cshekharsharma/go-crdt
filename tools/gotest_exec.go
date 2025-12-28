//go:build exclude_from_tests
// +build exclude_from_tests

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TestResult represents the JSON output from `go test -v -json`
type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"` // Field to capture elapsed time
}

const lineWidth = 105

var (
	packagesToSkipInTests = []string{
		"mocks",
		"testdata",
		"testutil",
	}
)

func main() {
	skipMocks := flag.Bool("skip-mocks", false, "Skip packages with names containing 'mocks'")
	flag.Parse()

	startTime := time.Now().UnixMilli()
	cmd := getTestCaseExecutionCommand(skipMocks)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	// Parse each line of the output
	dec := json.NewDecoder(&out)

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedDirectories := ""
	skippedCount := 0
	breakingDirectories := ""
	breakingDirectoriesCount := 0

	for dec.More() {
		var result TestResult
		if err := dec.Decode(&result); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		status := ""

		if result.Action == "run" {
			totalTests++
		}

		// Process and print results with elapsed time
		if result.Action == "pass" && result.Test != "" {
			passedTests++
			status = "\033[1;32mPASS\033[0m"
		} else if result.Action == "fail" && result.Test != "" {
			failedTests++
			status = "\033[1;31mFAIL\033[0m"
		} else if result.Action == "fail" {
			breakingDirectories += fmt.Sprintf(">> \033[0m %s/%s\n", result.Package, result.Test)
			breakingDirectoriesCount++
			continue
		} else if result.Action == "skip" {
			skippedDirectories += fmt.Sprintf(">> \033[0m %s/%s\n", result.Package, result.Test)
			skippedCount++
			continue
		} else {
			continue
		}

		fmt.Printf(">> %s: \033[36m[%.2fs]\033[0m %s/%s\n", status, result.Elapsed, result.Package, result.Test)
	}

	passedPercent := fmt.Sprintf("%.2f", float64(passedTests)/float64(totalTests)*100)
	failedPercent := fmt.Sprintf("%.2f", float64(failedTests)/float64(totalTests)*100)

	fmt.Printf("%s\n\n", strings.Repeat("=", lineWidth))
	fmt.Printf("\033[1;32mPASSED:  \033[0m %d/%d \t[ %v%% ]\n", passedTests, totalTests, passedPercent)
	fmt.Printf("\033[1;31mFAILED:  \033[0m %d/%d \t[ %v%% ]\n\n\n", failedTests, totalTests, failedPercent)

	if breakingDirectoriesCount > 0 {
		fmt.Printf("\033[1;31mFew of the test cases are breaking. Please check the following directories:\n\n")
		fmt.Printf("\033[0m%v \n\n", breakingDirectories)
	}

	fmt.Printf("\033[1;33mSKIPPED Directories: \033[0m %d\n\n", skippedCount)
	fmt.Printf("%v \n\n", skippedDirectories)

	fmt.Printf("\033[1;36mDURATION: \033[0m \033[1;32m\u2605\u2605\u2605\033[0m %.3f seconds\n", float64(time.Now().UnixMilli()-startTime)/1000)
	fmt.Printf("%s\n\n", strings.Repeat("=", lineWidth))

	if failedTests > 0 || breakingDirectoriesCount > 0 {
		os.Exit(1)
	}
}

// getTestCaseExecutionCommand returns the command to execute test cases.
func getTestCaseExecutionCommand(skipMocks *bool) *exec.Cmd {
	cmd := exec.Command("go", "test", "./...", "-v", "-json", "-coverprofile=./coverage.txt")

	if *skipMocks {
		pkgListRaw, _ := exec.Command("go", "list", "./...").Output()
		allPkgs := strings.Split(strings.TrimSpace(string(pkgListRaw)), "\n")

		filteredPkgs := []string{}
		for _, pkg := range allPkgs {
			if inArrayMatch(pkg, packagesToSkipInTests) {
				continue
			}

			filteredPkgs = append(filteredPkgs, pkg)
		}

		if len(filteredPkgs) == 0 {
			fmt.Println("No packages to test after filtering.")
			os.Exit(0)
		}

		args := append([]string{"test", "-v", "-json", "-coverprofile=./coverage.txt"}, filteredPkgs...)
		cmd = exec.Command("go", args...)
	}

	return cmd
}

func inArrayMatch(input string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(input, keyword) {
			return true
		}
	}
	return false
}

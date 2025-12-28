//go:build exclude_from_tests
// +build exclude_from_tests

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"

	coverageFileName = "coverage.txt"
	coverageFilePath = coverageFileName

	lineWidth = 107
)

type blockKey struct {
	file  string
	start string
	end   string
}

type packageStat struct {
	covered   float64
	total     float64
	packageID string
}

func colorize(percent float64) string {
	switch {
	case percent >= 80:
		return Green
	case percent >= 50:
		return Yellow
	default:
		return Red
	}
}

func coverageBar(percent float64) string {
	blocks := int(percent / 4)
	return strings.Repeat("█", blocks) + strings.Repeat("░", 25-blocks)
}

func extractPackage(path string) string {
	dirs := strings.Split(filepath.ToSlash(path), "/")
	if len(dirs) <= 1 {
		return "."
	}
	return strings.Join(dirs[:len(dirs)-1], "/")
}

func main() {
	file, err := os.Open(coverageFilePath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", coverageFileName, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	packageData := make(map[string]*packageStat)
	blockSeen := make(map[blockKey]bool)

	var totalCovered, totalStatements float64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		filename := parts[0]
		rest := parts[1]

		fields := strings.Fields(rest)
		if len(fields) != 3 {
			continue
		}

		block := blockKey{
			file:  filename,
			start: fields[0],
			end:   fields[1],
		}
		if blockSeen[block] {
			continue
		}
		blockSeen[block] = true

		statementsF, err1 := strconv.ParseFloat(fields[1], 64)
		countF, err2 := strconv.ParseFloat(fields[2], 64)
		if err1 != nil || err2 != nil {
			continue
		}

		covered := 0.0
		if countF > 0 {
			covered = statementsF
		}

		pkg := extractPackage(filename)
		if _, exists := packageData[pkg]; !exists {
			packageData[pkg] = &packageStat{packageID: pkg}
		}
		packageData[pkg].total += statementsF
		packageData[pkg].covered += covered

		totalStatements += statementsF
		totalCovered += covered
	}

	fmt.Printf("\n%s📦 PACKAGE COVERAGE REPORT (FROM COVERAGE FILE)%s\n\n", Bold, Reset)
	fmt.Printf("%s%-70s %-10s %s%s\n", Bold, "Package", "Coverage", "Bar", Reset)
	fmt.Println(strings.Repeat("─", lineWidth))

	packages := make([]string, 0, len(packageData))
	for pkg := range packageData {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	for _, pkg := range packages {
		stat := packageData[pkg]
		pct := 100.0 * stat.covered / stat.total
		bar := coverageBar(pct)
		color := colorize(pct)
		fmt.Printf("%-70s %s%6.1f%%%s   %s\n", pkg, color, pct, Reset, bar)
	}

	overallPct := 100.0 * totalCovered / totalStatements
	color := colorize(overallPct)
	bar := coverageBar(overallPct)

	fmt.Println(strings.Repeat("─", lineWidth))
	fmt.Printf("%s%-70s%s %s%6.2f%%%s   %s\n\n", Yellow, "TOTAL COVERAGE", Reset, color, overallPct, Reset, bar)
}

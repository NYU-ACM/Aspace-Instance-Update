package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Row struct {
	Resource               string
	RefID                  string
	URI                    string
	ContainerIndicator1    string
	ContainerIndicator2    string
	ContainerIndicator3    string
	Title                  string
	ComponentId            string
	NewContainerIndicator1 string
	NewContainerIndicator2 string
}

func main() {
	tsv, _ := os.Open("test.tsv")
	rows := GetTSVRows(tsv)
	for _, row := range rows {
		fmt.Println(row)
	}
}

func GetTSVRows(tsv *os.File) []Row {
	rows := []Row{}
	scanner := bufio.NewScanner(tsv)
	scanner.Scan() // skip the header line
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		rows = append(rows, Row{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6], cols[7], cols[8], cols[9]})
	}
	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	return rows
}

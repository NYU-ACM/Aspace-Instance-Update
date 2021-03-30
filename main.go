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
	//open a work order and check for errors
	tsv, err := os.Open("test.tsv")
	if err != nil {
		panic(err)
	}
	//get the rows of the tsv file as an array
	rows, err := GetTSVRows(tsv)
	if err != nil {
		panic(err)
	}

	//print each row in the Array
	for _, row := range rows {
		fmt.Println(row)
	}
}

func GetTSVRows(tsv *os.File) ([]Row, error) {
	//create an empty array or Rows
	rows := []Row{}
	//create a scanner object and read the tsv file
	scanner := bufio.NewScanner(tsv)
	// skip the header line
	scanner.Scan()
	// scan the tsv file line by line
	for scanner.Scan() {
		//split the line by tab chars
		cols := strings.Split(scanner.Text(), "\t")
		//marshal the split line into a Row struct and add to the array of Rows
		rows = append(rows, Row{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6], cols[7], cols[8], cols[9]})
	}
	//Check for any read errors
	if scanner.Err() != nil {
		return rows, scanner.Err()
	}

	//return the array of Rows
	return rows, nil
}

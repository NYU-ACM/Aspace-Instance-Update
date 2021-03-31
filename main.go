package main

import (
	"bufio"
	"fmt"
	"github.com/nyudlts/go-aspace"
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
	fmt.Println("aspace-instance-update")
	//open a work order and check for errors
	fmt.Println("1. Parsing work order")
	tsv, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	//get the rows of the tsv file as an array
	rows, err := GetTSVRows(tsv)
	if err != nil {
		panic(err)
	}

	fmt.Println("2. Getting Aspace client")
	//get a go-aspace client
	client, err := aspace.NewClient("dev", 20)
	if err != nil {
		panic(err)
	}

	fmt.Println("3. Getting repository and resource IDs")
	//get the repository ID from the first row of the TSV
	repositoryId,aoID, err := aspace.URISplit(rows[1].URI)
	if err != nil {
		panic(err)
	}

	//Get the resource ID from the first row of the TSV -- this is a hack
	ao, err := client.GetArchivalObject(repositoryId, aoID)
	if err != nil {
		panic(err)
	}
	_, resourceId, err := aspace.URISplit(ao.Resource["ref"])

	//Get a map of Top Containers from aspace  for the resource
	fmt.Println("4. Getting Top Containers for resource")
	topContainers, err := client.GetTopContainersForResource(repositoryId, resourceId)
	if err != nil {
		panic(err)
	}

	fmt.Println("5. Updating AO indicators and Top Container URI")
	//iterate each row in the Array
	for _, row := range rows {
		tc := topContainers[row.ContainerIndicator1]
		fmt.Println(tc.Indicator, "->", tc.Barcode)  //the function to test,update,and undo topcontainer info goes here.
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

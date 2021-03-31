package main

import (
	"bufio"
	"flag"
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

var client *aspace.ASClient
var topContainers map[string]aspace.TopContainer
var wo string
var test bool
var undo bool

func init() {
	flag.StringVar(&wo, "workorder", "", "work order location")
	flag.BoolVar(&test, "test", false, "run in test mode")
	flag.BoolVar(&undo, "undo", false, "run in undo mode")
	flag.Parse()
}

func main() {
	fmt.Println("aspace-instance-update")

	//check if the work order exists or is null
	if wo == "" {
		panic(fmt.Errorf("No work order specified, exiting"))
	}

	if _, err := os.Stat(wo); os.IsNotExist(err) {
		panic(fmt.Errorf("Work order location is not valid, exiting"))
	}

	//open a work order and check for errors
	fmt.Println("1. Parsing work order")
	tsv, err := os.Open(wo)
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
	client, err = aspace.NewClient("dev", 20)
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
	topContainers, err = client.GetTopContainersForResource(repositoryId, resourceId)
	if err != nil {
		panic(err)
	}

	fmt.Println("5. Updating AO indicators and Top Container URI")
	//iterate each row in the Array
	for _, row := range rows {
		if (row.ContainerIndicator1 != row.NewContainerIndicator1 || row.ContainerIndicator2 != row.NewContainerIndicator2) {
			err = UpdateAO(row)
		}
	}
}
func UpdateAO(row Row) error {
	fmt.Println(row.URI)

	repoId, aoID, err := aspace.URISplit(row.URI)
	if err != nil {
		return err
	}

	ao, err := client.GetArchivalObject(repoId, aoID)
	if err != nil {
		return err
	}

	//update top Container Reference
	if row.ContainerIndicator1 != row.NewContainerIndicator1 {
		newTopContainer := topContainers[row.NewContainerIndicator1]
		ao.Instances[0].SubContainer.TopContainer["ref"] = newTopContainer.URI
	}

	//update indicator 2
	if row.ContainerIndicator2 != row.NewContainerIndicator2 {
		ao.Instances[0].SubContainer.Indicator_2 = row.NewContainerIndicator2
	}

	return nil
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

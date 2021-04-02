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
	Resource               	string
	RefID                  	string
	URI                    	string
	ContainerIndicator1    	string
	ContainerIndicator2    	string
	ContainerIndicator3    	string
	Title                  	string
	ComponentId            	string
	Barcode				   	string
	NewContainerIndicator2	string
	NewBarcode 				string
}

var (
	client *aspace.ASClient
	topContainers []aspace.TopContainer
	topContainerMap map[string]aspace.TopContainer
	wo string
	test bool
	undo bool
	env string
	helpmsg bool
)

func init() {
	flag.StringVar(&env, "environment", "", "environment to run script")
	flag.StringVar(&wo, "workorder", "", "work order location")
	flag.BoolVar(&test, "test", false, "run in test mode")
	flag.BoolVar(&undo, "undo", false, "run in undo mode")
	flag.BoolVar(&helpmsg, "help", false, "display the help message")
	flag.Parse()
}

func main() {
	//check if the help flag is set
	if helpmsg == true {
		help()
	}

	fmt.Println("aspace-instance-update")

	//check if the work order exists or is null
	if wo == "" {
		fmt.Printf("No work order specified, exiting")
		help()

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
	if env == "" {
		panic(fmt.Errorf("Environment must be defined"))
	}
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

	//Get a list of Top Containers from aspace  for the resource
	fmt.Println("4. Getting Top Containers for resource")

	topContainers, err = client.GetTopContainersForResource(repositoryId, resourceId)
	if err != nil {
		panic(err)
	}
	//create a map of top containers indexed by barcode
	topContainerMap = MapTopContainers(topContainers)

	fmt.Println("5. Updating AO indicators and Top Container URI")
	//iterate each row in the Array
	for _, row := range rows {
		if (row.Barcode != row.NewBarcode || row.ContainerIndicator2 != row.NewContainerIndicator2) {
			msg, err := UpdateAO(row)
			if err != nil {
				panic(err)
			}
			fmt.Println("    Result:", msg)
		}
	}
}

func MapTopContainers(tcs []aspace.TopContainer) map[string]aspace.TopContainer {
	tcMap := map[string]aspace.TopContainer{}
	for _, tc := range tcs {
		if(tc.Barcode != "") {
			tcMap[tc.Barcode] = tc
		}
	}
	return tcMap
}

func UpdateAO(row Row) (string, error) {
	fmt.Println("  Updating:", row.URI)

	repoId, aoID, err := aspace.URISplit(row.URI)
	if err != nil {
		return "", err
	}

	ao, err := client.GetArchivalObject(repoId, aoID)
	if err != nil {
		return "", err
	}

	fmt.Println("    Before: ", ao.Instances)
	//update top Container Reference
	if row.Barcode != row.NewBarcode {
		var newTopContainer aspace.TopContainer
		if undo != true {
			newTopContainer = topContainerMap[row.NewBarcode]
		} else {
			newTopContainer = topContainerMap[row.Barcode]
		}
		ao.Instances[0].SubContainer.TopContainer["ref"] = newTopContainer.URI
	}

	//update indicator 2
	if row.ContainerIndicator2 != row.NewContainerIndicator2 {
		if undo != true {
			for i, instance := range ao.Instances {
				if instance.SubContainer.Indicator_2 == row.ContainerIndicator2 {
					ao.Instances[i].SubContainer.Indicator_2 = row.NewContainerIndicator2
				}
			}

		} else {
			for i, instance := range ao.Instances {
				if instance.SubContainer.Indicator_2 == row.NewContainerIndicator2 {
					ao.Instances[i].SubContainer.Indicator_2 = row.ContainerIndicator2
				}
			}
		}
	}

	fmt.Println("    After: ", ao.Instances)

	if test == true {
		return "Test Mode - not Updating AO", nil
	} else {
		//update the ao
		msg, err := client.UpdateArchivalObject(repoId, aoID, ao)
		if err != nil {
			 return "", nil
		}

		return msg, nil
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
		rows = append(rows, Row{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6], cols[7], cols[8], cols[9], cols[10]})
	}
	//Check for any read errors
	if scanner.Err() != nil {
		return rows, scanner.Err()
	}

	//return the array of Rows
	return rows, nil
}

func help() {
	fmt.Println(`$ aspace-instance-update options
options:
  --workorder, required, /path/to/workorder.tsv
  --environment, required, aspace environment to be used: dev/stage/prod
  --undo, optional, runs a work order in revrse, undo a previous run
  --test, optional, test mode does not execute any POSTs, this is recommended before running on any data
  --help print this help message`)
	os.Exit(0)
}
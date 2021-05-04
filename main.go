package main

import (
	"bufio"
	"encoding/json"
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
	Barcode                string
	NewContainerIndicator2 string
	NewBarcode             string
}

var (
	versionNum      = "v1.0.4b"
	client          *aspace.ASClient
	topContainers   []aspace.TopContainer
	topContainerMap map[string]aspace.TopContainer
	wo              string
	test            bool
	undo            bool
	env             string
	helpmsg         bool
	version         bool
	writer          *bufio.Writer
)

func init() {
	flag.BoolVar(&version, "version", false, "display the version")
	flag.StringVar(&env, "environment", "", "environment to run script")
	flag.StringVar(&wo, "workorder", "", "work order location")
	flag.BoolVar(&test, "test", false, "run in test mode")
	flag.BoolVar(&undo, "undo", false, "run in undo mode")
	flag.BoolVar(&helpmsg, "help", false, "display the help message")
	flag.Parse()
}

func help() {
	fmt.Println(`usage:
  $ aspace-instance-update [options]
options:
  --workorder, required, /path/to/workorder.tsv
  --environment, required, aspace environment to be used: dev/stage/prod
  --undo, optional, runs a work order in revrse, undo a previous run
  --test, optional, test mode does not execute any POSTs, this is recommended before running on any data
  --help print this help message
  --version print the version info`)

}

func main() {

	//check if the help flag is set
	if helpmsg == true {
		help()
		os.Exit(0)
	}

	if version == true {
		fmt.Println("aspace-instance-update", versionNum)
		os.Exit(0)
	}

	fmt.Println("aspace-instance-update", versionNum)

	//check if the work order exists or is null
	if wo == "" {
		fmt.Printf("No work order file specified, exiting")
		help()
	}

	if _, err := os.Stat(wo); os.IsNotExist(err) {
		panic(fmt.Errorf("Work order location is not valid, exiting"))
	}

	//open a work order and check for errors
	fmt.Println("1. Opening work work order")
	tsv, err := os.Open(wo)
	if err != nil {
		panic(err)
	}

	//create a logger
	logName := "AIU-" + tsv.Name()
	fmt.Println("2. Creating Logfile", logName)
	logFile, err := os.Create(logName)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	writer = bufio.NewWriter(logFile)
	writer.WriteString("AO URI\tResult\tOriginal Barcode\tUpdated Barcode\tOriginal Child Ind 2\tUpdated Child Ind 2\tError Msg\n")
	writer.Flush()

	fmt.Println("3. Parsing Work Order")
	//get the rows of the tsv file as an array
	rows, err := GetTSVRows(tsv)
	if err != nil {
		panic(err)
	}

	fmt.Println("4. Getting Aspace client")
	//get a go-aspace client
	if env == "" {
		panic(fmt.Errorf("Environment must be defined"))
	}
	client, err = aspace.NewClient(env, 20)
	if err != nil {
		panic(err)
	}

	fmt.Println("5. Getting repository and resource IDs")
	//get the repository ID from the first row of the TSV
	repositoryId, aoID, err := aspace.URISplit(rows[1].URI)
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
	fmt.Println("6. Getting Top Containers for resource")

	topContainers, err = client.GetTopContainersForResource(repositoryId, resourceId)
	if err != nil {
		panic(err)
	}
	//create a map of top containers indexed by barcode
	topContainerMap = MapTopContainers(topContainers)




	fmt.Println("7. Updating AO indicators and Top Container URI")
	//iterate each row in the Array

	for _, row := range rows {
		if row.Barcode != row.NewBarcode || row.ContainerIndicator2 != row.NewContainerIndicator2 {
			msg, err := UpdateAO(row)
			if err != nil {
				panic(err)
			}
			fmt.Println("    Result:", msg, "\n")
		}
	}

	fmt.Println("Update Complete, exiting.")

}

func MapTopContainers(tcs []aspace.TopContainer) map[string]aspace.TopContainer {
	tcMap := map[string]aspace.TopContainer{}
	for _, tc := range tcs {
		if tc.Barcode != "" {
			tcMap[tc.Barcode] = tc
		}
	}
	return tcMap
}

func UpdateAO(row Row) (string, error) {
	fmt.Println("Updating:", row.URI)

	repoId, aoID, err := aspace.URISplit(row.URI)
	if err != nil {
		WriteToLog(row.URI, "ERROR", "", "", "", "", err.Error())
		return "Could not parse URI " + row.URI + " , Skipping", nil
	}

	ao, err := client.GetArchivalObject(repoId, aoID)
	if err != nil {
		WriteToLog(row.URI, "ERROR", "", "", "", "", err.Error())
		return "ArchivesSpace return a 404 for " + row.URI, nil
	}

	var beforeBarcode string
	var afterBarcode string
	var beforeCI2 string
	var afterCI2 string

	for i, instance := range ao.Instances {

		if undo != true {
			//update barcode
			if instance.SubContainer.TopContainer["ref"] == topContainerMap[row.Barcode].URI {
				ao.Instances[i].SubContainer.TopContainer["ref"] = topContainerMap[row.NewBarcode].URI
			}
			beforeBarcode = row.Barcode
			afterBarcode = row.NewBarcode

			//update indicator 2
			if instance.SubContainer.Indicator_2 == row.ContainerIndicator2 {
				ao.Instances[i].SubContainer.Indicator_2 = row.NewContainerIndicator2
			}
			beforeCI2 = row.ContainerIndicator2
			afterCI2 = row.NewContainerIndicator2

		} else {
			//update barcode undo
			if instance.SubContainer.TopContainer["ref"] == topContainerMap[row.NewBarcode].URI {
				ao.Instances[i].SubContainer.TopContainer["ref"] = topContainerMap[row.Barcode].URI
			}
			beforeBarcode = row.NewBarcode
			afterBarcode = row.Barcode

			//update indicator 2 undo
			if instance.SubContainer.Indicator_2 == row.NewContainerIndicator2 {
				ao.Instances[i].SubContainer.Indicator_2 = row.ContainerIndicator2
			}
			beforeCI2 = row.NewContainerIndicator2
			afterCI2 = row.ContainerIndicator2
		}
	}

	//Output to console
	fmt.Printf("    Original Barcode: %s, Updated Barcode: %s, Original Child Ind 2: %s, Updated Child Ind 2: %s\n", beforeBarcode, afterBarcode, beforeCI2, afterCI2)

	if test == true {
		writer.WriteString(fmt.Sprintf("%s\tSUCCESS\t%s\t%s\t%s\t%s\t\n", ao.URI, beforeBarcode, afterBarcode, beforeCI2, afterCI2))
		writer.Flush()
		return "Test Mode - not Updating AO", nil
	} else {
		//update the ao
		msg, err := client.UpdateArchivalObject(repoId, aoID, ao)
		msg = strings.ReplaceAll(msg, "\n", "")

		if err != nil {
			WriteToLog(ao.URI, "ERROR", beforeBarcode, afterBarcode, beforeCI2, afterCI2, err.Error())
			return msg, nil
		}
		WriteToLog(ao.URI, "SUCCESS", beforeBarcode, afterBarcode, beforeCI2, afterCI2, "")
		return msg, nil
	}

}

func WriteToLog(aourl string, status string, beforeBarcode string, afterBarcode string, beforeCI2 string, afterCI2 string, errormsg string) {
	writer.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", aourl, status, beforeBarcode, afterBarcode, beforeCI2, afterCI2, errormsg))
	writer.Flush()
}

func GetTSVRows(tsv *os.File) ([]Row, error) {
	//create an empty array or Rows
	rows := []Row{}
	//create a scanner object and read the tsv file
	scanner := bufio.NewScanner(tsv)
	// skip the header line
	scanner.Scan()
	// scan the tsv file line by line
	currentRow := 1;
	for scanner.Scan() {
		//split the line by tab chars
		cols := strings.Split(scanner.Text(), "\t")
		//marshal the split line into a Row struct and add to the array of Rows
		row, err := tryParse(cols)
		if err != nil {
			fmt.Printf("\t%v line %d [%s]\n", err.Error(), currentRow, scanner.Text())
			fmt.Printf("\tA row should have 12 columns, had %d columns, skipping.\n", len(cols))
			WriteToLog(cols[2],"SKIPPED", "", "", "", "", err.Error())
		}
		rows = append(rows, row)
		currentRow = currentRow + 1;
	}
	//Check for any read errors
	if scanner.Err() != nil {
		return rows, scanner.Err()
	}

	//return the array of Rows
	return rows, nil
}

func tryParse(cols []string) (Row, error) {
	if(len(cols) != 11) {
		return Row{}, fmt.Errorf("Malformed Line Errror")
	} else {
		return Row{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6], cols[7], cols[8], cols[9], cols[10]}, nil
	}
}

func GetInstanceAsJson(instances []aspace.Instance) string {
	instanceJson, _ := json.Marshal(instances)
	return string(instanceJson)
}

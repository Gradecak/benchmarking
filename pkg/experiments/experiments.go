package experiments

import (
	"encoding/csv"
	"fmt"
	"github.com/gradecak/benchmark/pkg/benchmark"
	"os"
	"strings"
)

type Zone struct {
	expected string // zone that all tasks should execute on
	numTasks int    // number of tasks within the workflow
}

func NewZoneExperiment(nTasks int, exp string) *Zone {
	return &Zone{
		expected: exp,
		numTasks: nTasks,
	}
}

func (z Zone) Parse(results chan benchmark.Result) {
	// construct CSV header
	header := []string{"expectedZone", "responseTime"}
	for i := 0; i < z.numTasks; i++ {
		header = append(header, fmt.Sprintf("T%v", i+1))
	}

	//open file
	file, err := os.Create("zoneResult.csv")
	if err != nil {
		panic("Cannot open results file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	err = writer.Write(header)
	if err != nil {
		panic("cannot write header to csv file")
	}

	for r := range results {
		// split will return original string if separator is not found
		zones := strings.Split(r.Response, "-")

		row := []string{z.expected, r.Time.String()}
		row = append(row, zones...)
		writer.Write(row)
		// expected results NL, DE, IR etc
		//fmt.Printf("%+v", r)
	}
}

func (z Zone) Save(file string) {

}

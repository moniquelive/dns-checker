package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"gitlab.com/m8127/produtos/automation/dns-checker/checker"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	dynamic string
	static  string
)

type (
	job struct {
		source, target string
		statusCode     int
	}
	result struct {
		job     job
		success bool
		err     error
	}
)

func (j *job) execute() result {
	success, err := checker.Check(j.source, j.target, j.statusCode)
	return result{
		job:     *j,
		success: success,
		err:     err,
	}
}

func scan(filename string) error {
	records, err := parseCSV(filename)
	if err != nil {
		return err
	}
	// workers & channels & sync's
	workerPoolSize := 4
	jobsCh := make(chan job, workerPoolSize)
	resultsCh := make(chan result, len(records)-1)
	var wg sync.WaitGroup
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go worker(&wg, jobsCh, resultsCh)
	}
	// work!
	for _, row := range records[1:] { // skip header
		var source, target string
		var status int
		switch len(row) {
		case 1:
			split := strings.Split(row[0], ";")
			source, target = split[0], split[1]
			if status, err = strconv.Atoi(split[2]); err != nil {
				return err
			}
		case 3:
			source, target = row[0], row[1]
			if status, err = strconv.Atoi(row[2]); err != nil {
				return err
			}
		}
		source = strings.ReplaceAll(source, `"`, "")
		target = strings.ReplaceAll(target, `"`, "")
		log.Infof("Enqueueing: %v => %v (%v)\n", source, target, status)
		jobsCh <- job{
			source:     source,
			target:     target,
			statusCode: status,
		}
	}
	close(jobsCh)
	wg.Wait()
	close(resultsCh)
	for r := range resultsCh {
		log.Infof("Success(%v) == %v (%v)\n", r.job.source, r.success, r.err)
	}
	return nil
}

func worker(wg *sync.WaitGroup, jobsCh <-chan job, resultsCh chan<- result) {
	defer wg.Done()
	for job := range jobsCh {
		resultsCh <- job.execute()
	}
}

func parseCSV(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a batch of URLs",
	Long: `Scans a batch of urls given the CSV file with the following format:

"source";"target";"status"
"https://google.com";"https://www.google.com/";301`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("scan called: -d %#v -s %#v\n", dynamic, static)
		if dynamic != "" {
			if _, err := os.Stat(dynamic); err != nil && os.IsNotExist(err) {
				log.Errorf("The file %q does not exist. Skipping")
			} else {
				err := scan(dynamic)
				if err != nil {
					panic(err)
				}
			}
		}
		if static != "" {
			if _, err := os.Stat(static); err != nil && os.IsNotExist(err) {
				log.Errorf("The file %q does not exist. Skipping")
			} else {
				err := scan(static)
				if err != nil {
					panic(err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	scanCmd.Flags().StringVarP(&dynamic, "dynamic", "d", "", "CSV file path with dynamic urls")
	scanCmd.Flags().StringVarP(&static, "static", "s", "", "CSV file path with static urls")
}

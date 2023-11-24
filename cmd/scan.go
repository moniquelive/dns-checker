package cmd

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moniquelive/dns-checker/checker"
)

var (
	dynamic        string
	static         string
	workerPoolSize int
)

type result struct {
	originalJob jobUnit
	success     bool
	err         error
}

type jobUnit struct {
	source, target string
	statusCode     int
}

func (j jobUnit) execute() result {
	success, err := checker.Check(j.source, j.target, j.statusCode)
	return result{
		originalJob: j,
		success:     success,
		err:         err,
	}
}

func scan(jobs []jobUnit) int {
	const (
		AllGood = iota
		ErrorsFound
	)
	if len(jobs) < 1 {
		return AllGood
	}
	//
	// workers & channels & sync's
	//
	log.Debugf("Starting %d jobs...", workerPoolSize)
	jobsCh := make(chan jobUnit, workerPoolSize)
	resultsCh := make(chan result, len(jobs))
	var wg sync.WaitGroup
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go worker(&wg, jobsCh, resultsCh)
	}
	//
	// do work (fan out)
	//
	for _, job := range jobs {
		jobsCh <- job
	}
	close(jobsCh)
	//
	// collect results (fan in)
	//
	wg.Wait()
	close(resultsCh)
	retval := AllGood
	for r := range resultsCh {
		if !r.success {
			retval = ErrorsFound
			log.Errorf("%q;%q;%q", r.originalJob.source, r.originalJob.target, r.err)
		}
	}
	return retval
}

func worker(wg *sync.WaitGroup, jobsCh <-chan jobUnit, resultsCh chan<- result) {
	defer wg.Done()
	for job := range jobsCh {
		resultsCh <- job.execute()
	}
}

func parseCSV(filename string) ([]jobUnit, error) {
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
	jobs := make([]jobUnit, 0, len(records)-1)
	for _, row := range records[1:] { // skip header
		var source, target string
		var status int
		switch len(row) {
		case 1:
			split := strings.Split(row[0], ";")
			source, target = split[0], split[1]
			if status, err = strconv.Atoi(split[2]); err != nil {
				return nil, err
			}
		case 3:
			source, target = row[0], row[1]
			if status, err = strconv.Atoi(row[2]); err != nil {
				return nil, err
			}
		}
		source = strings.ReplaceAll(source, `"`, "")
		target = strings.ReplaceAll(target, `"`, "")
		log.Debugf("Enqueueing: %v => %v (%v)\n", source, target, status)
		jobs = append(jobs, jobUnit{source, target, status})
	}
	return jobs, nil
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a batch of URLs",
	Long: `Scans a batch of urls given the CSV file with the following format (the first line is always skipped):

"source";"target";"status"
"https://google.com";"https://www.google.com/";301`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().NFlag() == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var allJobs []jobUnit
		for _, filename := range []string{dynamic, static} {
			if filename == "" {
				continue
			}
			if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
				log.Fatalf("File %q not found", filename)
			}
			log.Debugf("Parsing %q", filename)
			jobs, err := parseCSV(filename)
			if err != nil {
				log.Fatalf("Error parsing file %q", filename)
			}
			allJobs = append(allJobs, jobs...)
		}
		errLevel := scan(allJobs)
		os.Exit(errLevel)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringVarP(&dynamic, "dynamic", "d", "", "CSV file path with dynamic urls")
	scanCmd.Flags().StringVarP(&static, "static", "s", "", "CSV file path with static urls")
	scanCmd.Flags().IntVarP(&workerPoolSize, "jobs", "j", 4, "Number of parallel jobs")
}

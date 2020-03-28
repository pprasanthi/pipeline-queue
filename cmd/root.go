package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/pprasanthi/job-queue/internal/client"
)

// AccessToken GitLab API access token
var AccessToken string

// Hostname GitLab hostname, such as https://myprivateinstance.com or https://gitlab.com
var Hostname string

// Interval Amount of time between intervals
var Interval time.Duration

// ProjectID The Project ID of the pipelines we care to block on
var ProjectID int

//// PipelineID The Pipeline ID the current Job belongs to
//var PipelineID int

// SingletonJobs, list of jobs that should not run concurrently.
var SingletonJobs []string

// JobID, current JobID in the pipeline.
var JobID int

func errorExit(err error) {
	fmt.Fprint(os.Stderr, "[error] ")
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func timedPrint(message string) {
	now := time.Now()
	fmt.Printf("[%v] %v\n", now.Format(time.RFC3339), message)
}

func waitItOut(duration time.Duration) {
	time.Sleep(duration)
}

var rootCmd *cobra.Command

func init() {
	exe := filepath.Base(os.Args[0])
	shortDesc := fmt.Sprintf("%s blocks until older job finish running", exe)
	longDesc := `
pipeline-queue queries the GitLab API for the current project's list of running
jobs, sorted in ascending order (oldest to newest). If the current job is not the latest, 
pipeline-queue will sleep for a given time interval before repeating the polling process. 
When the job becomes the oldest running one, it will cease to block and will exit.
`
	ciProjectID := os.Getenv("CI_PROJECT_ID")
	defaultProjectID, _ := strconv.Atoi(ciProjectID)
	defaultSingletonJobs := []string{""}
	ciJobID := os.Getenv("CI_JOB_ID")
	defaultJobId, _ := strconv.Atoi(ciJobID)

	rootCmd = &cobra.Command{
		Use:   exe,
		Short: shortDesc,
		Long:  strings.Trim(fmt.Sprintf("%s\n%s", shortDesc, longDesc), "\n"),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New(nil, Hostname, AccessToken)
			if err != nil {
				errorExit(err)
			}
			projID := strconv.Itoa(ProjectID)
			singletonJobs := SingletonJobs
			jobId := strconv.Itoa(JobID)

			fmt.Println("Jobs that should not run concurrently are: ",SingletonJobs)
			areWeFirst := false

			for !areWeFirst {
				timedPrint("Checking if we're first in line...")
				areWeFirst, err = client.DetermineIfJobIsFirst(projID, singletonJobs, jobId)
				if err != nil {
					errorExit(err)
				}

				if !areWeFirst {
					timedPrint(fmt.Sprintf("We're not first. Trying again in %v", Interval))
					waitItOut(Interval)
				}
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&AccessToken, "token", "t", os.Getenv("CI_JOB_TOKEN"), "API access token. (default $CI_JOB_TOKEN)")
	rootCmd.PersistentFlags().StringVarP(&Hostname, "hostname", "n", "https://gitlab.com", "Hostname of the Gitlab instance.")
	rootCmd.PersistentFlags().DurationVarP(&Interval, "interval-time", "i", 30*time.Second, "Amount of time to wait in-between polls in time.Duration format.")
	rootCmd.PersistentFlags().IntVarP(&ProjectID, "project", "j", defaultProjectID, "Project ID of the pipeline to run in. (default $CI_PROJECT_ID)")
	rootCmd.PersistentFlags().StringSliceVarP(&SingletonJobs, "singletonjobs", "s", defaultSingletonJobs, "List of Jobs that are not supposed to run concurrently. (default [$CI_JOB_NAME1,$CI_JOB_NAME2])")
	rootCmd.PersistentFlags().IntVarP(&JobID, "jobid", "o", defaultJobId, "Job ID of the current pipeline. (default $CI_JOB_ID)")
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

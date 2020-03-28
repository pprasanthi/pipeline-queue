package client

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/plouc/go-gitlab-client/gitlab"
)

// GitLabClient - Simplified interface for a GitLab client to wrap
type GitLabClient interface {
	ProjectJobs(string, *gitlab.JobsOptions) (*gitlab.JobCollection, *gitlab.ResponseMeta, error)
	ProjectJob(string, int) (*gitlab.Job, *gitlab.ResponseMeta, error)
}

// Client - Wrapper struct for a GitLab client
type Client struct {
	Client GitLabClient
}

// ListRunningJobs - Query the GitLab API for a Project's list of Jobs, sorted
//						  in ascending order (oldest to newest)
func (c *Client) ListRunningJobs(projectID string, singletonJobs []string) ([]*gitlab.Job, error) {
	var (
		details *gitlab.Job
		singletonJobsRunning []*gitlab.Job
	)
	options := &gitlab.JobsOptions{
		Scope:     []string{"running"},
		SortOptions: gitlab.SortOptions{
			OrderBy: "id",
			Sort:    gitlab.SortDirectionAsc,
		},
	}

    //Collect all running jobs.
	jobs, _, err := c.Client.ProjectJobs(projectID, options)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return nil, err
	}

	//Get the count of running jobs that match the singleton jobName
	for _, job := range jobs.Items{
		_, isSingletonJob := Find(singletonJobs,job.Name)
		if isSingletonJob {
			singletonJobsRunning = append(singletonJobsRunning, job)
		}
	}
	result := make([]*gitlab.Job, len(singletonJobsRunning))

	//Collect job details of the singleton jobs.
	for index, job := range jobs.Items {
		_, isSingletonJob := Find(singletonJobs,job.Name)
		if isSingletonJob {
			details, _, err = c.Client.ProjectJob(projectID, job.Id)
			if err != nil {
				fmt.Printf("Error: %v", err)
				return nil, err
			}
			result[index] = details
		}
	}
	return result, nil
}

// SortJobsByStartedAt - Given a slice of JobDetails, sort them by the StartedAt attribute
//				   Sorts the slice in-place (e.g., modifies the slice passed in)
func (c *Client) SortJobsByStartedAt(jobs []*gitlab.Job) []*gitlab.Job {
	sort.Slice(jobs, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, jobs[i].StartedAt)
		timeJ, _ := time.Parse(time.RFC3339, jobs[j].StartedAt)

		return timeI.Before(timeJ)
	})

	fmt.Println("Project has following singleton jobs running at the moment:")

	for index, job := range jobs {
		fmt.Println(fmt.Sprintf("JobId: %d, JobName: %s, at Index %d", job.Id, job.Name, index))
	}
	return jobs
}

// IndexOfJob - Given a sorted list of Jobs, determine if the given Job name is
//					 first in the list (0th index).
func (c *Client) IndexOfJob(jobs []*gitlab.Job, jobId string) (int, error) {
	targetID, err := strconv.Atoi(jobId)
	if err != nil {
		return -1, err
	}

	for i, job := range jobs {
		if job.Id == targetID {
			fmt.Println(fmt.Sprintf("Our Job %d is at index %d", targetID,i))
			return i, nil
		}
	}

	return -1, fmt.Errorf("JobID %s not found in collection", jobId)
}



// DetermineIfFirst - Checks to see if the Job's Pipeline is the oldest running pipeline
//                    for the Project.
func (c *Client) DetermineIfJobIsFirst(projectID string, singletonJobs []string, jobId string) (bool, error) {
	jobs, err := c.ListRunningJobs(projectID, singletonJobs)
	if err != nil {
		return false, err
	}

	c.SortJobsByStartedAt(jobs)

	position, err := c.IndexOfJob(jobs, jobId)
	if err != nil {
		return false, err
	}

	return position == 0, err
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}


// New - Factory method for creating a new GitLab client wrapper
func New(desiredClient GitLabClient, hostname string, token string) (*Client, error) {
	var gitlabClient GitLabClient

	if desiredClient != nil {
		gitlabClient = desiredClient
	} else {
		gitlabClient = gitlab.NewGitlab(hostname, "/api/v4", token)
	}

	return &Client{
		Client: gitlabClient,
	}, nil

}

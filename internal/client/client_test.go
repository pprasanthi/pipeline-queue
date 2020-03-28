package client_test

import (
	"testing"

	"github.com/plouc/go-gitlab-client/gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/pprasanthi/job-queue/internal/client"
)

type ClientTestSuite struct {
	suite.Suite
}

type MockGitlabClient struct {
	mock.Mock
}



func (m *MockGitlabClient) ProjectJobs(projectId string, opts *gitlab.JobsOptions) (*gitlab.JobCollection, *gitlab.ResponseMeta, error) {
	args := m.Called(projectId)
	var collection *gitlab.JobCollection = args.Get(0).(*gitlab.JobCollection)

	return collection, nil, args.Error(2)
}

func (m *MockGitlabClient) ProjectJob(projectID string, jobID int) (*gitlab.Job, *gitlab.ResponseMeta, error) {
	args := m.Called(projectID, jobID)

	var job *gitlab.Job = args.Get(0).(*gitlab.Job)

	return job, nil, args.Error(2)
}

func (suite *ClientTestSuite) TestCreateClient() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, err := client.New(fakeClient, "", "")

	assert.Nil(err, "Got error: %v", err)
	assert.NotNil(c, "Client is Nil")
}




func (suite *ClientTestSuite) TestListJobes() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, err := client.New(fakeClient, "", "")

	testJob := &gitlab.Job{
		Id: 5318499,
		Name: "test",
		StartedAt: "2018-08-08T22:45:23.801Z",
	}
	testCollection := &gitlab.JobCollection{
		Items: []*gitlab.Job{
			testJob,
		},
	}
	fakeClient.On("ProjectJobs", "103").Return(testCollection, nil, nil)
	fakeClient.On("ProjectJob", "103", 5318499).Return(testJob, nil, nil)
	jobs, err := c.ListRunningJobs("103", []string{"test","test2"})
	assert.Nil(err, "Got error: %v", err)
	assert.Len(jobs, 1, "Pipeline count was not accurate: %v", jobs)
	assert.Equal(testJob, jobs[0], "Unexpected pipeline: %v", jobs[0])
}

func (suite *ClientTestSuite) TestSortByStarted() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	jobs := []*gitlab.Job{
		&gitlab.Job{
			Id: 5318499,
			Name: "test",
			StartedAt: "2018-08-08T22:01:23.801Z",
		},
		&gitlab.Job{
			Id: 5318500,
			Name: "test",
			StartedAt: "2018-08-08T22:02:23.801Z",
		},
		&gitlab.Job{
			Id: 5318501,
			Name: "test",
			StartedAt: "2018-08-08T22:03:23.801Z",
		},
		&gitlab.Job{
			Id: 5318611,
			Name: "test",
			StartedAt: "2018-08-08T22:04:23.801Z",
		},
	}

	c.SortJobsByStartedAt(jobs)
	assert.Equal(jobs[2].Id, 5318501, "Incorrect pipeline at index 2, got %v", jobs[2].Id)
}

func (suite *ClientTestSuite) TestIndexOfJob() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	jobs := []*gitlab.Job{
		&gitlab.Job{
			Id: 5318499,
			Name: "test",
			StartedAt: "2018-08-08T22:01:23.801Z",
		},
		&gitlab.Job{
			Id: 5318500,
			Name: "test",
			StartedAt: "2018-08-08T22:02:23.801Z",
		},
		&gitlab.Job{
			Id: 5318501,
			Name: "test",
			StartedAt: "2018-08-08T22:03:23.801Z",
		},
		&gitlab.Job{
			Id: 5318611,
			Name: "test",
			StartedAt: "2018-08-08T22:04:23.801Z",
		},
	}

	targetIndex, err := c.IndexOfJob(jobs, "5318501")

	assert.Nil(err, "Got error: %v", err)
	assert.Equal(targetIndex, 2, "Expected index 2, got %v", targetIndex)
}

func (suite *ClientTestSuite) TestDetermineIfJobIsFirst() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	jobs := &gitlab.JobCollection{
		Items: []*gitlab.Job{
			&gitlab.Job{
				Id:        5318499,
				Name:      "test",
				StartedAt: "2018-08-08T22:01:23.801Z",
			},
			&gitlab.Job{
				Id:        5318500,
				Name:      "test",
				StartedAt: "2018-08-08T22:02:23.801Z",
			},
		},
	}

	job := []*gitlab.Job{
		&gitlab.Job{
			Id: 5318499,
			Name: "test",
			StartedAt: "2018-08-08T22:01:23.801Z",
		},
		&gitlab.Job{
			Id: 5318500,
			Name: "test",
			StartedAt: "2018-08-08T22:02:23.801Z",
		},
	}

	fakeClient.On("ProjectJobs", "987").Return(jobs, nil, nil)
	fakeClient.On("ProjectJob", "987", 5318499).Return(job[0], nil, nil)
	fakeClient.On("ProjectJob", "987", 5318500).Return(job[1], nil, nil)

	singletonjob := []string{"test"}
	isFirst, err := c.DetermineIfJobIsFirst("987",singletonjob,"5318499")

	assert.Nil(err, "Got error: %v", err)
	assert.True(isFirst, "Expected to be first, but was: %v", isFirst)
}


//func (suite *ClientTestSuite) TestDetermineIfFirst() {
//	assert := assert.New(suite.T())
//	fakeClient := new(MockGitlabClient)
//	c, _ := client.New(fakeClient, "", "")
//
//	testCollection := &gitlab.PipelineCollection{
//		Items: []*gitlab.Pipeline{
//			&gitlab.Pipeline{
//				Id: 1234,
//			},
//			// Although it's second, the timestamp comes first
//			&gitlab.Pipeline{
//				Id: 1027,
//			},
//		},
//	}
//	detailedPipelines := []*gitlab.PipelineWithDetails{
//		&gitlab.PipelineWithDetails{
//			Pipeline: gitlab.Pipeline{
//				Id: 1027,
//			},
//			UpdatedAt: "2018-08-08T22:27:23.801Z",
//		},
//		&gitlab.PipelineWithDetails{
//			Pipeline: gitlab.Pipeline{
//				Id: 1234,
//			},
//			UpdatedAt: "2018-08-08T22:59:23.801Z",
//		},
//	}
//	fakeClient.On("ProjectPipelines", "987").Return(testCollection, nil, nil)
//	fakeClient.On("ProjectPipeline", "987", "1027").Return(detailedPipelines[0], nil, nil)
//	fakeClient.On("ProjectPipeline", "987", "1234").Return(detailedPipelines[1], nil, nil)
//
//	isFirst, err := c.DetermineIfFirst("987", "1027")
//
//	assert.Nil(err, "Got error: %v", err)
//	assert.True(isFirst, "Expected to be first, but was: %v", isFirst)
//}
//
func TestClienTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

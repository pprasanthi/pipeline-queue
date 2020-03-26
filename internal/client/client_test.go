package client_test

import (
	"testing"

	"github.com/plouc/go-gitlab-client/gitlab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/pprasanthi/pipeline-queue/internal/client"
)

type ClientTestSuite struct {
	suite.Suite
}

type MockGitlabClient struct {
	mock.Mock
}

func (m *MockGitlabClient) ProjectPipelines(projectId string, opts *gitlab.PipelinesOptions) (*gitlab.PipelineCollection, *gitlab.ResponseMeta, error) {
	args := m.Called(projectId)
	var collection *gitlab.PipelineCollection = args.Get(0).(*gitlab.PipelineCollection)

	return collection, nil, args.Error(2)
}

func (m *MockGitlabClient) ProjectPipeline(projectID, pipelineID string) (*gitlab.PipelineWithDetails, *gitlab.ResponseMeta, error) {
	args := m.Called(projectID, pipelineID)

	var pipeline *gitlab.PipelineWithDetails = args.Get(0).(*gitlab.PipelineWithDetails)

	return pipeline, nil, args.Error(2)
}

func (suite *ClientTestSuite) TestCreateClient() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, err := client.New(fakeClient, "", "")

	assert.Nil(err, "Got error: %v", err)
	assert.NotNil(c, "Client is Nil")
}

func (suite *ClientTestSuite) TestListPipelines() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, err := client.New(fakeClient, "", "")

	testPipeline := &gitlab.Pipeline{
		Id: 1027,
	}
	testCollection := &gitlab.PipelineCollection{
		Items: []*gitlab.Pipeline{
			testPipeline,
		},
	}
	fakeClient.On("ProjectPipelines", "103").Return(testCollection, nil, nil)
	testPipelineDetails := &gitlab.PipelineWithDetails{
		Pipeline:  *testPipeline,
		UpdatedAt: "2018-08-08T22:45:23.801Z",
	}
	fakeClient.On("ProjectPipeline", "103", "1027").Return(testPipelineDetails, nil, nil)

	pipelines, err := c.ListRunningPipelines("103")

	assert.Nil(err, "Got error: %v", err)
	assert.Len(pipelines, 1, "Pipeline count was not accurate: %v", pipelines)
	assert.Equal(testPipelineDetails, pipelines[0], "Unexpected pipeline: %v", pipelines[0])
}

func (suite *ClientTestSuite) TestSortByUpdated() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	pipelines := []*gitlab.PipelineWithDetails{
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 0,
			},
			UpdatedAt: "2018-08-08T22:01:23.801Z",
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 333,
			},
			UpdatedAt: "2018-08-08T22:02:23.801Z",
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1027,
			},
			UpdatedAt: "2018-08-08T22:05:23.801Z",
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1234,
			},
			UpdatedAt: "2018-08-08T22:04:23.801Z",
		},
	}

	c.SortByUpdated(pipelines)

	assert.Equal(pipelines[3].Id, 1027, "Incorrect pipeline at index 3, got %v", pipelines[3].Id)
}

func (suite *ClientTestSuite) TestIndexOfPipeline() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	testCollection := []*gitlab.PipelineWithDetails{
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 0,
			},
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 333,
			},
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1027,
			},
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1234,
			},
		},
	}

	targetIndex, err := c.IndexOfPipeline(testCollection, "1027")

	assert.Nil(err, "Got error: %v", err)
	assert.Equal(targetIndex, 2, "Expected index 2, got %v", targetIndex)
}

func (suite *ClientTestSuite) TestDetermineIfFirst() {
	assert := assert.New(suite.T())
	fakeClient := new(MockGitlabClient)
	c, _ := client.New(fakeClient, "", "")

	testCollection := &gitlab.PipelineCollection{
		Items: []*gitlab.Pipeline{
			&gitlab.Pipeline{
				Id: 1234,
			},
			// Although it's second, the timestamp comes first
			&gitlab.Pipeline{
				Id: 1027,
			},
		},
	}
	detailedPipelines := []*gitlab.PipelineWithDetails{
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1027,
			},
			UpdatedAt: "2018-08-08T22:27:23.801Z",
		},
		&gitlab.PipelineWithDetails{
			Pipeline: gitlab.Pipeline{
				Id: 1234,
			},
			UpdatedAt: "2018-08-08T22:59:23.801Z",
		},
	}
	fakeClient.On("ProjectPipelines", "987").Return(testCollection, nil, nil)
	fakeClient.On("ProjectPipeline", "987", "1027").Return(detailedPipelines[0], nil, nil)
	fakeClient.On("ProjectPipeline", "987", "1234").Return(detailedPipelines[1], nil, nil)

	isFirst, err := c.DetermineIfFirst("987", "1027")

	assert.Nil(err, "Got error: %v", err)
	assert.True(isFirst, "Expected to be first, but was: %v", isFirst)
}

func TestClienTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

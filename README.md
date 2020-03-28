# pipeline-queue

A naive solution to ensuring a singleton job queue for GitLab pipelines.

## Usage

```
# full command with default values as explicit options
$ job-queue --hostname https://gitlab.com --interval-time 30s --jobname $CI_JOB_NAME --project $CI_PROJECT_ID --token $GITLAB_API_TOKEN

# full command with using shorthand flags
$ job-queue -n https://gitlab.com -i 30s -l $CI_JOB_NAME -j $CI_PROJECT_ID -t $GITLAB_API_TOKEN

# equivalent abbreviated version of the above
$ job-queue -t $GITLAB_API_TOKEN
```
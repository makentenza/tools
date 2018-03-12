package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// GitlabProject to define the structure to store each repository name and git format url
type GitlabProject struct {
	Name    string `json:"name"`
	Repourl string `json:"http_url_to_repo"`
}

// Function to check for errors. Existing with 0 code to avoid re-swpan on k8s CronJobs
func check(e error) {
	if e != nil {
		fmt.Println("error:", e)
		os.Exit(0)
	}
}

func main() {
	// Required when not using trusted certificates on GitLab
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Load required variables from environment vars
	sourceGitlab := os.Getenv("SOURCE_GITLAB_URL")
	sourceToken := os.Getenv("SOURCE_GITLAB_TOKEN")
	sourceUser := os.Getenv("SOURCE_GITLAB_USER")
	destRepo := os.Getenv("DEST_REPOSITORY")
	destToken := os.Getenv("DEST_GITLAB_TOKEN")
	destUser := os.Getenv("DEST_GITLAB_USER")

	// Rewrite destination repository git URL to include auth credentials
	authDestURL := strings.Replace(destRepo, "https://", "https://"+destUser+":"+destToken+"@", 1)
	// Generate GitLab API query from env vars
	apiquery := sourceGitlab + "/api/v4/projects?private_token=" + sourceToken

	// Consume projects from GitLab API
	response, err := http.Get(apiquery)
	check(err)

	// Store projects JSON Array
	responseData, err := ioutil.ReadAll(response.Body)
	check(err)

	// Create an Array of structures to store every repository info
	var GitLabProjects []GitlabProject

	// Transform the API consumed JSON representation
	err1 := json.Unmarshal(responseData, &GitLabProjects)
	check(err1)

	// Clone destination repository
	destDir := os.Getenv("TMPDIR") + "dest"
	git.PlainClone(destDir, false, &git.CloneOptions{
		URL:               authDestURL,
		Progress:          os.Stdout,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	// Iterate in every project and clone them to separate directories
	for k := range GitLabProjects {
		authSourceURL := strings.Replace(GitLabProjects[k].Repourl, "https://", "https://"+sourceUser+":"+sourceToken+"@", 1)
		path := destDir + "/" + GitLabProjects[k].Name
		fmt.Println("Cloning repository", GitLabProjects[k].Name)
		git.PlainClone(path, false, &git.CloneOptions{
			URL:               authSourceURL,
			Progress:          os.Stdout,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
	}

	// Open repository and worktree
	r, err := git.PlainOpen(destDir)
	check(err)
	w, err := r.Worktree()
	check(err)

	// Get status for untracked files to be added
	status, err := w.Status()
	check(err)
	fmt.Println("Logging repository status:")
	fmt.Println(status)

	// Add any repository change to the commit
	for path := range status {
		w.Add(path)
	}

	// Commit and Push to origin remote for the destination repository
	commit, err := w.Commit("Git mirror automation task", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Labs SRE Team",
			Email: "rhc-labs-sre@redhat.com",
			When:  time.Now(),
		},
	})
	obj, err := r.CommitObject(commit)
	check(err)
	fmt.Println("Commiting changes:")
	fmt.Println(obj)

	err = r.Push(&git.PushOptions{})
	check(err)

	// Remove directory used for cloning repositories
	os.RemoveAll(destDir)

	// Existing with 0 code to avoid re-swpan on k8s CronJobs
	//os.Exit(0)

	// Run forever workaround for using k8s RC instead jobs
	select {}
}

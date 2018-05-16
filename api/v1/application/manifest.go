package application

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/ovh/lhasa/api/config"
	"github.com/ovh/lhasa/api/ext/bitbucket"
	"github.com/ovh/lhasa/api/v1"
)

// MetaAssistant create a manifest with a pull request
type MetaAssistant func(application v1.ApplicationVersion, parameters *PullRequest) error

// PullRequest : pull request struct
type PullRequest struct {
	// Repository name
	Repository string
	// Manifest structure
	Manifest map[string]interface{}
	// Pull request owner
	Creator string
	// Manifest filename
	ManifestName string
}

// GitMetaAssistant help to create manifest on bitbucket
func GitMetaAssistant(depRepo *Repository) MetaAssistant {
	return func(app v1.ApplicationVersion, parameters *PullRequest) error {
		switch parameters.Repository {
		case "bitbucket":
			return CreatePullRequestBitBucket(parameters)
		default:
			logrus.WithFields(logrus.Fields{
				"parameters": parameters,
			}).Error("Unknown assistant")
			return errors.New("Unknown assistant")
		}
	}
}

// CreatePullRequestBitBucket create a pull request on bitbucket
func CreatePullRequestBitBucket(parameters *PullRequest) error {
	// Create client
	conf := config.ExtractValue("bitbucket").(map[string]interface{})
	client := bitbucket.NewAccessToken(conf["token"].(string))

	// Extract repo and project from repository url
	var project = strings.Split(parameters.Manifest["repository"].(string), "/")[4]
	var repo = strings.Split(parameters.Manifest["repository"].(string), "/")[6]

	logrus.WithFields(logrus.Fields{
		"project":    project,
		"repository": repo,
	}).Info("Create branch")

	// Create branch
	branch, errBranch := client.Repositories.Branch.Create(&bitbucket.BranchOptions{
		Owner:      project,
		RepoSlug:   repo,
		Name:       "branch-to-update-manifest",
		StartPoint: "master",
		Message:    "Message de test",
	})
	if errBranch != nil {
		logrus.WithFields(logrus.Fields{
			"error": errBranch,
		}).Error("Create branch")
		return errBranch
	}
	var _checkJSONManifest = make(map[string]interface{})
	bin, _ := json.Marshal(parameters.Manifest)
	json.Unmarshal(bin, &_checkJSONManifest)

	jsonMapBranch := branch.(map[string]interface{})

	logrus.WithFields(logrus.Fields{
		"branch": jsonMapBranch,
	}).Info("Create branch")

	// build content with pretty print
	content, _ := json.MarshalIndent(_checkJSONManifest, "", "\t")

	logrus.WithFields(logrus.Fields{
		"manifest": string(content),
	}).Info("Create path")

	path, errPath := client.Repositories.Path.Create(&bitbucket.PathOptions{
		Owner:    project,
		RepoSlug: repo,
		Name:     parameters.ManifestName,
		Message:  "Create manifest initial version",
		Branch:   "branch-to-update-manifest",
		Content:  string(content),
	})
	if errPath != nil {
		logrus.WithFields(logrus.Fields{
			"error": errPath,
		}).Error("Create path")
		return errPath
	}
	jsonMapPath := path.(map[string]interface{})

	logrus.WithFields(logrus.Fields{
		"path":     jsonMapPath,
		"commitId": jsonMapPath["id"],
	}).Info("Create path")

	res, errPullRequest := client.Repositories.PullRequests.Create(&bitbucket.PullRequestsOptions{
		Owner:        project,
		RepoSlug:     repo,
		SourceBranch: "branch-to-update-manifest",
		Title:        "Update manifest data",
		Reviewers:    []string{parameters.Creator},
	})
	if errPullRequest != nil {
		logrus.WithFields(logrus.Fields{
			"error": errPullRequest,
		}).Error("Create pull request")
		return errPullRequest
	}
	jsonMapPull := res.(map[string]interface{})

	logrus.WithFields(logrus.Fields{
		"result": jsonMapPull,
	}).Info("Pull request creation")

	return nil
}

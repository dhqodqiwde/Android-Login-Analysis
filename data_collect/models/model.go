package models

import "time"

type GithubRepoInfo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type GithubIssueInfo struct {
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Url        string    `json:"url"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type Dependency struct {
	PackageName string `json:"package_name"`
	MethodName  string `json:"method_name"`
	Version     string `json:"version"`
	Number      int    `json:"number"`
}

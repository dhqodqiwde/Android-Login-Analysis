package util

import (
	"context"
	"fmt"
	"data_collection/models"
	"sync"
	"time"

	"github.com/google/go-github/v55/github"
)

func GetIssues(token string, github_repo_infos []models.GithubRepoInfo) []*github.Issue {
	list_options := github.ListOptions{
		Page:    1,
		PerPage: 99,
	}
	client := github.NewClient(nil).WithAuthToken(token)
	// fmt.Printf("client:%v", client)
	opt := &github.IssueListByRepoOptions{
		// Milestone: "*",
		// Assignee:  "none",
		// Labels:      []string{"bug"},
		Sort:      "created",
		Direction: "desc",
		// Since:       since, // This is a timestamp in ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ
		ListOptions: list_options,
	}
	var allIssues []*github.Issue
	for _, info := range github_repo_infos {
		var wg sync.WaitGroup
		var allIssues_closed, allIssues_open []*github.Issue
		time.Sleep(15 * time.Second)
		wg.Add(1)
		go func() {
			defer wg.Done()
			allIssues_closed = ListIssuesByState(opt, "closed", client, info, allIssues_closed)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			allIssues_open = ListIssuesByState(opt, "open", client, info, allIssues_open)
		}()
		wg.Wait()
		allIssues = append(allIssues, append(allIssues_closed, allIssues_open...)...)
		fmt.Println(info.Owner+"/"+info.Name+": ", len(allIssues_closed)+len(allIssues_open), "total length: ", len(allIssues))
	}
	return allIssues
}

func ListIssuesByState(opt *github.IssueListByRepoOptions, state string, client *github.Client, info models.GithubRepoInfo, allIssues []*github.Issue) []*github.Issue {
	opt.Page = 0
	for {
		ctx := context.Background()
		opt.State = state
		issues, resp, err := client.Issues.ListByRepo(ctx, info.Owner, info.Name, opt)
		if err != nil {
			fmt.Printf("ListByRepo Err:%v \n", err)
		}
		issues_filtered := make([]*github.Issue, 0)
		for _, issue := range issues {
			if issue.Body != nil {
				issues_filtered = append(issues_filtered, issue)
			}
		}
		allIssues = append(allIssues, issues_filtered...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allIssues
}

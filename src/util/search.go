package util

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v55/github"
)

// Search Repositories
func SearchRepos(token string, query string) []*github.Repository {
	client := github.NewClient(nil).WithAuthToken(token)
	client.Client().Timeout = 10 * time.Second
	ctx := context.Background()
	list_options := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	opt := &github.SearchOptions{
		Sort:        "stars",
		TextMatch:   true,
		ListOptions: *list_options,
	}
	search_results := make([]*github.Repository, 0)
	for {
		time.Sleep(2 * time.Second)
		result, resp, err := client.Search.Repositories(ctx, query, opt)
		if err != nil {
			fmt.Printf("Search.Repositories Err:%v, \n", err)
		}
		if resp.NextPage == 0 {
			break
		}
		search_results = append(search_results, result.Repositories...)
		opt.Page = resp.NextPage
	}
	return search_results
}

// Search Issues
func SearchIssues(token string, query string) []github.IssuesSearchResult {
	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()
	list_options := &github.ListOptions{
		Page:    1,
		PerPage: 10,
	}
	opt := &github.SearchOptions{
		Sort:        "comments",
		TextMatch:   true,
		ListOptions: *list_options,
	}
	search_results := make([]github.IssuesSearchResult, 0)
	for {
		result, resp, err := client.Search.Issues(ctx, query, opt)
		if err != nil {
			fmt.Printf("Search.Issues Err:%v, \n", err)
		}
		if resp.NextPage == 0 {
			break
		}
		search_results = append(search_results, *result)
	}
	return search_results

}

// Search code in specific repo and file
func SearchCode(token, query, repo_name string) []github.CodeSearchResult {
	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()
	list_options := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	opt := &github.SearchOptions{
		Sort:        "indexed",
		TextMatch:   true,
		ListOptions: *list_options,
	}
	search_results := make([]github.CodeSearchResult, 0)
	for {
		result, resp, err := client.Search.Code(ctx, query, opt)
		if err != nil {
			fmt.Printf("Search.Issues Err:%v, \n", err)
		}
		if resp.NextPage == 0 {
			break
		}
		search_results = append(search_results, *result)
	}
	return search_results
}

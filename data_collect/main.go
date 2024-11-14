package main

import (
	"data_collection/models"
	"data_collection/util"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v55/github"
)

// token
const (
	// TODO: insert your own token here
	token = ""
)

// input file
const (
	file_path                   = "input/github_repos.txt"
	repos_search_topic          = "input/repos_topics.txt"
	repos_search_filter         = "input/repos_filter.txt"
	dependenies_from_build_file = "input/input_dependencies.txt"
	exist_packages              = "input/exist_packages.txt"
	unrelated_packages_path     = "input/exist_unrelated_packages.txt"
	repos_clone_path            = "input/repos_clone.txt"
)

// output file
const (
	output_path                       = "output/github_issues.txt"
	words_count_file_path             = "output/words_count.txt"
	TF_IDF_path                       = "output/TF_IDF.txt"
	searched_repos_path               = "output/searched_repos.txt"
	repos_cantain_aimed_packages_path = "output/repos_cantain_aimed_packages.txt"
	repos_with_new_dependencies_path  = "output/repos_with_new_dependencies.txt"
	all_new_dependencies_path         = "output/all_new_dependencies.txt"
)

// For repos search
func main() {
	// Search repos
	topics := util.ReadWordsList(repos_search_topic)
	search_filter := util.ReadWordsList(repos_search_filter)
	// YYYY-MM-DD
	searched_repos := make([]*github.Repository, 0)
	for _, topic := range topics {
		query := topic + " stars:>50" + " pushed:>=2023-01-01" + " language:java language:kotlin"
		searched_repos = append(searched_repos, util.SearchRepos(token, query)...)
	}
	repo_name_and_owners := util.GetReposNameAndOwnerList(util.SortReposByStars(util.FilterRepos(searched_repos, search_filter)))
	util.WriteReposInfoFile(searched_repos_path, repo_name_and_owners)
}

// For TFIDF
func main1() {
	github_repo_infos := util.ReadGithubRepoInfo(file_path)
	issues := util.GetIssues(token, github_repo_infos)
	// fmt.Printf("issues:%+v", issues[0])
	issue_infos := make([]models.GithubIssueInfo, 0)
	issue_body_words_matrix := [][]string{}
	for _, issue := range issues {
		if issue.Body != nil {
			issue_infos = append(issue_infos, models.GithubIssueInfo{Title: *issue.Title,
				Body: *issue.Body, Url: *issue.HTMLURL, UpdateTime: issue.UpdatedAt.Time,
				CreateTime: issue.CreatedAt.Time})
			issue_body_words := strings.Split(*issue.Body, " ")

			issue_body_words_matrix = append(issue_body_words_matrix, issue_body_words)
		}
	}
	var (
		wg sync.WaitGroup
	)
	fmt.Println("lenissue_infos:%v", len(issue_infos))
	util.WriteIssuesFile(output_path, issue_infos)

	wg.Add(1)
	go func() {
		defer wg.Done()
		// words counter
		util.CountTestBase(output_path, words_count_file_path)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		TF_IDF := new(util.TfidfUtil)
		TF_IDF.Fit(issue_body_words_matrix, TF_IDF_path)
	}()

	wg.Wait()

}

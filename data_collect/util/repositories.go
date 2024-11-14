package util

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
)

// Get repositories by repo_ids and return map of repositories
func GetRepositoriesByIDs(token string, repo_ids []int64) map[int64]github.Repository {
	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()
	repos_map := make(map[int64]github.Repository)
	for _, id := range repo_ids {
		repo, _, err := client.Repositories.GetByID(ctx, id)
		if err != nil {
			fmt.Printf("Get Repositories id:%v,Err:%v, \n", id, err)
		}
		repos_map[id] = *repo
	}
	return repos_map
}

// Deduplicate and filter repos with same name
func FilterRepos(repos []*github.Repository, filter_words []string) []*github.Repository {
	repos_map := make(map[string]*github.Repository, 0)
	repos_return := make([]*github.Repository, 0)
	for _, repo := range repos {
		if (repo.Name != nil && repo.Owner.Name != nil) || repo.FullName != nil {
			needFiltered := false
			for _, filter_word := range filter_words {
				needFiltered = strings.Contains(*repo.FullName, filter_word)
				if needFiltered {
					break
				}
			}
			if _, ok := repos_map[*repo.FullName]; !ok && !needFiltered {
				repos_return = append(repos_return, repo)
				repos_map[*repo.FullName] = repo
			}
		}
	}
	return repos_return
}

// Sort repos by stars and latest push date
func SortReposByStars(repos []*github.Repository) []*github.Repository {
	sort.SliceStable(repos, func(i, j int) bool {
		if repos[i].GetStargazersCount() != repos[j].GetStargazersCount() {
			return repos[i].GetStargazersCount() > repos[j].GetStargazersCount()
		}
		return repos[i].GetPushedAt().After(repos[j].GetPushedAt().Time)
	})
	return repos
}

// Get full name
func GetReposNameAndOwnerList(repos []*github.Repository) []*string {
	reposInfo := make([]*string, 0)
	for _, repo := range repos {
		reposInfo = append(reposInfo, repo.FullName)
	}
	return reposInfo
}

// Get repositories content with specific path,size must smaller than 100mb
func GetRepositoryContent(token, owner, name string, paths []string) []string {
	client := github.NewClient(nil).WithAuthToken(token)
	client.Client().Timeout = 10 * time.Second
	ctx := context.Background()
	opt := &github.RepositoryContentGetOptions{}
	content := make([]string, 0)
	for _, path := range paths {
		// we do not need dictionary content
		file_content, _, _, err := client.Repositories.GetContents(ctx, owner, name, path, opt)
		if err != nil {
			fmt.Printf("Get Repositories content owner/name:%v,Err:%v, \n", owner+"/"+name, err)
		}
		str_file_content, err := file_content.GetContent()
		if err != nil {
			fmt.Printf("Decode file content Err:%v, \n", err)
		}
		content = append(content, str_file_content)
	}
	return content
}

// Download Repository
func DownloadRepository(token, owner, name string, paths []string) []string {
	client := github.NewClient(nil).WithAuthToken(token)
	client.Client().Timeout = 10 * time.Second
	ctx := context.Background()
	opt := &github.RepositoryContentGetOptions{}
	content := make([]string, 0)
	for _, path := range paths {
		// we do not need dictionary content
		readCloser, file_content, _, err := client.Repositories.DownloadContentsWithMeta(ctx, owner, name, path, opt)
		if err != nil {
			fmt.Printf("Get Repositories content owner/name:%v,Err:%v, \n", owner+"/"+name, err)
		}
		str_file_content, err := file_content.GetContent()
		if err != nil {
			fmt.Printf("Decode file content Err:%v, \n", err)
		}
		content = append(content, str_file_content)
		readCloser.Close()
	}
	return content
}

// Clone Repo
func CloneRepository(ownerAndName string) {
	// TODO Add your own repo path
	git_clone := "" + ownerAndName + ".git"
	clone_command := exec.Command("git", "clone", git_clone)
	clone_command.Dir = "./repos"
	err := clone_command.Run()
	if err != nil {
		fmt.Println("Execute Command failed:"+err.Error(), "\nownerAndName:", ownerAndName)
		return
	}
}

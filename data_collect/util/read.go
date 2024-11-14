package util

import (
	"io"
	"os"
	"strconv"
	"strings"

	"data_collection/models"
)

func ReadGithubRepoInfo(file_path string) []models.GithubRepoInfo {
	file, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, _ := io.ReadAll(file)
	content_str := string(content)
	github_repo_infos := make([]models.GithubRepoInfo, 0)
	content_detail := strings.Split(content_str, "\n")
	for _, line := range content_detail {
		info := strings.Split(line, " ")
		if len(info) >= 2 {
			github_repo_infos = append(github_repo_infos, models.GithubRepoInfo{Name: info[1], Owner: info[0]})
		}
	}
	return github_repo_infos
}

// Read the searching words from txt
func ReadWordsList(file_path string) []string {
	file, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, _ := io.ReadAll(file)
	content_str := string(content)
	content_detail := strings.Split(content_str, "\n")
	return content_detail
}

// Read dependencies from build file into Dependency Struct
func ReadDependenciesFromBuildFile(file_path string) map[string][]string {
	file, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, _ := io.ReadAll(file)
	content_str := string(content)
	content_detail := strings.Split(content_str, "\n")
	all_dependencies := make(map[string][]string, 0)
	dependencies := make([]string, 0)
	repoName := ""
	for _, line := range content_detail {
		// keywords indicates following dependencies from a new repo
		if strings.Contains(line, "DependenciesAndProjectName:") {
			info := strings.Split(line, " ")
			if len(info) >= 2 {
				if len(dependencies) > 0 {
					all_dependencies[repoName] = dependencies
					dependencies = make([]string, 0)
				}
				repoName = info[1]
				continue
			}
		}
		dependencies = append(dependencies, line)
	}
	all_dependencies[repoName] = dependencies
	return all_dependencies
}

// Read dependencies from build file into Dependency Struct, return existed dependencies list and map
func ReadExistDependencies(file_path string) ([]models.Dependency, map[string]models.Dependency) {
	file, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, _ := io.ReadAll(file)
	content_str := string(content)
	existed_dependencies := make([]models.Dependency, 0)
	existed_dependencies_map := make(map[string]models.Dependency, 0)
	content_detail := strings.Split(content_str, "\n")
	for _, line := range content_detail {
		info := strings.Split(line, " ")
		number := 0
		if len(info) >= 3 {
			number, _ = strconv.Atoi(info[3])
			new_dependency := models.Dependency{PackageName: info[0],
				MethodName: info[1], Version: info[2], Number: number}
			existed_dependencies = append(existed_dependencies, new_dependency)
			existed_dependencies_map[new_dependency.PackageName] = new_dependency
		}
	}
	return existed_dependencies, existed_dependencies_map
}

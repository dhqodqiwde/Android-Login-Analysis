package util

import (
	"bufio"
	"fmt"
	"data_collection/models"
	"os"
	"regexp"
	"strconv"
)

func WriteIssuesFile(filePath string, issues []models.GithubIssueInfo) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	for i, issue := range issues {
		write.WriteString("No." + strconv.Itoa(i) + "    Title: " + issue.Title + "\n")
		// write.WriteString("No." + strconv.Itoa(i) + "    Title: " + issue.Title + "\n" +
		// "Body: " + issue.Body + "\n" +
		// "Url: " + issue.Url + "          UpdateTime: " + issue.UpdateTime.String() + "\n" + "\n")
	}
	write.Flush()
}

func WriteReposInfoFile(filePath string, repos []*string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	for _, repo := range repos {
		write.WriteString(*repo + "\n")
	}
	write.Flush()
}

func output(filePath string, words_tfidfs wordTfidfs) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	for i, issue := range words_tfidfs {
		b, err := regexp.MatchString("^([A-z]+)$", issue.word)
		if err != nil {
			fmt.Println(err)
		}
		if b && len(issue.word) > 4 {
			write.WriteString(fmt.Sprintf("%-10s %-20s %-20s\n", strconv.Itoa(i), issue.word, fmt.Sprintf("%.6f", issue.frequency)))
		}
		// fmt.Sprintf("No.%-20v" + strconv.Itoa(i) + "%-20v" + issue.word + "%20v" + fmt.Sprintf("%.6f", issue.frequency) + "\n")
		// write.WriteString(fmt.Sprintf("No.%-20s" + strconv.Itoa(i) + "%-20s" + issue.word + "%20s" + fmt.Sprintf("%.6f", issue.frequency) + "\n"))

	}
	write.Flush()
}

// write name and owner of repos which contain aimed package
func WriteRepoContainAimedPackages(filePath string, repos []string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	for i, issue := range repos {
		write.WriteString("No." + strconv.Itoa(i) + "	" + issue + "\n")
		// write.WriteString("No." + strconv.Itoa(i) + "    Title: " + issue.Title + "\n" +
		// "Body: " + issue.Body + "\n" +
		// "Url: " + issue.Url + "          UpdateTime: " + issue.UpdateTime.String() + "\n" + "\n")
	}
	write.Flush()
}

// write repos and its new dependencies
func WriteSpecificRepoAndNewDependencies(filePath string, repo string, dependencies []models.Dependency) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(repo + "\n")
	for _, dependency := range dependencies {
		if dependency.PackageName != "" {
			write.WriteString(dependency.PackageName + " " + dependency.MethodName + " " +
				dependency.Version + " " + strconv.Itoa(dependency.Number) + "\n")
		}
	}
	write.Flush()
}

// sort and write all new dependencies this time collected
func SortAndWriteNewDependencies(filePath string, new_dependencies_map map[string]models.Dependency) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Open file err:%v \n", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	dependencies_list := make([]models.Dependency, 0)
	for _, dependency := range new_dependencies_map {
		dependencies_list = append(dependencies_list, dependency)
	}
	dependencies_list = SortDependenciesByUsedNumber(dependencies_list)
	for _, dependency := range dependencies_list {
		if dependency.PackageName != "" {
			write.WriteString(dependency.PackageName + " " + dependency.MethodName + " " +
				dependency.Version + " " + strconv.Itoa(dependency.Number) + "\n")
		}
	}
	write.Flush()
}

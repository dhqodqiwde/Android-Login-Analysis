package util

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type WordCountBean struct {
	word  string
	count int
}

func NewWordCountBean(word string, count int) *WordCountBean {
	return &WordCountBean{word, count}
}

type WordCountBeanList []*WordCountBean

func (list WordCountBeanList) Len() int {
	return len(list)
}

func (list WordCountBeanList) Less(i, j int) bool {
	if list[i].count > list[j].count {
		return true
	} else if list[i].count < list[j].count {
		return false
	} else {
		return list[i].word < list[j].word
	}
}

func (list WordCountBeanList) Swap(i, j int) {
	var temp *WordCountBean = list[i]
	list[i] = list[j]
	list[j] = temp
}

func (list WordCountBeanList) totalCount() int {
	totalCount := 0
	for _, v := range list {
		totalCount += v.count
	}

	return totalCount
}

func CountTestBase(inputFilePath string, outputFilePath string) {
	start := time.Now().UnixNano() / 1e6
	fileData, err := os.ReadFile(inputFilePath)
	CheckError(err, "read file")
	var fileText string = string(fileData)
	newRountineCount := runtime.NumCPU()*2 - 1
	runtime.GOMAXPROCS(newRountineCount + 1)
	parts := splitFileText(fileText, newRountineCount)

	var ch chan map[string]int = make(chan map[string]int, newRountineCount)
	for i := 0; i < newRountineCount; i++ {
		go countTest(parts[i], ch)
	}

	var totalWordsMap map[string]int = make(map[string]int, 0)
	completeCount := 0
	for {
		receiveData := <-ch
		for k, v := range receiveData {
			totalWordsMap[strings.ToLower(k)] += v
		}
		completeCount++

		if newRountineCount == completeCount {
			break
		}
	}

	list := make(WordCountBeanList, 0)
	for k, v := range totalWordsMap {
		list = append(list, NewWordCountBean(k, v))
	}
	sort.Sort(list)

	end := time.Now().UnixNano() / 1e6
	fmt.Printf("time consume:%dms\n", end-start)

	wordsCount := list.totalCount()
	var data bytes.Buffer
	data.WriteString(fmt.Sprintf("Cost Time：%dms\n", end-start))
	data.WriteString(fmt.Sprintf("Total words：%d\n\n", wordsCount))
	for _, v := range list {
		b, _ := regexp.MatchString("^([A-z]+)$", v.word)
		if len(v.word) > 4 && b {
			var percent float64 = 100.0 * float64(v.count) / float64(wordsCount)
			_, err := data.WriteString(fmt.Sprintf("%-10s %-5d  %3.2f%%\n", v.word, v.count, percent))
			CheckError(err, "bytes.Buffer, WriteString")
		}
	}

	err = os.WriteFile(outputFilePath, []byte(data.String()), os.ModePerm)
	CheckError(err, "ioutil.WriteFile")
}

func countTest(text string, ch chan map[string]int) {
	var wordMap map[string]int = make(map[string]int, 0)

	startIndex := 0
	letterStart := false
	for i, v := range text {
		if (v >= 65 && v <= 90) || (v >= 97 && v <= 122) {
			if !letterStart {
				letterStart = true
				startIndex = i
			}
		} else {
			if letterStart {
				wordMap[text[startIndex:i]]++
				letterStart = false
			}
		}
	}
	if letterStart {
		wordMap[text[startIndex:]]++
	}
	ch <- wordMap
}

func splitFileText(fileText string, n int) []string {
	length := len(fileText)
	parts := make([]string, n)

	lastPostion := 0
	for i := 0; i < n-1; i++ {
		position := length / n * (i + 1)
		for string(fileText[position]) != " " {
			position++
		}

		parts[i] = fileText[lastPostion:position]
		lastPostion = position
	}

	parts[n-1] = fileText[lastPostion:]
	return parts
}

func CheckError(err error, msg string) {
	if err != nil {
		panic(msg + "," + err.Error())
	}
}

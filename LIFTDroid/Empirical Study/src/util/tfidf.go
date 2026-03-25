package util

import (
	"math"
	"sort"
)

type TfidfUtil struct{}

type wordTfidf struct {
	word      string
	frequency float64
}

type wordTfidfs []wordTfidf

type Interface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

func (wts wordTfidfs) Len() int {
	return len(wts)
}
func (wts wordTfidfs) Less(i, j int) bool {
	return wts[i].frequency > wts[j].frequency
}
func (wts wordTfidfs) Swap(i, j int) {
	wts[i], wts[j] = wts[j], wts[i]
}

func (wts *wordTfidfs) Sort() {
	sort.Sort(wts)
}


func (tu *TfidfUtil) Fit(listWords [][]string, file_path string) {
	docFrequency := make(map[string]float64, 0)
	sumWorlds := 0
	for _, wordList := range listWords {
		for _, v := range wordList {
			docFrequency[v] += 1
			sumWorlds++
		}
	}
	wordTf := make(map[string]float64)
	for k, _ := range docFrequency {
		wordTf[k] = docFrequency[k] / float64(sumWorlds)
	}
	docNum := float64(len(listWords))
	wordIdf := make(map[string]float64)
	wordDoc := make(map[string]float64, 0)
	for k, _ := range docFrequency {
		for _, v := range listWords {
			for _, vs := range v {
				if k == vs {
					wordDoc[k] += 1
					break
				}
			}
		}
	}
	for k, _ := range docFrequency {
		wordIdf[k] = math.Log(docNum / (wordDoc[k] + 1))
	}
	var words wordTfidfs
	for k, _ := range docFrequency {
		words = append(words, wordTfidf{
			word:      k,
			frequency: wordTf[k] * wordIdf[k],
		})
	}
	words.Sort()
	output(file_path, words)
}

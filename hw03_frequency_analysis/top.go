package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

type frequencyInf struct {
	freq int
	text string
}

var dashWorld = regexp.MustCompile("--+")

func Top10(src string) []string {
	freqMap := make(map[string]int)

	for _, word := range strings.Fields(src) {
		word = strings.ToLower(word)
		if !dashWorld.MatchString(word) {
			word = strings.Trim(word, `,-!;.'"`)
			if len(word) == 0 {
				continue
			}
		}

		curVal := freqMap[word] + 1
		freqMap[word] = curVal
	}

	frequencyList := make([]frequencyInf, 0, len(freqMap))
	for k, v := range freqMap {
		frequencyList = append(frequencyList, frequencyInf{
			freq: v,
			text: k,
		})
	}

	sort.Slice(frequencyList, func(i, j int) bool {
		return frequencyList[i].freq > frequencyList[j].freq ||
			frequencyList[i].freq == frequencyList[j].freq && strings.Compare(frequencyList[i].text, frequencyList[j].text) < 0
	})

	var top []string
	topSize := 10
	if topSize > len(frequencyList) {
		topSize = len(frequencyList)
	}
	for i := 0; i < topSize; i++ {
		top = append(top, frequencyList[i].text)
	}

	return top
}

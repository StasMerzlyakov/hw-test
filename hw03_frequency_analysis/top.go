package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

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
		freqMap[word]++
	}

	wordsSlice := make([]string, 0, len(freqMap))
	for k := range freqMap {
		wordsSlice = append(wordsSlice, k)
	}

	sort.Slice(wordsSlice, func(i, j int) bool {
		a, b := wordsSlice[i], wordsSlice[j]
		return freqMap[a] > freqMap[b] ||
			freqMap[a] == freqMap[b] && strings.Compare(a, b) < 0
	})

	topSize := 10
	if topSize > len(wordsSlice) {
		topSize = len(wordsSlice)
	}

	return wordsSlice[:topSize]
}

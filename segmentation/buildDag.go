package segmentation

import (
	"os"
	"bufio"
	"fmt"
	"strings"
	"github.com/KKRainbow/segmentation-server/aca"
)

type DAGBuilder struct {
	matcher *aca.AhoCorasickMatcher
	phraseIndexMap map[string]int32
	maxStep int
}


func NewDAGBuilder(wordFile, phraseFile string, maxStep int) *DAGBuilder {
	dag := &DAGBuilder{
		phraseIndexMap: make(map[string]int32),
	}

	var wordFd, phraseFd *os.File
	var err error

	wordFd, err = os.Open(wordFile)
	if err != nil {
		panic(err)
	}
	defer wordFd.Close()

	phraseFd, err = os.Open(phraseFile)
	if err != nil {
		panic(err)
	}
	defer phraseFd.Close()

	wordReader := bufio.NewScanner(wordFd)
	phraseReader := bufio.NewScanner(phraseFd)

	phrases := make([]string, 0)

	line_num := int32(0)

	dag.phraseIndexMap["<bos>"] = int32(0)
	dag.phraseIndexMap["<eos>"] = int32(1)
	dag.phraseIndexMap["<unk>"] = int32(2)

	line_num = int32(3)

	readLineByLine := func(scanner *bufio.Scanner) {
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) != 0 && line[0] == ' ' {
				line = " "
			} else if  len(line) == 0 {
				line = "\n"
			} else {
				line = strings.TrimSpace(line)
			}
			if len([]rune(line)) > maxStep {
				continue
			}
			if _, ok := dag.phraseIndexMap[line]; ok {
				continue
			}
			dag.phraseIndexMap[line] = line_num
			phrases = append(phrases, line)
			line_num++
		}
	}

	readLineByLine(wordReader)
	fmt.Println("Word size:", len(phrases))
	readLineByLine(phraseReader)
	fmt.Println("Total size:", len(phrases))

	dag.matcher = aca.NewAhoCorasickMatcher()
	dag.matcher.Build(phrases)


	dag.maxStep = maxStep
	return dag
}

func (dag *DAGBuilder) buildMatrix(str []rune, maxLength int, forward bool) [][]int32 {
	matrix := make([][]int32, maxLength)
	for i := range matrix {
		matrix[i] = make([]int32, dag.maxStep)
		for j := range matrix[i] {
			matrix[i][j] = -1
		}
	}

	// '<bos>'
	matrix[0][0] = dag.phraseIndexMap["<bos>"]

	matches, matchIdx := dag.matcher.MatchRunes(str)
	for i, word := range matches {
		idx := matchIdx[i]
		wordLen := len([]rune(word))
		if forward {
			matrix[idx + wordLen - 1 + 1][wordLen - 1] = dag.phraseIndexMap[word]
		} else {
			matrix[idx + 1][wordLen - 1] = dag.phraseIndexMap[word]
		}
	}

	unkIdx := dag.phraseIndexMap["<unk>"]
	for i := 1; i < len(str); i++ {
		if matrix[i][0] < 0 {
			matrix[i][0] = unkIdx
		}
	}

	eosIdx := dag.phraseIndexMap["<eos>"]
	for i := len(str) + 1; i < maxLength; i++ {
		matrix[i][0] = eosIdx
	}

	return matrix
}

package main

// package main

import (
	"bufio"
	seg "chtseg/subs"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type strInputs struct {
	s string
	//	b int
	//	l int
}

type structResult struct {
	OrigInput  string
	UnsymInput string
	//	segOutput  string
	Score    float64
	SegItems []string
	NumWords uint
	Guessed  map[string]float64
}

var (
	restrASCII     = "^[\x22-\x7e]+$"
	restrCJKSymbol = "[\uFF01-\uFF5E\u3000-\u303F\u0020\u0028\u0029\u003C\u003E\u007B\u007D\\[\\]]"
)

func (x strInputs) String() string {
	return fmt.Sprintf("輸入=%s, 長度=%d, 位元數=%d", x.s, len([]rune(x.s)), len(x.s))
}

func main() {
	defer seg.CloseDB()
	reader := bufio.NewScanner(os.Stdin)
	inp := strInputs{""}
	var i int
	//	var reASCII = regexp.MustCompile(restrASCII)
	var reCJKSymbol = regexp.MustCompile(restrCJKSymbol)
	var result seg.StructSegments
	var out structResult
	for {
		fmt.Println("Please type a line to be segmented:")
		reader.Scan()
		text := reader.Text()
		if len(text) == 0 {
			fmt.Println("bye")
			break
		}
		// text = reCJKSymbol.ReplaceAllString(text, " ")
		out.OrigInput = text
		out.UnsymInput = ""
		out.SegItems = make([]string, 0, len([]rune(text)))
		out.Score = 0.0
		out.NumWords = 0
		out.Guessed = make(map[string]float64)

		words := reCJKSymbol.Split(text, -1)
		//		inp.b = len(inp.s)
		for i = 0; i < len(words); i++ {
			inp.s = words[i]
			fmt.Println(inp)
			if len(words[i]) == 0 {
				continue
			}
			if len(words[i]) > 0 {

				if len(out.UnsymInput) > 0 {
					out.UnsymInput = seg.StringJoin(out.UnsymInput, " ", words[i])
				} else {
					out.UnsymInput = words[i]
				}

				result = seg.Segepoch(words[i])
				elements := strings.Fields(result.SegSentence)
				out.SegItems = append(out.SegItems, elements...)
				totalScore := out.Score*float64(out.NumWords) + result.Score
				out.NumWords += result.Wordcnt
				out.Score = totalScore / float64(out.NumWords)
				for k, v := range result.Guessedwords {
					if _, exist := out.Guessed[k]; !exist {
						out.Guessed[k] = v
					}
				}
				//	fmt.Printf("out=%#v", out)
			}

		}

		jret, _ := json.Marshal(out)
		fmt.Printf("斷詞結果：%s\n", string(jret))

	}
}

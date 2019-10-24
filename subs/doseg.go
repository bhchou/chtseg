package seg

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

/*
current 目前位子的字
segsentence 斷完的句子，用TAB切
score 目前句子的權重
wordcnt 斷詞數
guessedwords 斷出來的詞中有含在詞庫者，加上這個詞的出現頻率
guesscnt 參考到詞庫的詞數
*/

type StructSegments struct {
	Current      string //scanned word
	SegSentence  string //A\tBC\tD\tEF
	Score        float64
	Wordcnt      uint
	Guessedwords map[string]float64
}

// word: 詞庫的詞 freq: 詞的出現頻率
// 用來暫放找到的詞與詞頻
// 好像用maps比較好
type StructGuessword struct {
	Word string
	Freq uint
}

// instead using the following map for a try

var (
	restrAlphabet = "^[\x22-\x7e\u3000\uFF01-\uFF5E\u2018\u2019]+$"
	maxWordLen    = 5
	mapGuessword  map[string]float64
)

func Segepoch(a string) StructSegments {

	chs := []rune(a) //input sentence seperated by space
	/*	chpos := 0       // current referenced character
		wpos := 0        // position to save current state, for gSegment
		var c string */
	var (
		nextch      string
		freq        uint
		tobeGuessed string
		current     string
		// bestpos      int     // the best pos to combine
		// bestscore    float64 // temperary best score
		// chpos        = 0
		wpos = 0
		//		prevpos      = 0
		segments []StructSegments
	)
	mapGuessword = make(map[string]float64)

	reASCII := regexp.MustCompile(restrAlphabet)
	InfoOut(fmt.Sprintf("quoted string: %+q\n", a))
	if reASCII.MatchString(a) { //純英數字
		return neednotSeg(a)
	}

	segments = make([]StructSegments, 1)
	var currentEpoch StructSegments

	for chndx := 0; chndx < len(chs); chndx++ {

		/* 1. find current chinese character or alphabets */

		current = string([]rune{chs[chndx]})
		if reASCII.MatchString(current) { //代表是英數字，必須全部抓完算一個字
			for chndx < len(chs)-1 {
				nextch = string([]rune{chs[chndx+1]})
				if reASCII.MatchString(nextch) {
					current = StringJoin(current, nextch)
					chndx++
					continue
				} else {
					break
				}
			}
		}
		InfoOut("Parse: ", current)
		currentEpoch.Current = current
		currentEpoch.Guessedwords = make(map[string]float64)
		currentEpoch.Score = 0.0

		/* 2. check DB for every possible new concated strings */

		tobeGuessed = current
		for i := 0; i < 7 && i <= wpos; i++ { //from current back to previous 7 stats if any
			InfoOut("wpos=", wpos, "i=", i)
			/*prevpos = wpos - i - 1
			if prevpos < 0 {
				break
			} */
			if i > 0 {
				tobeGuessed = StringJoin(segments[wpos-i].Current, tobeGuessed)
			}
			InfoOut("Guess:", tobeGuessed)
			if _, exist := mapGuessword[tobeGuessed]; !exist {
				errSelect := gdb.QueryRow("select freq from guess_words where guess = ?", fnMysqlRealEscapeString(tobeGuessed)).Scan(&freq)
				// fmt.Println("Select Error=", errSelect)
				if errSelect == nil { //Guessed
					mapGuessword[tobeGuessed] = math.Log10(float64(freq))*float64(i+1) + 1.0 //i+1 才是正確的“字“數
					InfoOut("finding", tobeGuessed, "got freq=", freq, "SCORE =", mapGuessword[tobeGuessed])
				} else if i == 0 {
					mapGuessword[tobeGuessed] = 0.1
				} else {
					mapGuessword[tobeGuessed] = 0.0
				}
			}
			//		fmt.Println("map", mapGuessword)
			if wpos == 0 {
				currentEpoch.Score = mapGuessword[tobeGuessed]
				currentEpoch.Wordcnt = 1
				currentEpoch.SegSentence = tobeGuessed
				if mapGuessword[tobeGuessed] > 1.0 { //有找到
					currentEpoch.Guessedwords[tobeGuessed] = mapGuessword[tobeGuessed]
				}
				//	segments[wpos] = currentEpoch
			} else {
				if i == 0 {
					prevEpoch := segments[wpos-i-1]
					// currentEpoch.Score = (prevEpoch.Score*float64(prevEpoch.Wordcnt) + mapGuessword[tobeGuessed]) / float64(prevEpoch.Wordcnt+1)
					/* try to degrade wordcount effect */
					currentEpoch.Score = (prevEpoch.Score*math.Sqrt(float64(prevEpoch.Wordcnt)) + mapGuessword[tobeGuessed]) / math.Sqrt(float64(prevEpoch.Wordcnt+1))
					currentEpoch.Wordcnt = prevEpoch.Wordcnt + 1
					currentEpoch.SegSentence = StringJoin(prevEpoch.SegSentence, "\t", tobeGuessed)
					//////// copy (currentEpoch.Guessedwords, prevEpoch.Guessedwords)
					for k, v := range prevEpoch.Guessedwords {
						currentEpoch.Guessedwords[k] = v
					}
					if mapGuessword[tobeGuessed] > 1.0 { //有找到
						currentEpoch.Guessedwords[tobeGuessed] = mapGuessword[tobeGuessed]
					}
					//	segments = append(segments, currentEpoch)
				} else if wpos == i { // i 不是 0 但也沒有前面的了，代表 tobeguess 是一個字以上
					tempScore := mapGuessword[tobeGuessed]
					if tempScore > currentEpoch.Score {
						currentEpoch.Wordcnt = 1
						currentEpoch.Score = tempScore
						currentEpoch.SegSentence = tobeGuessed
						currentEpoch.Guessedwords = make(map[string]float64)
						if mapGuessword[tobeGuessed] > 1.0 { //有找到
							currentEpoch.Guessedwords[tobeGuessed] = mapGuessword[tobeGuessed]
						}
						//	segments = append(segments, currentEpoch)
					}
				} else {
					prevEpoch := segments[wpos-i-1]
					// tempScore := (prevEpoch.Score*float64(prevEpoch.Wordcnt) + mapGuessword[tobeGuessed]) / float64(prevEpoch.Wordcnt+1)
					/* try to degrade wordcount effect */
					tempScore := (prevEpoch.Score*math.Sqrt(float64(prevEpoch.Wordcnt)) + mapGuessword[tobeGuessed]) / math.Sqrt(float64(prevEpoch.Wordcnt+1))
					InfoOut(fmt.Sprintf("P=%#v, to=%s, m=%#v\n", prevEpoch, tobeGuessed, mapGuessword[tobeGuessed]))
					if tempScore > currentEpoch.Score && mapGuessword[tobeGuessed] > 0 { //查找的詞必須要有找到或是一個字才會去比
						currentEpoch.Wordcnt = prevEpoch.Wordcnt + 1
						currentEpoch.Score = tempScore
						currentEpoch.SegSentence = StringJoin(prevEpoch.SegSentence, "\t", tobeGuessed)
						currentEpoch.Guessedwords = make(map[string]float64)
						//////// copy (currentEpoch.Guessedwords, prevEpoch.Guessedwords)
						for k, v := range prevEpoch.Guessedwords {
							currentEpoch.Guessedwords[k] = v
						}
						if mapGuessword[tobeGuessed] > 1.0 { //有找到
							currentEpoch.Guessedwords[tobeGuessed] = mapGuessword[tobeGuessed]
						}
						//	segments = append(segments, currentEpoch)
					}
				}
			}
			InfoOut(fmt.Sprintf("CURRENT = %#v\n", currentEpoch))
		}
		if wpos == 0 {
			segments[0] = currentEpoch
		} else {
			segments = append(segments, currentEpoch)
		}
		InfoOut(fmt.Sprintf("SEGMENT = %#v\n", segments))
		wpos++
	}
	return segments[wpos-1]
}

func neednotSeg(a string) StructSegments {

	var currentEpoch StructSegments
	var freq uint

	currentEpoch.Current = a
	currentEpoch.Guessedwords = make(map[string]float64)
	currentEpoch.Score = 0.0

	if _, exist := mapGuessword[a]; !exist {
		errSelect := gdb.QueryRow("select freq from guess_words where guess = ?", fnMysqlRealEscapeString(a)).Scan(&freq)
		// fmt.Println("Select Error=", errSelect)
		if errSelect == nil { //Guessed
			mapGuessword[a] = math.Log10(float64(freq)) + 1.0
			//			fmt.Println("finding", a, "got freq=", freq, "SCORE =", mapGuessword[a])
		} else {
			mapGuessword[a] = 1.0
		}
	}
	currentEpoch.Score = mapGuessword[a]
	currentEpoch.Wordcnt = 1
	currentEpoch.SegSentence = a
	if mapGuessword[a] > 1.0 {
		currentEpoch.Guessedwords[a] = mapGuessword[a]
	}

	return currentEpoch
}

func StringJoin(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

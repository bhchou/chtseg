package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"C"

	seg "chtseg/subs"
)

type structResult struct {
	OrigInput  string
	UnsymInput string
	Score    float64
	SegItems []string
	NumWords uint
	Guessed  map[string]float64
}

var (
	restrASCII     = "^[\x22-\x7e]+$"
	restrCJKSymbol = "[\uFF01-\uFF5E\u3000-\u303F\u0020\u0028\u0029\u003C\u003E\u007B\u007D\\[\\]]"
	batchfile     = ""
	mysqlconnstr  = ""
	sqliteconnstr = ""
	confconnstr   = ""
	connstr       = ""
	dbengine      = ""
	nohelp        = false
	mtx sync.Mutex
)

//export Getchtseg
func Getchtseg(j_dbtype *C.char, j_dbconn *C.char, j_teststr *C.char) (ret_json, ret_error *C.char) {
	mtx.Lock()
	defer mtx.Unlock()
	dbtype := C.GoString(j_dbtype)
	dbconn := C.GoString(j_dbconn)
	teststr := C.GoString(j_teststr)
	
	rjson := ""
	rerror := ""
	seg.InitDB(dbtype, dbconn, false)
	if seg.Err != nil {
		fmt.Printf("There is something wrong: %s\n", seg.Err)
		seg.CloseDB()
		//		os.Exit(1)
		dir, _ := os.Getwd()
		rerror = fmt.Sprintf("DBTYPE=%s, Connection=%s, Current dir=%s, Error=%s", dbtype, dbconn, dir, seg.Err.Error())
		return C.CString(rjson), C.CString(rerror)
	}

	defer seg.CloseDB()
	out := doSeg(teststr)
	jret, _ := json.Marshal(out)
	rjson = string(jret)
	return C.CString(rjson), C.CString(rerror)
}

func doSeg(text string) structResult {
	var out structResult
	var reCJKSymbol = regexp.MustCompile(restrCJKSymbol)
	var result seg.StructSegments

	out.OrigInput = text
	out.UnsymInput = ""
	out.SegItems = make([]string, 0, len([]rune(text)))
	out.Score = 0.0
	out.NumWords = 0
	out.Guessed = make(map[string]float64)

	words := reCJKSymbol.Split(text, -1)
	for i := 0; i < len(words); i++ {
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
		}

	}
	return out
}

// should not run this function
func main() { panic("Error if you see this line") }

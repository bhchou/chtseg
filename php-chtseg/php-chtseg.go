package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kitech/php-go/phpgo"

	seg "chtseg/subs"
)

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
	//	app            = cli.NewApp()
	//	flags          []cli.Flag
	batchfile     = ""
	mysqlconnstr  = ""
	sqliteconnstr = ""
	confconnstr   = ""
	connstr       = ""
	dbengine      = ""
	nohelp        = false

//	cf             cnf.Configurations
)

type PHPseg struct {
}

func NewPHPseg() *PHPseg {
	//	log.Println("NewPGDemo...")
	return &PHPseg{}
}

func (this *PHPseg) Getchtseg(dbtype string, dbconn string, teststr string) map[string]string {
	ret := make(map[string]string)
	ret["json"] = ""
	ret["error"] = ""
	seg.InitDB(dbtype, dbconn, false)
	if seg.Err != nil {
		fmt.Printf("There is something wrong: %s\n", seg.Err)
		seg.CloseDB()
		//		os.Exit(1)
		dir, _ := os.Getwd()
		ret["error"] = fmt.Sprintf("DBTYPE=%s, Connection=%s, Current dir=%s, Error=%s", dbtype, dbconn, dir, seg.Err.Error())
		return ret
	}

	defer seg.CloseDB()
	out := doSeg(teststr)
	jret, _ := json.Marshal(out)
	ret["json"] = string(jret)
	return ret

}

func module_startup(ptype int, module_number int) int {
	println("module_startup", ptype, module_number)
	return rand.Int()
}
func module_shutdown(ptype int, module_number int) int {
	println("module_shutdown", ptype, module_number)
	return rand.Int()
}
func request_startup(ptype int, module_number int) int {
	println("request_startup", ptype, module_number)
	return rand.Int()
}
func request_shutdown(ptype int, module_number int) int {
	println("request_shutdown", ptype, module_number)
	return rand.Int()
}

func init() {
	log.Println("run us init...")
	rand.Seed(time.Now().UnixNano())

	phpgo.InitExtension("pg0", "")
	phpgo.RegisterInitFunctions(module_startup, module_shutdown, request_startup, request_shutdown)

	
	if true {
		phpgo.AddClass("PHPseg", NewPHPseg)
	}
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

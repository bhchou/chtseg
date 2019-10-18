package main

// package main

import (
	"bufio"
	cnf "chtseg/config"
	seg "chtseg/subs"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
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
	app            = cli.NewApp()
	flags          []cli.Flag
	batchfile      = ""
	mysqlconnstr   = ""
	confconnstr    = ""
	nohelp         = false
	cf             cnf.Configurations
)

func (x strInputs) String() string {
	return fmt.Sprintf("輸入=%s, 長度=%d, 位元數=%d", x.s, len([]rune(x.s)), len(x.s))
}

func init() {
	flags = []cli.Flag{
		cli.StringFlag{
			Name:        "with-batchfile, b",
			Value:       "",
			Usage:       "input file from `BATCHFILE` for batch segmentation",
			Destination: &batchfile,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "print debug info",
		},
		cli.BoolFlag{
			Name:  "test, t",
			Usage: "Just test, no segmentation, will output info regardless -v",
		},
		cli.StringFlag{
			Name:  "with-mysql, m",
			Value: "",
			Usage: "mysql connect string in \"username:password@conntype(ip:port)/db\" format(\" quoted)",
			//			Required:    true,
			Destination: &mysqlconnstr,
		},
	}

	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable VIPER to read Environment Variables
	//viper.AutomaticEnv()

	viper.SetConfigType("yml")

	err1 := viper.ReadInConfig()
	if err1 == nil {
		err := viper.Unmarshal(&cf)
		if err == nil {
			confconnstr = fmt.Sprintf("%s:%s@%s(%s:%d)/%s", cf.Database.DBUser, cf.Database.DBPassword, cf.Server.Type, cf.Server.IP, cf.Server.Port, cf.Database.DBName)
		}
	}

}

func cliInfo() {
	app.Name = "ChtSeg"
	app.Usage = "Chinese word segmentation"
	app.Author = "Jack Chou"
	// app.Version = "0.1.0"
	app.HideVersion = true
	app.Flags = flags
	app.Action = cliAction
}

func cliAction(c *cli.Context) error {
	if len(batchfile) > 0 {
		if !fileExists(batchfile) {
			return cli.NewExitError("batch file to be segmented does not exist", 5)
		} else {
			seg.Verbose = false
		}
	}
	if c.Bool("v") {
		seg.Verbose = true
	}
	if c.Bool("t") {
		seg.Verbose = true
		seg.Test = true
	}
	/* if len(mysqlconnstr) == 0 {
		return cli.NewExitError("db connect string not set", 1)
	}*/
	nohelp = true
	return nil
}

func main() {
	cliInfo()
	cliErr := app.Run(os.Args)
	if cliErr != nil {
		os.Exit(1)
	}
	if !nohelp {
		os.Exit(0)
	}
	if len(mysqlconnstr) == 0 {
		if len(confconnstr) == 0 {
			fmt.Println("The db connection info should be set in config.yml or --with-mysql in command line")
			os.Exit(1)
		}
		mysqlconnstr = confconnstr
	}
	seg.InitDB(mysqlconnstr)
	if seg.Err != nil {
		fmt.Printf("There is something wrong: %s\n", seg.Err)
		seg.CloseDB()
		os.Exit(1)
	}

	defer seg.CloseDB()

	reader := bufio.NewScanner(os.Stdin)
	//	inp := strInputs{""}
	//	var i int
	//	var reASCII = regexp.MustCompile(restrASCII)
	//	var reCJKSymbol = regexp.MustCompile(restrCJKSymbol)
	//	var result seg.StructSegments
	var out structResult
	var segResult string
	if len(batchfile) == 0 {
		for {
			fmt.Println("Please drop a line to be segmented, or just enter to quit:")
			reader.Scan()
			text := reader.Text()
			if len(text) == 0 {
				fmt.Println("bye")
				break
			}
			out = doSeg(text)
			if seg.Verbose {
				jret, _ := json.Marshal(out)
				fmt.Printf("---Result---\n%s\n", string(jret))
			} else {
				segResult = strings.Join(out.SegItems, "|")
				fmt.Printf("%s\n---Found Keywords---\n", segResult)
				for k, v := range out.Guessed {
					fmt.Printf("%s[%f]\n", k, v)
				}
			}
		}
	} else {
		outfile := seg.StringJoin(batchfile, ".out")
		if fileExists(outfile) {
			fmt.Println("File for output:", outfile, "exists, please figure it out.")
			seg.CloseDB()
			os.Exit(2)
		}
		ofile, ferr := os.Create(outfile)
		if ferr != nil {
			fmt.Println("Output file ", outfile, "cannot be created, error =", ferr)
			seg.CloseDB()
			os.Exit(2)
		}

		ifile, ferr := os.Open(batchfile)
		if ferr != nil {
			fmt.Println("Input file ", batchfile, "cannot be opened, error =", ferr)
			seg.CloseDB()
			os.Exit(2)
		}
		defer ifile.Close()
		defer ofile.Close()

		owfile := bufio.NewWriter(ofile)
		scanner := bufio.NewScanner(ifile)
		for i := 0; scanner.Scan(); i++ {
			text := scanner.Text()
			if len(text) == 0 {
				continue
			}
			out = doSeg(text)
			segResult = strings.Join(out.SegItems, "|")
			fmt.Fprintf(owfile, "%s\n%s\n---\n%#v\n===\n", text, segResult, out.Guessed)
			fmt.Printf("\r%d", i)
			if i%100 == 0 {
				owfile.Flush()
			}
		}
		owfile.Flush()
		fmt.Println()
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func doSeg(text string) structResult {
	// text = reCJKSymbol.ReplaceAllString(text, " ")
	var out structResult
	var reCJKSymbol = regexp.MustCompile(restrCJKSymbol)
	var result seg.StructSegments

	out.OrigInput = text
	out.UnsymInput = ""
	out.SegItems = make([]string, 0, len([]rune(text)))
	out.Score = 0.0
	out.NumWords = 0
	out.Guessed = make(map[string]float64)

	//	fmt.Println("doseg:", text)
	words := reCJKSymbol.Split(text, -1)
	//		inp.b = len(inp.s)
	for i := 0; i < len(words); i++ {
		//		inp.s = words[i]
		//		fmt.Println(inp)
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
	return out
	//	jret, _ := json.Marshal(out)
	//	fmt.Printf("斷詞結果：%s\n", string(jret))
}

package main

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
}

type structResult struct {
	OrigInput  string
	UnsymInput string
	Score      float64
	SegItems   []string
	NumWords   uint
	Guessed    map[string]float64
}

var (
	restrASCII     = "^[\x22-\x7e]+$"
	restrCJKSymbol = "[\uFF01-\uFF5E\u3000-\u303F\u0020\u0028\u0029\u003C\u003E\u007B\u007D\\[\\]]"
	app            = cli.NewApp()
	flags          []cli.Flag
	batchfile      = ""
	mysqlconnstr   = ""
	sqliteconnstr  = ""
	confconnstr    = ""
	connstr        = ""
	dbengine       = ""
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
			Name:        "with-mysql, m",
			Value:       "",
			Usage:       "mysql connect string in \"username:password@conntype(ip:port)/db\" format(\" quoted)",
			Destination: &mysqlconnstr,
		},
		cli.StringFlag{
			Name:        "with-sqlite3, q",
			Value:       "",
			Usage:       "sqlite3 connect string (db file path)",
			Destination: &sqliteconnstr,
		},
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")

	err1 := viper.ReadInConfig()
	if err1 == nil {
		err := viper.Unmarshal(&cf)
		if err == nil {
			if strings.TrimSpace(strings.ToLower(cf.Server.Engine)) == "mysql" {
				confconnstr = fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
					strings.TrimSpace(cf.Database.DBUser),
					strings.TrimSpace(cf.Database.DBPassword),
					strings.TrimSpace(cf.Server.Type),
					strings.TrimSpace(cf.Server.IP),
					cf.Server.Port,
					strings.TrimSpace(cf.Database.DBName))
			} else if strings.TrimSpace(strings.ToLower(cf.Server.Engine)) == "sqlite3" {
				confconnstr = strings.TrimSpace(cf.Server.DBFile)
			} else {
				confconnstr = ""
			}
		}
	}

}

func cliInfo() {
	app.Name = "ChtSeg"
	app.Usage = "Chinese word segmentation"
	app.Author = "Jack Chou"
	app.Version = "1.0.0"
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
	if len(mysqlconnstr) > 0 && len(sqliteconnstr) > 0 {
		return cli.NewExitError("I am confused on which db engine will be used", 1)
	}
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

	if len(mysqlconnstr) > 0 {
		connstr = mysqlconnstr
		dbengine = "mysql"
	} else if len(sqliteconnstr) > 0 {
		connstr = sqliteconnstr
		dbengine = "sqlite3"
	} else {
		connstr = confconnstr
		dbengine = cf.Server.Engine
	}
	if len(connstr) == 0 || !(dbengine == "mysql" || dbengine == "sqlite3") {
		fmt.Println("The db info should be set correctly in config.yml or --with-mysql/--with-sqlite3 in command line")
		os.Exit(1)
	}

	seg.InitDB(dbengine, connstr, false)
	if seg.Err != nil {
		fmt.Printf("There is something wrong: %s\n", seg.Err)
		seg.CloseDB()
		os.Exit(1)
	}

	defer seg.CloseDB()

	reader := bufio.NewScanner(os.Stdin)
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

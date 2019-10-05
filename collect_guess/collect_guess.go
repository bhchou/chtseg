package main

import (
	"bufio"
	seg "chtseg/subs"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/urfave/cli"
)

var (
	app                   = cli.NewApp()
	flags                 []cli.Flag
	csvfile, stopwordfile string
	giLines               = 0
	restrEngCJK           = "^([\x41-\x5A\x61-\x7A]*)([\u3400-\u4DBF\u4E00-\u9FFF]{1,7})$"
)

func init() {
	flags = []cli.Flag{
		cli.StringFlag{
			Name:        "with-csvfile, c",
			Value:       "",
			Usage:       "input file from `CSVFILE`",
			Required:    true,
			Destination: &csvfile,
		},
		cli.StringFlag{
			Name:        "with-stopword-file, s",
			Value:       "",
			Usage:       "optional stopword file",
			Destination: &stopwordfile,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "print debug info",
		},
		cli.BoolFlag{
			Name:  "test, t",
			Usage: "Just test, no write to DB, will output info regradless -t",
		},
	}
}

func cliInfo() {
	app.Name = "Words Collecting"
	app.Usage = "Collecting global known words into mysql DB"
	app.Author = "Jack Chou"
	// app.Version = "0.1.0"
	app.HideVersion = true
	app.Flags = flags
	app.Action = cliAction
}

func cliAction(c *cli.Context) error {
	/* csvfile = ""
	if c.NArg() > 0 {
		csvfile = c.Args().Get(0)
		if len(csvfile) < 0 {
			cli.ShowAppHelp(c)
			return cli.NewExitError("please set csv file", 4)
		} else {
			if fileExists(csvfile) {
				fmt.Println("EC", ec, csvfile)
			} else {
				return cli.NewExitError("csv file does not exist", 5)
			}
		}
	} else {
		cli.ShowAppHelp(c)
		return cli.NewExitError("please set csv file", 4)
	} */
	if c.NumFlags() < 1 {
		cli.ShowAppHelp(c)
		return cli.NewExitError("please indicate csvfile", 2)
	}
	if len(csvfile) < 0 {
		cli.ShowAppHelp(c)
		return cli.NewExitError("please set csv file", 4)
	} else {
		if !fileExists(csvfile) {
			return cli.NewExitError("csv file does not exist", 5)
		}
	}
	if len(stopwordfile) > 0 && (!fileExists(stopwordfile)) {
		return cli.NewExitError("stopword file does not exist", 5)
	}
	if c.Bool("v") {
		seg.Verbose = true
	}
	if c.Bool("t") {
		seg.Verbose = true
		seg.Test = true
	}

	/*		switch ec {
			case "pchome", "shopee", "ruten":
			default:
				cli.ShowAppHelp(c)
				return cli.NewExitError("platform must be pchome/shopee/ruten", 3)
			} */

	return nil
}

func main() {
	cliInfo()
	cliErr := app.Run(os.Args)
	if cliErr != nil {
		os.Exit(1)
	}
	seg.InfoOut("get", csvfile)

	if seg.Err != nil {
		defer seg.CloseDB()
	} else {
		log.Fatal(seg.Err)
		os.Exit(2)
	}

	file, err := os.Open(csvfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if len(stopwordfile) > 0 {
		sf, err := ioutil.ReadFile(stopwordfile)
		if err != nil {
			log.Fatal(err)
		}
		seg.Stopwords = string(sf)
	}

	scanner := bufio.NewScanner(file)
	//	inp := strInputs{""}
	//	ignore := strInputs{""}
	re := regexp.MustCompile("[\\s+\\t+]")
	reSymbol := regexp.MustCompile("[\u3000-\u303F]+")
	reEngCjk := regexp.MustCompile(restrEngCJK)
	for scanner.Scan() {
		seg.InfoOut(scanner.Text(), "===>")
		text := scanner.Text()
		text1 := reSymbol.ReplaceAllString(text, " ")
		words := re.Split(text1, -1)
		//		inp.b = len(inp.s)
		for i := 1; i < len(words); i++ {
			if len(words[i]) > 0 {
				rex := reEngCjk.FindStringSubmatch(words[i])
				if rex != nil {
					seg.ProcessWord(rex[1])
					seg.ProcessWord(rex[2])
					giLines += 2
				} else {
					seg.ProcessWord(words[i])
					giLines++
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("CSVFILE scan error:", err)
	}

	fmt.Println("處理詞數：", giLines)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

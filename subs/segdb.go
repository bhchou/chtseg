package seg

import (
	"database/sql"
	"fmt"
	"log"

	// 	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type strInputs struct {
	s string
}

type strIgnore struct {
	s string
}

func (x strInputs) String() string {
	return fmt.Sprintf("輸入=%s, 長度=%d, 位元數=%d", x.s, len([]rune(x.s)), len(x.s))
}

func (x strIgnore) String() string {
	return fmt.Sprintf("不留=%s, 長度=%d, 位元數=%d", x.s, len([]rune(x.s)), len(x.s))
}

var (
	State                        uint
	Err                          error
	gdb                          *sql.DB
	gstmtInsWord, gstmtInsIgnore *sql.Stmt
	gstmtUpdWord, gstmtUpdIgnore *sql.Stmt
	gstrConnect                  = ""
	// "pi:0955raspberry@tcp(127.0.0.1:3306)/cloud"
	Verbose = false
	Test    = false
)

func InitDB(strEngine string, strConnect string, doPrepare bool) {
	gstrConnect = strConnect
	State, Err = prepareDB(strEngine, doPrepare)
	if Err == nil {
		fmt.Println(strEngine, "DB connected")
	} else {
		fmt.Println(strEngine, "DB connect state＝", State, ", error=", Err)
	}
}

func CloseDB() {
	if State&31 == 31 {
		gstmtInsWord.Close()
		gstmtInsIgnore.Close()
		gstmtUpdWord.Close()
		gstmtUpdIgnore.Close()
		gdb.Close()
	} else if State == 1 {
		gdb.Close()
	}
}

func prepareDB(strEngine string, doPrepare bool) (prepareState uint, dbErr error) {

	prepareState = 0 /* 00000 for 4 statements and db itself */

	gdb, dbErr = sql.Open(strEngine, gstrConnect)
	if dbErr != nil {
		return prepareState, dbErr
	}
	prepareState |= 1
	if !doPrepare {
		return prepareState, dbErr
	}

	gstmtInsWord, dbErr = gdb.Prepare("INSERT INTO guess_words(prefix,guess,freq) VALUES( ?, ?, ? )")
	if dbErr != nil {
		return prepareState, dbErr
	}
	prepareState |= 2

	gstmtInsIgnore, dbErr = gdb.Prepare("INSERT INTO ignore_words(prefix,guess,freq) VALUES( ?, ?, ? )")
	if dbErr != nil {
		return prepareState, dbErr
	}
	prepareState |= 4

	gstmtUpdWord, dbErr = gdb.Prepare("UPDATE guess_words set freq = freq+1 where guess = ?")
	if dbErr != nil {
		return prepareState, dbErr
	}
	prepareState |= 8

	gstmtUpdIgnore, dbErr = gdb.Prepare("UPDATE ignore_words set freq = freq+1 where guess = ?")
	if dbErr != nil {
		return prepareState, dbErr
	}
	prepareState |= 16

	return prepareState, dbErr
}

func ProcessWord(w string) {
	if len(w) == 0 {
		return
	}
	var freq int
	prefix := []rune(w)
	if !IsIgnore(w) {
		inp := strInputs{w}
		InfoOut(inp)
		errSelect := gdb.QueryRow("select freq from guess_words where guess = ?", fnMysqlRealEscapeString(w)).Scan(&freq)
		if errSelect != nil && !Test {
			if errSelect == sql.ErrNoRows {
				gstmtInsWord.Exec(fnMysqlRealEscapeString(string(prefix[0])), fnMysqlRealEscapeString(w), 1)
			} else {
				log.Fatal("Query freq error: ", errSelect)
			}
		} else if !Test {
			gstmtUpdWord.Exec(fnMysqlRealEscapeString(w))
		}
	} else {
		inp := strIgnore{w}
		InfoOut(inp)
		errSelect := gdb.QueryRow("select freq from ignore_words where guess = ?", fnMysqlRealEscapeString(w)).Scan(&freq)
		if errSelect != nil && !Test {
			if errSelect == sql.ErrNoRows {
				gstmtInsIgnore.Exec(fnMysqlRealEscapeString(string(prefix[0])), fnMysqlRealEscapeString(w), 1)
			} else {
				log.Fatal("Query ignore freq error:", errSelect)
			}
		} else if !Test {
			gstmtUpdIgnore.Exec(fnMysqlRealEscapeString(w))
		}
	}
}

func fnMysqlRealEscapeString(value string) string {

	/*	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}

		for b, a := range replace {
			value = strings.Replace(value, b, a, -1)
		}
	*/
	return value
}

func InfoOut(a ...interface{}) (n int, err error) {
	if Verbose {
		n, err := fmt.Println(a...)
		return n, err
	}
	return 0, nil
}

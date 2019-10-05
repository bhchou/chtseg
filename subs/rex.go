package seg

import (
	"regexp"
	"strings"
)

var (
	// Stopwords : collection of stopwords
	Stopwords string
	/*the CJK symbols have been removed from the following*/
	restrASCII    = "^[0-9A-Za-z]+([^0-9A-Za-z][0-9A-Za-z]+)?$"
	restrChinese  = "^[\u3400-\u4DBF\u4E00-\u9FFF]{1,7}[0-9a-zA-Z.]*$" // 1-7個中文字＋後面可能接英數字
	restrChinese1 = "^[0-9]+[\x22-\x27\x2A-\x2F\x3A\x3B\x40\x5C\x7C\x7E]?[0-9]*[\u3400-\u4DBF\u4E00-\u9FFF]{2,7}$" //數字＋不是括弧引號＋數字＋2－7個中文字
	restrChinese2 = "^([\x41-\x5A\x61-\x7A]*)([\u3400-\u4DBF\u4E00-\u9FFF]{1,7})$" //英文字＋1－7個中文字
	restrChinese3 = "^[\u3400-\u4DBF\u4E00-\u9FFF]{1,7}([0-9a-zA-Z.]+[\u3400-\u4DBF\u4E00-\u9FFF]+)*$" //中文中間含英數
	restrEnglish  = "^[a-zA-Z]{2,}[0-9]*([^0-9A-Za-z][0-9A-Za-z]+)*[0-9]*$" //2英文字＋可能的數字＋可能的符號與英數的組合＋可能的數字
)

// IsIgnore : check if mixed Chinese stopword
func IsIgnore(stringForTest string) bool {
	var ( /*only for checking ignored word*/
		//	reAlphabet    = regexp.MustCompile("^[a-zA-z]$")
		reASCII = regexp.MustCompile(restrASCII)
		//	reLongChinese = regexp.MustCompile("^[\u3000-\u303F\u3400-\u4DBF\u4E00-\u9FFF]{8,}$")
		/*the following will be the most case to speedup checking*/
		reChinese  = regexp.MustCompile(restrChinese)
		reChinese1 = regexp.MustCompile(restrChinese1)
		reChinese2 = regexp.MustCompile(restrChinese2)
		reChinese3 = regexp.MustCompile(restrChinese3)
		reEnglish  = regexp.MustCompile(restrEnglish)
	)
	ret := true
	switch {
	case reChinese.MatchString(stringForTest):
		ret = false
	case reChinese1.MatchString(stringForTest):
		if len([]rune(stringForTest)) > 7 {
			ret = true
		} else {
			ret = false
		}
	case reChinese2.MatchString(stringForTest):
		ret = false
	case reChinese3.MatchString(stringForTest):
		if len([]rune(stringForTest)) > 7 {
			ret = true
		} else {
			ret = false
		}
	case reEnglish.MatchString(stringForTest):
		if len(Stopwords) > 0 {
			ret, _ = regexp.MatchString("\\b"+stringForTest+"\\b", Stopwords)
			/*			ret = true
						} else {
							ret = false */
		}
	case reASCII.MatchString(stringForTest):
		if strings.Compare(strings.ToLower(stringForTest), "3m") == 0 || strings.Compare(strings.ToLower(stringForTest), "7-11") == 0 {
			ret = false
		} else {
			ret = true
		}
	}
	return ret
}

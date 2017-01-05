package enmime

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"errors"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

/* copy from golang.org/x/net/html/charset/table.go */
var encodings = map[string]struct {
	e    encoding.Encoding
	name string
}{
	"utf-8":          {encoding.Nop, "utf-8"},
	"ibm866":         {charmap.CodePage866, "ibm866"},
	"iso-8859-2":     {charmap.ISO8859_2, "iso-8859-2"},
	"iso-8859-3":     {charmap.ISO8859_3, "iso-8859-3"},
	"iso-8859-4":     {charmap.ISO8859_4, "iso-8859-4"},
	"iso-8859-5":     {charmap.ISO8859_5, "iso-8859-5"},
	"iso-8859-6":     {charmap.ISO8859_6, "iso-8859-6"},
	"iso-8859-7":     {charmap.ISO8859_7, "iso-8859-7"},
	"iso-8859-8":     {charmap.ISO8859_8, "iso-8859-8"},
	"iso-8859-8-i":   {charmap.ISO8859_8, "iso-8859-8-i"},
	"iso-8859-10":    {charmap.ISO8859_10, "iso-8859-10"},
	"iso-8859-13":    {charmap.ISO8859_13, "iso-8859-13"},
	"iso-8859-14":    {charmap.ISO8859_14, "iso-8859-14"},
	"iso-8859-15":    {charmap.ISO8859_15, "iso-8859-15"},
	"iso-8859-16":    {charmap.ISO8859_16, "iso-8859-16"},
	"koi8-r":         {charmap.KOI8R, "koi8-r"},
	"koi8-u":         {charmap.KOI8U, "koi8-u"},
	"macintosh":      {charmap.Macintosh, "macintosh"},
	"windows-874":    {charmap.Windows874, "windows-874"},
	"windows-1250":   {charmap.Windows1250, "windows-1250"},
	"windows-1251":   {charmap.Windows1251, "windows-1251"},
	"windows-1252":   {charmap.Windows1252, "windows-1252"},
	"windows-1253":   {charmap.Windows1253, "windows-1253"},
	"windows-1254":   {charmap.Windows1254, "windows-1254"},
	"windows-1255":   {charmap.Windows1255, "windows-1255"},
	"windows-1256":   {charmap.Windows1256, "windows-1256"},
	"windows-1257":   {charmap.Windows1257, "windows-1257"},
	"windows-1258":   {charmap.Windows1258, "windows-1258"},
	"x-mac-cyrillic": {charmap.MacintoshCyrillic, "x-mac-cyrillic"},
	"gbk":            {simplifiedchinese.GBK, "gbk"},
	"gb18030":        {simplifiedchinese.GB18030, "gb18030"},
	"hz-gb-2312":     {simplifiedchinese.HZGB2312, "hz-gb-2312"},
	"big5":           {traditionalchinese.Big5, "big5"},
	"euc-jp":         {japanese.EUCJP, "euc-jp"},
	"iso-2022-jp":    {japanese.ISO2022JP, "iso-2022-jp"},
	"shift_jis":      {japanese.ShiftJIS, "shift_jis"},
	"euc-kr":         {korean.EUCKR, "euc-kr"},
	"replacement":    {encoding.Replacement, "replacement"},
	"utf-16be":       {unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "utf-16be"},
	"utf-16le":       {unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le"},
	"x-user-defined": {charmap.XUserDefined, "x-user-defined"},
	"cp850":          {charmap.CodePage850, "cp850"},
}

var encodingAliases = map[string]string{
	"unicode-1-1-utf-8": "utf-8",
	"utf-8":             "utf-8",
	"utf8":              "utf-8",

	"866":      "ibm866",
	"cp866":    "ibm866",
	"csibm866": "ibm866",
	"ibm866":   "ibm866",

	"csisolatin2":     "iso-8859-2",
	"iso-8859-2":      "iso-8859-2",
	"iso-ir-101":      "iso-8859-2",
	"iso8859-2":       "iso-8859-2",
	"iso88592":        "iso-8859-2",
	"iso_8859-2":      "iso-8859-2",
	"iso_8859-2:1987": "iso-8859-2",
	"l2":              "iso-8859-2",
	"latin2":          "iso-8859-2",
	"8859-2":          "iso-8859-2",

	"csisolatin3":     "iso-8859-3",
	"iso-8859-3":      "iso-8859-3",
	"iso-ir-109":      "iso-8859-3",
	"iso8859-3":       "iso-8859-3",
	"iso88593":        "iso-8859-3",
	"iso_8859-3":      "iso-8859-3",
	"iso_8859-3:1988": "iso-8859-3",
	"l3":              "iso-8859-3",
	"latin3":          "iso-8859-3",
	"8859-3":          "iso-8859-3",

	"csisolatin4":     "iso-8859-4",
	"iso-8859-4":      "iso-8859-4",
	"iso-ir-110":      "iso-8859-4",
	"iso8859-4":       "iso-8859-4",
	"iso88594":        "iso-8859-4",
	"iso_8859-4":      "iso-8859-4",
	"iso_8859-4:1988": "iso-8859-4",
	"l4":              "iso-8859-4",
	"latin4":          "iso-8859-4",
	"8859-4":          "iso-8859-4",

	"csisolatincyrillic": "iso-8859-5",
	"cyrillic":           "iso-8859-5",
	"iso-8859-5":         "iso-8859-5",
	"iso-ir-144":         "iso-8859-5",
	"iso8859-5":          "iso-8859-5",
	"iso88595":           "iso-8859-5",
	"iso_8859-5":         "iso-8859-5",
	"iso_8859-5:1988":    "iso-8859-5",
	"8859-5":             "iso-8859-5",

	"arabic":           "iso-8859-6",
	"asmo-708":         "iso-8859-6",
	"csiso88596e":      "iso-8859-6",
	"csiso88596i":      "iso-8859-6",
	"csisolatinarabic": "iso-8859-6",
	"ecma-114":         "iso-8859-6",
	"iso-8859-6":       "iso-8859-6",
	"iso-8859-6-e":     "iso-8859-6",
	"iso-8859-6-i":     "iso-8859-6",
	"iso-ir-127":       "iso-8859-6",
	"iso8859-6":        "iso-8859-6",
	"iso88596":         "iso-8859-6",
	"iso_8859-6":       "iso-8859-6",
	"iso_8859-6:1987":  "iso-8859-6",
	"8859-6":           "iso-8859-6",

	"csisolatingreek": "iso-8859-7",
	"ecma-118":        "iso-8859-7",
	"elot_928":        "iso-8859-7",
	"greek":           "iso-8859-7",
	"greek8":          "iso-8859-7",
	"iso-8859-7":      "iso-8859-7",
	"iso-ir-126":      "iso-8859-7",
	"iso8859-7":       "iso-8859-7",
	"iso88597":        "iso-8859-7",
	"iso_8859-7":      "iso-8859-7",
	"iso_8859-7:1987": "iso-8859-7",
	"sun_eu_greek":    "iso-8859-7",
	"8859-7":          "iso-8859-7",

	"csiso88598e":      "iso-8859-8",
	"csisolatinhebrew": "iso-8859-8",
	"hebrew":           "iso-8859-8",
	"iso-8859-8":       "iso-8859-8",
	"iso-8859-8-e":     "iso-8859-8",
	"iso-ir-138":       "iso-8859-8",
	"iso8859-8":        "iso-8859-8",
	"iso88598":         "iso-8859-8",
	"iso_8859-8":       "iso-8859-8",
	"iso_8859-8:1988":  "iso-8859-8",
	"visual":           "iso-8859-8",
	"8859-8":           "iso-8859-8",

	"csiso88598i":  "iso-8859-8-i",
	"iso-8859-8-i": "iso-8859-8-i",
	"logical":      "iso-8859-8-i",
	"8859-8-i":     "iso-8859-8-i",

	"csisolatin6": "iso-8859-10",
	"iso-8859-10": "iso-8859-10",
	"iso-ir-157":  "iso-8859-10",
	"iso8859-10":  "iso-8859-10",
	"iso885910":   "iso-8859-10",
	"l6":          "iso-8859-10",
	"latin6":      "iso-8859-10",
	"8859-10":     "iso-8859-10",

	"iso-8859-13": "iso-8859-13",
	"iso8859-13":  "iso-8859-13",
	"iso885913":   "iso-8859-13",
	"8859-13":     "iso-8859-13",

	"iso-8859-14": "iso-8859-14",
	"iso8859-14":  "iso-8859-14",
	"iso885914":   "iso-8859-14",
	"8859-14":     "iso-8859-14",

	"csisolatin9": "iso-8859-15",
	"iso-8859-15": "iso-8859-15",
	"iso8859-15":  "iso-8859-15",
	"iso885915":   "iso-8859-15",
	"iso_8859-15": "iso-8859-15",
	"l9":          "iso-8859-15",
	"8859-15":     "iso-8859-15",

	"iso-8859-16": "iso-8859-16",
	"8859-16":     "iso-8859-16",

	"cskoi8r": "koi8-r",
	"koi":     "koi8-r",
	"koi8":    "koi8-r",
	"koi8-r":  "koi8-r",
	"koi8_r":  "koi8-r",

	"koi8-u": "koi8-u",

	"csmacintosh": "macintosh",
	"mac":         "macintosh",
	"macintosh":   "macintosh",
	"x-mac-roman": "macintosh",

	"dos-874":     "windows-874",
	"iso-8859-11": "windows-874",
	"iso8859-11":  "windows-874",
	"iso885911":   "windows-874",
	"tis-620":     "windows-874",
	"windows-874": "windows-874",

	"cp1250":       "windows-1250",
	"windows-1250": "windows-1250",
	"x-cp1250":     "windows-1250",

	"cp1251":       "windows-1251",
	"windows-1251": "windows-1251",
	"x-cp1251":     "windows-1251",

	"ansi_x3.4-1968":  "windows-1252",
	"ascii":           "windows-1252",
	"cp1252":          "windows-1252",
	"cp819":           "windows-1252",
	"csisolatin1":     "windows-1252",
	"ibm819":          "windows-1252",
	"iso-8859-1":      "windows-1252",
	"iso-ir-100":      "windows-1252",
	"iso8859-1":       "windows-1252",
	"iso8859_1":       "windows-1252",
	"iso88591":        "windows-1252",
	"iso_8859-1":      "windows-1252",
	"iso_8859-1:1987": "windows-1252",
	"l1":              "windows-1252",
	"latin1":          "windows-1252",
	"us-ascii":        "windows-1252",
	"windows-1252":    "windows-1252",
	"x-cp1252":        "windows-1252",
	"iso646-us":       "windows-1252", // ISO646 isn't us-ascii but 1991 version is.
	"iso: western":    "windows-1252", // same as iso-8859-1
	"we8iso8859p1":    "windows-1252", // same as iso-8859-1
	"8859-1":          "windows-1252",

	"cp1253":       "windows-1253",
	"windows-1253": "windows-1253",
	"x-cp1253":     "windows-1253",

	"cp1254":          "windows-1254",
	"csisolatin5":     "windows-1254",
	"iso-8859-9":      "windows-1254",
	"iso-ir-148":      "windows-1254",
	"iso8859-9":       "windows-1254",
	"iso88599":        "windows-1254",
	"iso_8859-9":      "windows-1254",
	"iso_8859-9:1989": "windows-1254",
	"l5":              "windows-1254",
	"latin5":          "windows-1254",
	"windows-1254":    "windows-1254",
	"x-cp1254":        "windows-1254",

	"cp1255":       "windows-1255",
	"windows-1255": "windows-1255",
	"x-cp1255":     "windows-1255",

	"cp1256":       "windows-1256",
	"windows-1256": "windows-1256",
	"x-cp1256":     "windows-1256",

	"cp1257":       "windows-1257",
	"windows-1257": "windows-1257",
	"x-cp1257":     "windows-1257",

	"cp1258":       "windows-1258",
	"windows-1258": "windows-1258",
	"x-cp1258":     "windows-1258",

	"x-mac-cyrillic":  "x-mac-cyrillic",
	"x-mac-ukrainian": "x-mac-cyrillic",

	"chinese":         "gbk",
	"csgb2312":        "gbk",
	"csiso58gb231280": "gbk",
	"gb2312":          "gbk",
	"gb_2312":         "gbk",
	"gb_2312-80":      "gbk",
	"gbk":             "gbk",
	"iso-ir-58":       "gbk",
	"x-gbk":           "gbk",
	"cp936":           "gbk", // same as gb2312

	"gb18030": "gb18030",

	"hz-gb-2312": "hz-gb-2312",

	"big5":       "big5",
	"big5-hkscs": "big5",
	"cn-big5":    "big5",
	"csbig5":     "big5",
	"x-x-big5":   "big5",
	"136":        "big5", // same as chinese big5

	"cseucpkdfmtjapanese": "euc-jp",
	"euc-jp":              "euc-jp",
	"x-euc-jp":            "euc-jp",

	"csiso2022jp": "iso-2022-jp",
	"iso-2022-jp": "iso-2022-jp",

	"csshiftjis":  "shift_jis",
	"ms_kanji":    "shift_jis",
	"shift-jis":   "shift_jis",
	"shift_jis":   "shift_jis",
	"sjis":        "shift_jis",
	"windows-31j": "shift_jis",
	"x-sjis":      "shift_jis",
	"cp932":       "shift_jis",

	"cseuckr":        "euc-kr",
	"csksc56011987":  "euc-kr",
	"euc-kr":         "euc-kr",
	"iso-ir-149":     "euc-kr",
	"korean":         "euc-kr",
	"ks_c_5601-1987": "euc-kr",
	"ks_c_5601-1989": "euc-kr",
	"ksc5601":        "euc-kr",
	"ksc_5601":       "euc-kr",
	"windows-949":    "euc-kr",

	"csiso2022kr":     "replacement",
	"iso-2022-kr":     "replacement",
	"iso-2022-cn":     "replacement",
	"iso-2022-cn-ext": "replacement",

	"utf-16be": "utf-16be",

	"utf-16":   "utf-16le",
	"utf-16le": "utf-16le",

	"x-user-defined": "x-user-defined",

	"cp850":  "cp850",
	"cp-850": "cp850",
	"ibm850": "cp850",
}

var charsetRegexp *regexp.Regexp
var errParsingCharset = errors.New("Could not find a valid charset in the HTML body")

func ConvertToUTF8String(charset string, textBytes []byte) (string, error) {
	if encodingAliases[strings.ToLower(charset)] == "utf-8" {
		return string(textBytes), nil
	}
	item, ok := encodings[encodingAliases[strings.ToLower(charset)]]
	if !ok {
		// Try to parse charset again here to see if we can salvage some badly formed ones
		// like charset="charset=utf-8"
		charsetp := strings.Split(charset, "=")
		if (strings.ToLower(charsetp[0]) == "charset" && len(charsetp) > 1) ||
			(strings.ToLower(charsetp[0]) == "iso" && len(charsetp) > 1) {
			charset = charsetp[1]
			item, ok = encodings[encodingAliases[strings.ToLower(charset)]]
			if !ok {
				return "", fmt.Errorf("Unsupport charset %s", charset)
			}
		} else {
			// Failed to get a conversion reader
			return "", fmt.Errorf("Unsupport charset %s", charset)
		}
	}
	input := bytes.NewReader(textBytes)
	reader := transform.NewReader(input, item.e.NewDecoder())
	output, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Look for charset in the html meta tag (v4.01 and v5)
func charsetFromHTMLString(htmlString string) (string, error) {
	if charsetRegexp == nil {
		var err error
		charsetRegexp, err = regexp.Compile(`(?i)<meta.*charset="?\s*(?P<charset>[a-zA-Z0-9_.:-]+)\s*"`)
		if err != nil {
			charsetRegexp = nil
			return "", err
		}
	}

	charsetMatches := charsetRegexp.FindAllStringSubmatch(htmlString, -1)

	if len(charsetMatches) > 0 {
		n1 := charsetRegexp.SubexpNames()
		r2 := charsetMatches[0]

		md := map[string]string{}
		for i, n := range r2 {
			md[n1[i]] = n
		}

		return md["charset"], nil
	}

	return "", errParsingCharset
}

package util

import (
	"regexp"
	"strings"
	"github.com/and-hom/wwmap/lib/model"
)

func HumanReadableCategoryNameWithBrackets(category model.SportCategory, translit bool) string {
	if category.Category == -1 {
		if translit {
			return "(Stop!)"
		} else {
			return "(Непроход)"
		}
	}
	if category.Category == 0 {
		return ""
	}
	if category.Sub == "" {
		return "(" + category.Serialize() + ")"
	}
	return "(" + category.Serialize() + ")"
}

func HumanReadableCategoryName(category model.SportCategory, translit bool) string {
	if category.Category == -1 {
		if translit {
			return "Stop!"
		} else {
			return "Непроход"
		}
	}
	if category.Category == 0 {
		return "-"
	}
	if category.Sub == "" {
		return category.Serialize()
	}
	return category.Serialize()
}

var translitCharMap = map[string]string{
	"a":"a",
	"б":"b",
	"в":"v",
	"г":"g",
	"д":"d",
	"е":"e",
	"ё":"e",
	"ж":"zh",
	"з":"z",
	"и":"i",
	"й":"j",
	"к":"k",
	"л":"l",
	"м":"m",
	"н":"n",
	"о":"o",
	"п":"p",
	"р":"r",
	"с":"s",
	"т":"t",
	"у":"u",
	"ф":"f",
	"х":"h",
	"ц":"ts",
	"ч":"ch",
	"ш":"sh",
	"щ":"sch",
	"ы":"y",
	"ь":"'",
	"э":"ye",
	"ю":"ju",
	"я":"ya",
}

func doReplace(data string, from string, to string) string {
	r, _ := regexp.Compile(from)
	return r.ReplaceAllString(data, to)
}

func CyrillicToTranslit(cyrillicString string) string {
	translitString := cyrillicString
	for k, v := range translitCharMap {
		translitString = doReplace(translitString, k, v)
		translitString = doReplace(translitString, strings.ToUpper(k), strings.Title(v))
	}
	return translitString
}

package aravia

import "regexp"

var (
	reCamealCase = regexp.MustCompile(`(_?[A-Z][a-z]+)`)
)

func GetWords(name string) (words []string) {
	matches := reCamealCase.FindAllStringSubmatch(name, -1)
	for _, match := range matches {
		words = append(words, match[0])
	}
	return
}

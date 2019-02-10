// markdown
package markdownhelper

import (
	"strings"
)

var mdSpecialChars []string = []string{"\\", "`", "*", "_", "{", "}", "[", "]", "(", ")", "#", "!"}

func MDEscape(s string) string {
	for _, c := range mdSpecialChars {
		s = strings.Replace(s, c, "\\"+c, -1)
	}
	return s
}

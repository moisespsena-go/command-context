package cmdctx

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var splitRegexp = regexp.MustCompile(`\s+`)

func splitString(s string, maxLen int) (res []string) {
	words := splitRegexp.Split(s, -1)

	var (
		line  []string
		lineC int
	)
	for len(words) > 0 {
		word := words[0]
		words = words[1:]

		if lineC+len(word) > maxLen {
			res = append(res, strings.Join(line, " "))
			lineC = 0
			line = line[:0]
		}
		lineC += len(word)
		line = append(line, word)
	}

	if len(line) > 0 {
		res = append(res, strings.Join(line, " "))
	}

	return
}

func consoleSize() (width int, heigth int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return
	}

	s := string(out)
	s = strings.TrimSpace(s)
	sArr := strings.Split(s, " ")

	heigth, err = strconv.Atoi(sArr[0])
	if err != nil {
		return
	}

	width, err = strconv.Atoi(sArr[1])
	if err != nil {
		return
	}
	return width, heigth
}

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	initNumsRegex        = regexp.MustCompile(`var [Aa-z0-9]{64}=[0-9]+`)
	basicMathRegex       = regexp.MustCompile(`[a-z0-9]{64}=(~|\^|\||&|[A-z0-9]{64})`)
	funcEndingRegex      = regexp.MustCompile(`}\([a-z0-9]{64},[a-z0-9]{64},[a-z0-9]{64}\)`)
	insideRightShiftFunc = false
	rightShiftFuncKey    = ""
	insideMathXORFunc    = false
	mathXORFuncKey       = ""
)

func main() {
	// load sample file
	f, err := os.ReadFile("./sample.js")
	if err != nil {
		panic(err)
	}
	script := strings.Split(string(f), "\n")[2]
	script = script[41 : len(script)-4]
	operations := strings.Split(script, ";")
	solution := make(map[string]int, 4)
	for _, op := range operations {
		// get initial numbers
		if initNumsRegex.MatchString(op) {
			parts := strings.Split(initNumsRegex.FindString(op), "=")
			value, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}
			solution[parts[0][4:]] = value
			continue
		}
		// basic math, like xxx ^ xxx, ~xxx, etc.
		if basicMathRegex.MatchString(op) && !strings.Contains(op, "new Date") {
			signChange := false
			mathDone := false
			// handle ~, which changes the sign
			if strings.Contains(op, "~") {
				// handle rather it's a `~(xxx ^ xxx)` op or not.
				if strings.Contains(op, "(") {
					// trim off `~(` and `)`
					tmp := strings.Split(op, "=")
					newPart := tmp[1]
					newPart = newPart[2 : len(newPart)-1]
					op = tmp[0] + "=" + newPart
				} else {
					// trim off just `~`
					tmp := strings.Split(op, "=")
					newPart := tmp[1]
					newPart = newPart[1:]
					op = tmp[0] + "=" + newPart
				}
				signChange = true
			}
			parts := strings.Split(op, "=")
			// handle all the different operations
			if strings.Contains(parts[1], "^") {
				tmp := strings.Split(parts[1], "^")
				solution[parts[0]] = solution[tmp[0]] ^ solution[tmp[1]]
				mathDone = true
			}
			if strings.Contains(parts[1], "|") {
				tmp := strings.Split(parts[1], "|")
				solution[parts[0]] = solution[tmp[0]] | solution[tmp[1]]
				mathDone = true
			}
			if strings.Contains(parts[1], "&") {
				tmp := strings.Split(parts[1], "&")
				solution[parts[0]] = solution[tmp[0]] & solution[tmp[1]]
				mathDone = true
			}
			if signChange {
				// if math was done, then the answer should be answers[parts[0]] instead of parts[1]
				if mathDone {
					solution[parts[0]] = -(solution[parts[0]] + 1)
				} else {
					solution[parts[0]] = -(solution[parts[1]] + 1)
				}
			}
		}
		if strings.Contains(op, "new Date") {
			parts := strings.Split(op, "=")
			opParts := strings.Split(parts[1], "^")
			solution[parts[0]] = solution[opParts[0]] ^ time.UnixMilli(int64(solution[strings.Split(strings.Split(opParts[1], "*")[0], "(")[1]]*10000000000)).UTC().Day()
		}
		// detect the rightShiftFunc starting
		if strings.Contains(op, "document.createElement('div')") && !insideRightShiftFunc {
			insideRightShiftFunc = true
			rightShiftFuncKey = strings.Split(op, "=function")[0]
		}
		// detect the rightShiftFunc ending
		if funcEndingRegex.MatchString(op) && insideRightShiftFunc {
			insideRightShiftFunc = false
			in := strings.Split(op[2:len(op)-1], ",")
			solution[rightShiftFuncKey] = mathRightShift(solution[in[0]], solution[in[1]], solution[in[2]])
			rightShiftFuncKey = ""
		}
		// detect the mathXORFunc starting
		if strings.Contains(op, "function(){return this.") && !insideMathXORFunc {
			insideMathXORFunc = true
			mathXORFuncKey = strings.Split(op, "=")[0]
		}
		// detect the mathXORFunc ending
		if funcEndingRegex.MatchString(op) && insideMathXORFunc {
			insideMathXORFunc = false
			in := strings.Split(op[2:len(op)-1], ",")
			solution[mathXORFuncKey] = mathXOR(solution[in[0]], solution[in[1]], solution[in[2]])
			mathXORFuncKey = ""
		}
		if strings.HasPrefix(op, "return {'rf") {
			break
		}
	}
	fmt.Printf("%+v\n", solution)
}

func mathXOR(a, b, c int) int {
	return (b ^ a) | (c ^ b)
}

func mathRightShift(a, b, c int) int {
	num := 0
	for i := 0; i < 8; i++ {
		if (a & 1) == 0 {
			num += a
		}
		if (b & 1) == 0 {
			num += b
		}
		if (c & 1) == 0 {
			num += c
		}
		a = a >> 1
		b = b >> 1
		c = c >> 1
	}
	return num % 256
}

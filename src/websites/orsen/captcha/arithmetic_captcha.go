package captcha

import (
	"strconv"
	"strings"
)

// SolveArithmeticCaptcha returns answer to a given captcha as string
// captcha - arithmetic equation with + and - operations and one unknown variable on either side.
func SolveArithmeticCaptcha(captcha string) (answer string) {
	//assuming captcha consists of numbers, +, -, =, and ? on either side of the expression
	sides := strings.Split(captcha, "=")

	splitFn := func(c rune) bool {
		return c == ' '
	}
	leftSide, rightSide := strings.FieldsFunc(sides[0], splitFn), strings.FieldsFunc(sides[1], splitFn)
	diff := solveExpression(rightSide) - solveExpression(leftSide)
	diff *= getDiffSign(leftSide, rightSide)
	return strconv.Itoa(diff)
}

func solveExpression(e []string) int {
	result := 0

	for i := 0; i < len(e); i += 2 {
		if val, err := strconv.Atoi(e[i]); err == nil {
			if i == 0 || e[i-1] == "+" {
				result += val
			} else if e[i-1] == "-" {
				result -= val
			} else {
				panic("Unknown operation: " + e[i-1])
			}
		}
	}
	return result
}

// Is 'e' not a number and not a sign?
func isVariable(e string) bool {
	_, err := strconv.Atoi(e)
	return err != nil && e != "+" && e != "-" && e != "="
}

// Determine if the equation sum should be negated (is x negative, is x on the right side?)
func getDiffSign(leftSide, rightSide []string) int {
	for i, e := range leftSide {
		if isVariable(e) {
			if i == 0 || leftSide[i-1] == "+" {
				return 1
			} else {
				return -1
			}
		}
	}

	for i, e := range rightSide {
		if isVariable(e) {
			if i == 0 || rightSide[i-1] == "+" {
				return -1
			} else {
				return 1
			}
		}
	}
	panic("Variable not found in the expression")
}

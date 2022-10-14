package prompt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Ask is a show a generic prompt asking for user input, with validation
func Ask(prompt string, validate func(s string) error) {
	s := ""
	for {
		fmt.Printf("%s ", prompt)
		fmt.Scanln(&s)
		if err := validate(s); err == nil {
			break
		} else {
			fmt.Printf("%s, please try again (Ctrl+C to exit)\n\n", err.Error())
		}
	}
}

// YNAuto allows to configure global automatic answers to YN, this may be
// useful for cli applications that expose '-y' or '-n' flags to bypass
// interactive Y/N choices (see AutoYN function)
type YNAuto int

const (
	AlwaysAsk = YNAuto(iota) // always prompt
	AssumeY                  // ignore prompt, automatically assume yes
	AssumeN                  // ignore prompt, automatically assume no
)

// YNPolicy provides tuning options for AskYN prompts
type YNPolicy int

const (
	NoAuto   = YNPolicy(1 << iota) // ignore YNAuto, always show prompt
	DefaultY                       // assume 'y' when user provides empty input (hits RETURN key without typing anything)
	DefaultN                       // assume 'n' when user provides empty input (hits RETURN key without typing anything)
)

// AutoYN configures automatic YN answers for non-interactive runs
func AutoYN(v YNAuto) {
	autoYN = v
}

var autoYN YNAuto = AlwaysAsk
var ErrInvalidInput = errors.New("invalid input")

// YN displays a [y/n] prompt, returning a boolean value after the user types a response
// - 'y'|'yes' -> returns true
// - 'n'|'no' -> returns false
//
// # Automatic answers are configured globally with AutoYN() and can be overriden with NoAuto policy
//
// Default answers (when user hits RETURN key without typing anything) can be configured with DefaultY or DefaultN policy
func YN(prompt string, policies ...YNPolicy) bool {
	hasPolicy := func(policy YNPolicy) bool {
		for _, p := range policies {
			if p == policy {
				return true
			}
		}
		return false
	}

	if autoYN != AlwaysAsk && !hasPolicy(NoAuto) {
		return autoYN == AssumeY
	}

	ret := false
	Ask(prompt, func(s string) error {
		s = strings.ToLower(s)
		switch strings.TrimSpace(strings.ToLower(s)) {
		case "y", "yes":
			ret = true
			return nil
		case "n", "no":
			ret = false
			return nil
		case "":
			if hasPolicy(DefaultY) {
				ret = true
				return nil
			} else if hasPolicy(DefaultN) {
				ret = false
				return nil
			}
		}
		return ErrInvalidInput
	})
	return ret
}

// Int shows a prompt and processes user input as a typed integer value, accepting values between min and max
func Int(prompt string, min, max int) int {
	if max < min {
		panic("invalid int prompt usage: max is smaller than min")
	}
	ret := 0
	Ask(prompt, func(s string) error {
		v, err := strconv.Atoi(s)
		ret = int(v)
		if err == nil && ret >= min && ret <= max {
			return nil
		}
		return ErrInvalidInput
	})
	return ret
}

// Choose shows a prompt as:
//
//	<prompt>
//	1: <choices[0]>
//	2: <choices[1]>
//	...
//	N: <choices[N-1]>
//	type a number [1...N]: _
//
// # Then it awaits user input and returns the numeric value within 1...N range
//
// Notice, that the numbers shown and the returned values are 1...N, (not 0...N-1)
func Choose(prompt string, choices ...string) int {
	if len(choices) == 0 {
		panic("invalid choice prompt usage: no choices provided")
	}
	fmt.Println(prompt)
	for i, choice := range choices {
		fmt.Printf("%d: %s\n", i+1, choice)
	}
	return Int(fmt.Sprintf("type a number [1...%d]: ", len(choices)), 1, len(choices))
}

// Enum shows a prompt as:
//
//	<promptPrefix> [<choices[0]>/<choices[1]>/.../<choices[N-1]>]: _
//
// Then it awaits user input as text value and returns the matching chose as a string
func Enum(promptPrefix string, choices ...string) string {
	ret := ""
	Ask(fmt.Sprintf("%s [%s]:", promptPrefix, strings.Join(choices, "/")),
		func(s string) error {
			for _, choice := range choices {
				if s == choice {
					ret = s
					return nil
				}
			}
			return ErrInvalidInput
		})
	return ret
}

// Text shows a prompt as:
//
//	<prompt> _
//
// Then it awaits user input as text, validates it with the provided validation
// callbacks and returns the typed value if validation was successful.
func Text(prompt string, validate ...func(string) error) string {
	ret := ""
	Ask(prompt, func(s string) error {
		ret = s
		for i := range validate {
			err := validate[i](ret)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return ret
}

// TextTrimmed works similarly to Text, but trims spaces from typed string
// before it is validated and returned
//
// some of the validators are defined below
//
// note that you can also use Validate* functions from filesystem.stats
func TextTrimmed(prompt string, validate ...func(string) error) string {
	ret := ""
	Ask(prompt, func(s string) error {
		ret = s
		for i := range validate {
			err := validate[i](ret)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return ret
}

// predefined validator errors
var ErrEmptyStringNotAllowed = errors.New("empty string is not allowed")
var ErrSpacesNotAllowed = errors.New("spaces are not allowed within the string")

// NoEmptyString can be used as validating callback for Text and TextTrimmed calls
func NoEmptyString(s string) error {
	if len(s) == 0 {
		return ErrEmptyStringNotAllowed
	}
	return nil
}

func NoSpaces(s string) error {
	if strings.IndexFunc(s, unicode.IsSpace) != -1 {
		return ErrSpacesNotAllowed
	}
	return nil
}

// Package parameters provides an argument parser for dealing with a mix of quoted and unquoted arguments.
package parameters

import (
	"strings"
)

var quotePairs = map[string]string{
	// Normal quotes
	"'":  "'",
	"\"": "\"",
	// Smart quotes
	"\u201C\u201D\u201F\u201E": "\u201C\u201D\u201F",
	"\u2018\u2019\u201B\u201A": "\u2018\u2019\u201B",
	// Chevrons
	"\u00AB\u300A": "\u00BB\u300B",
	"\u00BB\u300B": "\u00AA\u300A",
	"\u2039\u3008": "\u203A\u3009",
	"\u203A\u3009": "\u2039\u3008",
	// Corner brackets
	"\u300C\u300E": "\u300D\u300F",
}

// Parameters is an argument parser. It is not thread-safe.
type Parameters struct {
	flags      map[string]struct{}
	ptr        int
	matchFlags bool

	fullCommand []rune
}

type wordPosition struct {
	startPos, endPos int
	advanceAfterWord int
	wasQuoted        bool
}

// NewParameters returns a new Parameters.
// If matchFlags is false, flags (unquoted arguments starting with -) will be treated the same as other arguments.
func NewParameters(s string, matchFlags bool) *Parameters {
	return &Parameters{fullCommand: []rune(s), matchFlags: matchFlags}
}

func (p *Parameters) parseFlags() {
	p.flags = map[string]struct{}{}

	ptr := 0
	for {
		wp := p.nextWordPosition(ptr)
		if wp == nil {
			break
		}

		ptr = wp.endPos + wp.advanceAfterWord
		if p.fullCommand[wp.startPos] != '-' || wp.wasQuoted {
			continue
		}

		start := wp.startPos
		for start < len(p.fullCommand) && p.fullCommand[start] == '-' {
			start++
		}
		name := strings.TrimSpace(string(p.fullCommand[start:wp.endPos]))
		if name != "" {
			p.flags[strings.ToLower(name)] = struct{}{}
		}
	}
}

func (p *Parameters) nextWordPosition(position int) *wordPosition {
	// skip leading spaces before actual content
	for position < len(p.fullCommand) && p.fullCommand[position] == ' ' {
		position++
	}

	// if it's the end of the string, return nil
	if len(p.fullCommand) <= position {
		return nil
	}

	if right, ok := checkQuote(p.fullCommand[position]); ok {
		// found a quoted word, find an instance of the corresponding end quotes
		endQuotePosition := -1
		for i := position + 1; i < len(p.fullCommand); i++ {
			if strings.ContainsRune(right, p.fullCommand[i]) {
				endQuotePosition = i
				break
			}
		}

		if len(p.fullCommand) == endQuotePosition+1 || p.fullCommand[endQuotePosition+1] == ' ' {
			return &wordPosition{position + 1, endQuotePosition, 2, true}
		}
	}

	wordEnd := firstIndex(p.fullCommand, ' ', position+1)
	if wordEnd == -1 {
		return &wordPosition{position, len(p.fullCommand), 0, false}
	}
	return &wordPosition{position, wordEnd, 1, false}
}

// firstIndex returns the index of the first char in s, starting at min.
// If min is out of bounds, or char does not exist in s past min, returns -1.
func firstIndex(s []rune, char rune, min int) int {
	if min >= len(s) || min < 0 {
		return -1
	}

	substr := s[min:]
	for i, r := range substr {
		if char == r {
			return int(min) + i
		}
	}
	return -1
}

func checkQuote(potentialLeftQuote rune) (rightQuotes string, ok bool) {
	for left, right := range quotePairs {
		if strings.ContainsRune(left, potentialLeftQuote) {
			return right, true
		}
	}

	return "", false
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Pop returns the next argument in the parameters, skipping over flags.
func (p *Parameters) Pop() string {
	// ignore + skip flags
	for {
		wp := p.nextWordPosition(p.ptr)
		if wp == nil {
			break
		}
		p.ptr = wp.endPos + wp.advanceAfterWord
		if p.matchFlags && p.fullCommand[wp.startPos] == '-' && !wp.wasQuoted {
			continue
		}

		return strings.TrimSpace(
			string(p.fullCommand[wp.startPos:wp.endPos]),
		)
	}

	return ""
}

// Peek returns the next argument in the parameters, but doesn't move the pointer.
func (p *Parameters) Peek() string {
	s, _ := p.PeekPtr(p.ptr)
	return s
}

// PeekPtr is like Pop, but uses a user-supplied pointer rather than p's.
func (p *Parameters) PeekPtr(ptr int) (string, int) {
	// ignore + skip flags
	for {
		wp := p.nextWordPosition(ptr)
		if wp == nil {
			break
		}

		if p.matchFlags && p.fullCommand[wp.startPos] == ' ' && !wp.wasQuoted {
			continue
		}

		return strings.TrimSpace(
			string(p.fullCommand[wp.startPos:wp.endPos]),
		), ptr
	}

	return "", ptr
}

// Remainder returns the remaining arguments.
// skipFlags only has an effect if matchFlags is set to true.
func (p *Parameters) Remainder(skipFlags bool) string {
	if p.matchFlags && skipFlags {
		ptr := p.ptr

		for {
			wp := p.nextWordPosition(ptr)
			if wp == nil {
				break
			}

			if p.fullCommand[wp.startPos] != ' ' || wp.wasQuoted {
				break
			}
			ptr = wp.endPos + wp.advanceAfterWord
		}
	}

	return strings.TrimSpace(
		string(p.fullCommand[minInt(p.ptr, len(p.fullCommand)):]),
	)
}

// Flags returns p's parsed flags.
// if matchFlags is set to false, always returns nil.
func (p *Parameters) Flags() []string {
	if !p.matchFlags {
		return nil
	}

	if p.flags == nil {
		p.parseFlags()
	}

	flags := make([]string, 0, len(p.flags))
	for flag := range p.flags {
		flags = append(flags, flag)
	}
	return flags
}

// Args returns all remaining arguments as a slice of strings.
// Use p.Remainder to return all remaining arguments as a single string.
func (p *Parameters) Args() []string {
	var args []string

	for {
		s := p.Pop()
		if s == "" {
			break
		}

		args = append(args, s)
	}

	return args
}

// FullCommand returns the full command passed into parameters.New.
func (p *Parameters) FullCommand() string {
	return string(p.fullCommand)
}

// HasNext returns true if p has any parameters left.
func (p *Parameters) HasNext() bool {
	return p.Remainder(true) != ""
}

// Match returns (match, true) if any of the arguments matched, and ("", false) if none did.
// If an argument matches, the matched string is popped from p.
func (p *Parameters) Match(potentialMatches ...string) (string, bool) {
	arg := p.Peek()
	for _, match := range potentialMatches {
		if strings.EqualFold(arg, match) {
			return p.Pop(), true
		}
	}
	return "", false
}

// PeekMatch is like Match but doesn't pop the argument from p.
func (p *Parameters) PeekMatch(potentialMatches ...string) (string, bool) {
	arg := p.Peek()
	for _, match := range potentialMatches {
		if strings.EqualFold(arg, match) {
			return arg, true
		}
	}
	return "", false
}

// MatchFlags matches the given flags against the flags in p.
// Arguments should always be passed lowercase.
func (p *Parameters) MatchFlags(flags ...string) bool {
	for _, flag := range p.Flags() {
		for _, potentialFlag := range flags {
			if flag == potentialFlag {
				return true
			}
		}
	}
	return false
}

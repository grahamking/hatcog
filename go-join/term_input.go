package main

/*
 * A line of keyboard input. Supports basic line editing.
 */
type TermInput struct {
	input     []byte
	cursorPos int
}

func NewTermInput() *TermInput {
	return &TermInput{
		input: make([]byte, 0, 80),
	}
}

func (self *TermInput) Len() int {
	return len(self.input)
}

func (self *TermInput) Backspace() {

	if self.cursorPos == len(self.input) {
		self.input = self.input[:len(self.input)-1]
	} else {
		self.input = append(
			self.input[:self.cursorPos-1],
			self.input[self.cursorPos:]...)
	}
	self.cursorPos -= 1
}

func (self *TermInput) KeyLeft() {

	self.cursorPos -= 1
	if self.cursorPos < 0 {
		self.cursorPos = 0
	}
}

func (self *TermInput) KeyRight() {

	self.cursorPos += 1
	if self.cursorPos > len(self.input) {
		self.cursorPos = len(self.input)
	}
}

func (self *TermInput) KeyHome() {
	self.cursorPos = 0
}

func (self *TermInput) KeyEnd() {
	self.cursorPos = len(self.input)
}

func (self *TermInput) Add(char byte) {

	if self.cursorPos == len(self.input) {
		self.input = append(self.input, char)
	} else {
		// Insert
		self.input = append(
			self.input[:self.cursorPos],
			append([]byte{char}, self.input[self.cursorPos:]...)...)
	}
	self.cursorPos += 1
}

func (self *TermInput) String() string {
	return string(self.input)
}

func (self *TermInput) Reset() {
	self.input = make([]byte, 0)
	self.cursorPos = 0
}

// String between current position and previous space
func (self *TermInput) Prefix() string {
	prefix := make([]byte, 0, 10)
	for pos := self.cursorPos - 1; pos >= 0 && self.input[pos] != ' '; pos-- {
		prefix = append(prefix, self.input[pos])
	}
	return reverse(string(prefix))
}

// Replace from nearest space to the left to current cursor pos.
func (self *TermInput) ReplaceWord(word string) {
	for len(self.input) > 0 && self.input[self.cursorPos-1] != ' ' {
		self.Backspace()
	}
	self.input = append(
		self.input[:self.cursorPos],
		append([]byte(word), self.input[self.cursorPos:]...)...)
	self.cursorPos += len(word)
}

// http://stackoverflow.com/questions/1752414/how-to-reverse-a-string-in-go
func reverse(val string) string {
	n := 0
	rune := make([]int, len(val))
	for _, r := range val {
		rune[n] = r
		n++
	}
	rune = rune[0:n]
	// Reverse
	for i := 0; i < n/2; i++ {
		rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
	}
	// Convert back to UTF-8.
	output := string(rune)
	return output
}

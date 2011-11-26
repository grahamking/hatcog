package main

/*
 * A line of keyboard input. Supports basic line editing.
 */
type TermInput struct {
    input       []byte
    cursorPos   int
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
        self.input = append(self.input[:self.cursorPos], append([]byte{char}, self.input[self.cursorPos:]...)...)
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


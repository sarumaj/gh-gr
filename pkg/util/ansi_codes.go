package util

const (
	ESC = "\x1b"     // Begin ANSI sequence
	CSI = ESC + "["  // Control Sequence Introducer
	MUP = CSI + "1A" // Move cursor 1 line up
	CCT = "m"        // Color Code Terminator
)

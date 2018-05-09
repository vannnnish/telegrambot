// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
	"time"
)

var stdout io.Writer = os.Stdout

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109}) // 0x1B[97;42m
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109}) // 0x1B[90;47m
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109}) // 0x1B[97;43m
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109}) // 0x1B[97;41m
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109}) // 0x1B[97;44m
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109}) // 0x1B[97;45m 紫红色
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109}) // 0x1B[97;46m 青蓝色
	reset   = string([]byte{27, 91, 48, 109})                 // 0x1B[0m
)

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	format string
	w      chan *LogRecord
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	consoleWriter := &ConsoleLogWriter{
		format: "[%T %D] [%L] (%S) %M",
		w:      make(chan *LogRecord, LogBufferLength),
	}
	go consoleWriter.run(stdout)
	return consoleWriter
}

func (c *ConsoleLogWriter) SetFormat(format string) {
	c.format = format
}

func (c *ConsoleLogWriter) run(out io.Writer) {
	for rec := range c.w {
		switch rec.Level {
		case DEBUG:
			rec.Message = fmt.Sprintf("%s%s%s", green, rec.Message, reset)
		case ERROR:
			rec.Message = fmt.Sprintf("%s%s%s", red, rec.Message, reset)
		}
		fmt.Fprint(out, FormatLogRecord(c.format, rec))
	}
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (c *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	c.w <- rec
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (c *ConsoleLogWriter) Close() {
	close(c.w)
	time.Sleep(50 * time.Millisecond) // Try to give console I/O time to complete
}

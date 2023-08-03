/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package sse

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"
)

// Event holds all of the event source fields
type Event struct {
	timestamp time.Time
	// 事件 ID，会成为当前 EventSource 对象的内部属性“最后一个事件 ID”的属性值。
	ID []byte
	// 消息的数据字段。
	// 当 EventSource 接收到多个以 data: 开头的连续行时，会将它们连接起来，在它们之间插入一个换行符。
	// 末尾的换行符会被删除。
	Data []byte
	// 一个用于标识事件类型的字符串。
	// 如果指定了这个字符串，浏览器会将具有指定事件名称的事件分派给相应的监听器；
	// 网站源代码应该使用 addEventListener() 来监听指定的事件。
	// 如果一个消息没有指定事件名称，那么 onmessage 处理程序就会被调用。
	Event []byte
	// 重新连接的时间。
	// 如果与服务器的连接丢失，浏览器将等待指定的时间，然后尝试重新连接。
	// 这必须是一个整数，以毫秒为单位指定重新连接的时间。
	// 如果指定了一个非整数值，该字段将被忽略。
	Retry   []byte
	Comment []byte
}

func (e *Event) hasContent() bool {
	return len(e.ID) > 0 || len(e.Data) > 0 || len(e.Event) > 0 || len(e.Retry) > 0
}

// EventStreamReader scans an io.Reader looking for EventStream messages.
type EventStreamReader struct {
	scanner *bufio.Scanner
}

// NewEventStreamReader creates an instance of EventStreamReader.
func NewEventStreamReader(eventStream io.Reader, maxBufferSize int) *EventStreamReader {
	scanner := bufio.NewScanner(eventStream)
	initBufferSize := minPosInt(4096, maxBufferSize)
	scanner.Buffer(make([]byte, initBufferSize), maxBufferSize)

	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// We have a full event payload to parse.
		if i, nlen := containsDoubleNewline(data); i >= 0 {
			return i + nlen, data[0:i], nil
		}
		// If we're at EOF, we have all of the data.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	// Set the split function for the scanning operation.
	scanner.Split(split)

	return &EventStreamReader{
		scanner: scanner,
	}
}

// Returns a tuple containing the index of a double newline, and the number of bytes
// represented by that sequence. If no double newline is present, the first value
// will be negative.
func containsDoubleNewline(data []byte) (int, int) {
	// Search for each potentially valid sequence of newline characters
	crcr := bytes.Index(data, []byte("\r\r"))
	lflf := bytes.Index(data, []byte("\n\n"))
	crlflf := bytes.Index(data, []byte("\r\n\n"))
	lfcrlf := bytes.Index(data, []byte("\n\r\n"))
	crlfcrlf := bytes.Index(data, []byte("\r\n\r\n"))
	// Find the earliest position of a double newline combination
	minPos := minPosInt(crcr, minPosInt(lflf, minPosInt(crlflf, minPosInt(lfcrlf, crlfcrlf))))
	// Detemine the length of the sequence
	nlen := 2
	if minPos == crlfcrlf {
		nlen = 4
	} else if minPos == crlflf || minPos == lfcrlf {
		nlen = 3
	}
	return minPos, nlen
}

// Returns the minimum non-negative value out of the two values. If both
// are negative, a negative value is returned.
func minPosInt(a, b int) int {
	if a < 0 {
		return b
	}
	if b < 0 {
		return a
	}
	if a > b {
		return b
	}
	return a
}

// ReadEvent scans the EventStream for events.
func (e *EventStreamReader) ReadEvent() ([]byte, error) {
	if e.scanner.Scan() {
		event := e.scanner.Bytes()
		return event, nil
	}
	if err := e.scanner.Err(); err != nil {
		if err == context.Canceled {
			return nil, io.EOF
		}
		return nil, err
	}
	return nil, io.EOF
}

// Copyright 2023 The aichat Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

var ErrPushNotImplemented = errors.New("push not implemented")
var ErrHijackNotImplemented = errors.New("hijack not implemented")

type ResponseWriterWrapper struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{ResponseWriter: w}
}

// Unwrap returns the underlying ResponseWriter.
func (rww *ResponseWriterWrapper) Unwrap() http.ResponseWriter {
	return rww.ResponseWriter
}

// Flush implements http.Flusher. It simply calls the underlying
// ResponseWriter's Flush method if there is one.
func (rww *ResponseWriterWrapper) Flush() {
	if f, ok := rww.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker. It simply calls the underlying
// ResponseWriter's Hijack method if there is one, or returns
// ErrNotImplemented otherwise.
func (rww *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rww.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, ErrHijackNotImplemented
}

// Push implements http.Pusher. It simply calls the underlying
// ResponseWriter's Push method if there is one, or returns
// ErrNotImplemented otherwise.
func (rww *ResponseWriterWrapper) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rww.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return ErrPushNotImplemented
}

// ReadFrom implements io.ReaderFrom. It simply calls the underlying
// ResponseWriter's ReadFrom method if there is one, otherwise it defaults
// to io.Copy.
func (rww *ResponseWriterWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	if rf, ok := rww.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(rww.ResponseWriter, r)
}

// WriteHeader records the value of the status code before writing it.
func (rww *ResponseWriterWrapper) WriteHeader(code int) {
	rww.StatusCode = code
	rww.ResponseWriter.WriteHeader(code)
}

// Write computes the written len and stores it in ContentLength.
func (rww *ResponseWriterWrapper) Write(b []byte) (int, error) {
	n, err := rww.ResponseWriter.Write(b)
	rww.ContentLength += n
	return n, err
}

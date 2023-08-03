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

package requestid

import (
	"encoding/hex"

	"github.com/google/uuid"
)

// New 请求id  长度 = 32 char
func New() string {
	return ToString(uuid.New())
}

func NewByes() []byte {
	return ToBytes(uuid.New())
}

func ToString(v uuid.UUID) string {
	return string(ToBytes(v))
}

func ToBytes(v uuid.UUID) []byte {
	var buf [32]byte
	encodeHex(buf[:], v)
	return buf[:]
}

func encodeHex(dst []byte, v uuid.UUID) {
	hex.Encode(dst, v[:4])
	hex.Encode(dst[8:12], v[4:6])
	hex.Encode(dst[12:16], v[6:8])
	hex.Encode(dst[16:20], v[8:10])
	hex.Encode(dst[20:], v[10:])
}

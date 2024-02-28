// Copyright 2023-2024 The aichat Authors. All rights reserved.
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

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// -代表无权限，r代表读权限，w代表写权限，x代表执行权限
// 数字权限使用格式 r=4, w=2, x=1, -=0
// 0755 拥有者有读、写、执行权限；而属组用户和其他用户只有读、执行权限。
const (
	ModePerm0644 = 0644 // 文件保存perm
	ModePerm0755 = 0755 // 目录保存perm 一个用户需要进入一个目录查看文件,最少需要读和执行的权限。
)

// ExecAppDir gets compiled executable file directory: C:\option\_test
func ExecAppDir() string {
	return filepath.Dir(ExecAppPath())
}

// ExecAppPath gets compiled executable file absolute path: C:\option\_test\option.test.exe
func ExecAppPath() string {
	_path, _ := os.Executable()
	return _path
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	if fileInfo, err := os.Stat(path); err == nil {
		return fileInfo.IsDir()
	}
	return false
}

// CreateDir 创建目录
func CreateDir(path string) error {
	if !IsDir(path) {
		if err := os.MkdirAll(path, ModePerm0755); err != nil {
			return fmt.Errorf("os.Mkdir failed, directory: %q, cause %w", path, err)
		}
	}
	return nil
}

// FilenameWithoutExt 没有后缀的文件名称
func FilenameWithoutExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

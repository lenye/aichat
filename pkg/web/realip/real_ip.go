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

package realip

import (
	"net"
	"net/http"
	"strings"
)

const (
	hdrXRealIP        = "X-Real-IP"
	hdrCfConnectingIp = "Cf-Connecting-Ip"
	hdrXForwardedFor  = "X-Forwarded-For"
)

/*
https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/X-Forwarded-For
语法
  X-Forwarded-For: <client>, <proxy1>, <proxy2>
如果一个请求经过了多个代理服务器，那么每一个代理服务器的 IP 地址都会被依次记录在内。
也就是说，最右端的 IP 地址表示最近通过的代理服务器，而最左端的 IP 地址表示最初发起请求的客户端的 IP 地址。
*/

// ClientIP 客户端ip
func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	var ip string
	// cloudflare
	ip = r.Header.Get(hdrCfConnectingIp)
	if ip != "" {
		return ip
	}

	// nginx 配置 X-Real-IP
	ip = r.Header.Get(hdrXRealIP)
	if ip != "" {
		return ip
	}

	xForwardedFor := r.Header.Get(hdrXForwardedFor)
	if xForwardedFor == "" {
		// Get the remote addr
		var err error
		// Get the remote addr
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return ip
	}
	ip = strings.Split(xForwardedFor, ",")[0]

	return strings.TrimSpace(ip)
}

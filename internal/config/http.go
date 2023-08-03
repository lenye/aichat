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

package config

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

func HttpListenAndServe(handler http.Handler,
	cfg *HttpServerConfig,
	wg *sync.WaitGroup,
	logger *slog.Logger) (*http.Server, error) {

	address := fmt.Sprintf(":%d", cfg.Port)
	// http server
	ln, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("start http server failed",
			slog.Any("error", err),
		)
		return nil, err
	}
	logger.Info("http server listening on " + ln.Addr().String())

	svr := &http.Server{
		Handler: handler,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := svr.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http serve failed",
				slog.Any("error", err),
			)
		}
		logger.Info("http server stopped")
	}()

	return svr, nil
}

func HttpShutdown(svr *http.Server, logger *slog.Logger) {
	if svr == nil {
		logger.Debug("http server is not running, skip to shutdown")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		logger.Error("http server shutdown failed",
			slog.Any("error", err),
		)
	}
	logger.Info("http server shutdown")
}

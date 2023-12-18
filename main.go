// Copyright 2023 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"xconfwebconfig/dataapi"

	"xconfadmin/adminapi"
	xhttp "xconfadmin/http"
	"xconfwebconfig/common"
	xwhttp "xconfwebconfig/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	_ "go.uber.org/automaxprocs"
)

const (
	defaultConfigFile = "/app/xconfadmin/xconfwebconfig.conf"
)

// main function to boot up everything
func main() {
	started := time.Now().String()

	// parse flag
	configFile := flag.String("f", defaultConfigFile, "config file")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("webconfig version %s (branch %v) %v\n", common.BinaryVersion, common.BinaryBranch, common.BinaryBuildTime)
		os.Exit(0)
	}

	// read new hocon config
	sc, err := common.NewServerConfig(*configFile)
	if err != nil {
		panic(err)
	}
	server := xhttp.NewWebconfigServer(sc, false, nil)

	// setup logging
	logFile := server.XW_XconfServer.GetString("xconfwebconfig.log.file")
	if len(logFile) > 0 {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("ERROR opening file: %v", err)
			panic(err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: common.LOGGING_TIME_FORMAT,
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "timestamp",
		},
	})

	// default log level info
	logLevel := log.InfoLevel
	if parsed, err := log.ParseLevel(server.XW_XconfServer.GetString("xconfwebconfig.log.level")); err == nil {
		logLevel = parsed
	}
	log.SetLevel(logLevel)
	if server.XW_XconfServer.GetBoolean("xconfwebconfig.log.set_report_caller") {
		log.SetReportCaller(true)
	}

	// SAT token INIT
	xwhttp.InitSatTokenManager(server.XW_XconfServer)
	xwhttp.SetLocalSatToken(log.Fields{})

	// setup router
	router := server.XW_XconfServer.GetRouter(false)

	// register the notfound handler
	router.NotFoundHandler = http.HandlerFunc(server.XW_XconfServer.NotFoundHandler)

	tsr := adminapi.TrailingSlashRemover(router)

	if server.XW_XconfServer.MetricsEnabled() {
		router.Handle("/metrics", promhttp.Handler())
		metrics := xwhttp.NewMetrics()
		handler := server.XW_XconfServer.WebMetrics(metrics, tsr)
		server.XW_XconfServer.Handler = handler
	} else {
		server.XW_XconfServer.Handler = tsr
	}
	// setup xconf APIs and tables
	dataapi.XconfSetup(server.XW_XconfServer, router)
	adminapi.XconfSetup(server, router)

	// Exit gracefully on Ctrl+C etc.
	done := make(chan bool)

	// Catch the signal and set the channel
	quit := make(chan os.Signal, 1) // Buffered channel here to fix a go vet warning
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// goroutine 1 just waits for a kill signal
		sig := <-quit
		log.Errorf("caught the %+v signal, exiting", sig)

		// Send a message to the done channel, triggering server shutdown
		done <- true
	}()

	log.Infof("server is starting at %s, port %s", started, server.XW_XconfServer.Addr)
	go func() {
		// goroutine 2 is running the server
		if err := server.XW_XconfServer.ListenAndServe(); err != nil {
			log.Errorf("ListenAndServe Error %+v, exiting", err)
		}
		done <- true
	}()

	// Waiting for either a kill signal or a ListenAndServe error
	<-done

	// K8s has a default terminationGracePeriod as 30 seconds, app's wait period should be
	// slightly less. Using a wait period of 25 secs, should be enough to handle inflight reqs
	// may need to set this in config
	waitPeriod := time.Duration(25) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), waitPeriod)
	defer cancel()

	server.XW_XconfServer.SetKeepAlivesEnabled(false)
	// server.Shutdown will cause "Server closed" err for ListenAndServe, terminating goroutine #2 near line #99
	if err := server.XW_XconfServer.Shutdown(ctx); err != nil {
		log.Errorf("could not shutdown the web server: %+v\n", err)
	}
	log.Info("xconfadmin is shutdown")
}

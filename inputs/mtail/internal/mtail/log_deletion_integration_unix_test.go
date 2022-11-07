// Copyright 2020 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

//go:build unix
// +build unix

package mtail_test

import (
	"flashcat.cloud/categraf/inputs/mtail/internal/mtail"
	"flashcat.cloud/categraf/inputs/mtail/internal/testutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestLogDeletion is a unix-only test because on Windows files with open read handles cannot be deleted.
func TestLogDeletion(t *testing.T) {
	testutil.SkipIfShort(t)
	workdir := testutil.TestTempDir(t)

	// touch log file
	logFilepath := filepath.Join(workdir, "log")
	logFile := testutil.TestOpenFile(t, logFilepath)
	defer logFile.Close()

	m, stopM := mtail.TestStartServer(t, 1, mtail.LogPathPatterns(logFilepath))
	defer stopM()

	logCloseCheck := m.ExpectMapExpvarDeltaWithDeadline("log_closes_total", logFilepath, 1)
	logCountCheck := m.ExpectExpvarDeltaWithDeadline("log_count", -1)

	m.PollWatched(1) // Force sync to EOF
	log.Println("remove")
	testutil.FatalIfErr(t, os.Remove(logFilepath))

	m.PollWatched(0) // one pass to stop
	logCloseCheck()
	m.PollWatched(0) // one pass to remove completed stream
	logCountCheck()
}

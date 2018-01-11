// Copyright 2016 Klaus Birkelund Jensen <birkelund@nbi.ku.dk>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.
//
// Author: Klaus Birkelund Jensen <birkelund@nbi.ku.dk>

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"hpt.space/bltfs"
	filedebug "hpt.space/bltfs/backend/file"
	"hpt.space/bltfs/util/fsutil"
)

type severity int32

const (
	infoLog severity = iota
	warningLog
	errorLog
)

func perror(args ...interface{}) {
	fmt.Fprintln(os.Stderr, logfmt(errorLog, "", args))
}

func perrorf(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, logfmt(errorLog, format, args))
}

func pinfo(args ...interface{}) {
	fmt.Fprintln(os.Stderr, logfmt(infoLog, "", args))
}

func pinfof(format string, args ...interface{}) {
	fmt.Fprintln(os.Stdout, logfmt(infoLog, format, args))
}

func logfmt(sev severity, format string, args []interface{}) string {
	var buf bytes.Buffer

	switch sev {
	case infoLog:
		buf.WriteString("[\033[1;32m+\033[0m] ")
	case errorLog:
		buf.WriteString("[\033[1;31m-\033[0m] ")
	default:
		buf.WriteString("[\033[1m*\033[0m] ")
	}

	if format == "" {
		fmt.Fprint(&buf, args...)
	} else {
		fmt.Fprintf(&buf, format, args...)
	}

	return buf.String()
}

func reporter(report *bltfs.Report) {
	fmt.Printf("report obtained: %v\n", report)
	fmt.Printf("  %d of total %d bytes durable written\n", report.Durable(), report.Total())
	fmt.Printf("  files in-transfer:\n")

	for _, f := range report.InTransfer() {
		fmt.Printf("    %s\n", f)
	}

	fmt.Printf("  files durable stored:\n")

	for _, f := range report.Finished() {
		fmt.Printf("    %s\n", f)
	}

}

func main() {
	pinfo("bltfs test starting")

	// setup fixtures
	if err := fsutil.CopyDir("./fixtures/golden", "./tmp/tape"); err != nil {
		perrorf("failed to copy fixture directory: %v", err)
		os.Exit(1)
	}

	// clean up
	//defer os.RemoveAll("/tmp/bltfs/tape")

	pol := bltfs.RecoveryPolicy{
		FullIndexInterval: 1 * time.Hour,
		DifferentialAfter: 10632560640, // 10 GB
		IncrementalAfter:  1073741824,  // 1 GB
	}

	backend, err := filedebug.Open("./tmp/tape")
	if err != nil {
		perrorf("failed to open filedebug backend: %v", err)
		os.Exit(1)
	}

	// open bltfs store
	store, err := bltfs.Open(backend,
		// we'll use the file debug backend
		bltfs.WithFileDebug(),

		// set the recovery policy
		bltfs.WithRecoveryPolicy(pol),

		// please report whenever a file is durably stored
		bltfs.WithReporter(reporter),
	)

	if err != nil {
		perrorf("failed to open bLTFS store: %v", err)
		os.Exit(1)
	}

	// store a directory to the bltfs store
	src := "./fixtures/files"

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// process possible error first
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := store.Mkdir(path); err != nil {
				return errors.Wrapf(err, "failed to create directory '%s' in sotre", path)
			}

			return nil
		}

		// open local file
		f, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", path)
		}
		defer f.Close()

		// create the remote file
		basename := filepath.Base(path)
		rf, err := store.Create(basename)
		if err != nil {
			return errors.Wrapf(err, "failed to create file '%s' in store", path)
		}

		// copy the contents efficiently (use the block size of the device)
		written, err := store.Copy(rf, f)
		if err != nil {
			return errors.Wrapf(err, "failed to copy file '%s'", path)
		}

		pinfof("copied %d bytes from %s", written, path)

		// close the remote file
		if err := rf.Close(); err != nil {
			return errors.Wrapf(err, "failed to close remote file '%s'", rf)
		}

		return nil
	})

	if err != nil {
		perrorf("failed to walk directory '%s': %v", src, err)
		os.Exit(1)
	}

	// read some files
	_, err = store.Open("/foo/bar/baz/quux/foo/bar/baz/file1")
	if err != nil {
		perrorf("failed to open file: %v", err)
	}

	if err := store.Close(); err != nil {
		perrorf("failed to close backend store: %v", err)
		os.Exit(1)
	}
}

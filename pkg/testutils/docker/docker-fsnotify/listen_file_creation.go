package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
)

type result struct {
	finished bool
	err      error
}

const defaultTimeout = 30 * time.Second

func main() {
	__antithesis_instrumentation__.Notify(644149)
	if len(os.Args) < 2 {
		__antithesis_instrumentation__.Notify(644157)
		panic(errors.Wrap(
			fmt.Errorf("must provide the folder to watch and the file to listen to"),
			"fail to run fsnotify to listen to file creation"),
		)
	} else {
		__antithesis_instrumentation__.Notify(644158)
	}
	__antithesis_instrumentation__.Notify(644150)

	var err error

	folderPath := os.Args[1]
	wantedFileName := os.Args[2]

	timeout := defaultTimeout

	var timeoutVal int
	if len(os.Args) > 3 {
		__antithesis_instrumentation__.Notify(644159)
		timeoutVal, err = strconv.Atoi(os.Args[3])
		if err != nil {
			__antithesis_instrumentation__.Notify(644160)
			panic(errors.Wrap(err, "timeout argument must be an integer"))
		} else {
			__antithesis_instrumentation__.Notify(644161)
		}
	} else {
		__antithesis_instrumentation__.Notify(644162)
	}
	__antithesis_instrumentation__.Notify(644151)

	timeout = time.Duration(timeoutVal) * time.Second

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		__antithesis_instrumentation__.Notify(644163)
		panic(errors.Wrap(err, "cannot create new fsnotify file watcher"))
	} else {
		__antithesis_instrumentation__.Notify(644164)
	}
	__antithesis_instrumentation__.Notify(644152)
	defer func() {
		__antithesis_instrumentation__.Notify(644165)
		if err := watcher.Close(); err != nil {
			__antithesis_instrumentation__.Notify(644166)
			panic(errors.Wrap(err, "error closing the file watcher in docker-fsnotify"))
		} else {
			__antithesis_instrumentation__.Notify(644167)
		}
	}()
	__antithesis_instrumentation__.Notify(644153)

	done := make(chan result)

	go func() {
		__antithesis_instrumentation__.Notify(644168)
		for {
			__antithesis_instrumentation__.Notify(644169)
			if _, err := os.Stat(filepath.Join(folderPath, wantedFileName)); errors.Is(err, os.ErrNotExist) {
				__antithesis_instrumentation__.Notify(644171)
			} else {
				__antithesis_instrumentation__.Notify(644172)
				done <- result{finished: true, err: nil}
			}
			__antithesis_instrumentation__.Notify(644170)
			time.Sleep(time.Second * 1)
		}
	}()
	__antithesis_instrumentation__.Notify(644154)

	go func() {
		__antithesis_instrumentation__.Notify(644173)
		for {
			__antithesis_instrumentation__.Notify(644174)
			select {
			case event, ok := <-watcher.Events:
				__antithesis_instrumentation__.Notify(644175)
				if !ok {
					__antithesis_instrumentation__.Notify(644179)
					return
				} else {
					__antithesis_instrumentation__.Notify(644180)
				}
				__antithesis_instrumentation__.Notify(644176)
				fileName := event.Name[strings.LastIndex(event.Name, "/")+1:]
				if event.Op&fsnotify.Write == fsnotify.Write && func() bool {
					__antithesis_instrumentation__.Notify(644181)
					return fileName == wantedFileName == true
				}() == true {
					__antithesis_instrumentation__.Notify(644182)
					done <- result{finished: true, err: nil}
				} else {
					__antithesis_instrumentation__.Notify(644183)
				}
			case err, ok := <-watcher.Errors:
				__antithesis_instrumentation__.Notify(644177)
				if !ok {
					__antithesis_instrumentation__.Notify(644184)
					return
				} else {
					__antithesis_instrumentation__.Notify(644185)
				}
				__antithesis_instrumentation__.Notify(644178)
				done <- result{finished: false, err: err}
			}
		}
	}()
	__antithesis_instrumentation__.Notify(644155)

	err = watcher.Add(folderPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(644186)
		fmt.Printf("error: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(644187)
	}
	__antithesis_instrumentation__.Notify(644156)

	select {
	case res := <-done:
		__antithesis_instrumentation__.Notify(644188)
		if res.finished && func() bool {
			__antithesis_instrumentation__.Notify(644190)
			return res.err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(644191)
			fmt.Println("finished")
		} else {
			__antithesis_instrumentation__.Notify(644192)
			fmt.Printf("error in docker-fsnotify: %v", res.err)
		}

	case <-time.After(timeout):
		__antithesis_instrumentation__.Notify(644189)
		fmt.Printf("timeout for %s", timeout)
	}
}

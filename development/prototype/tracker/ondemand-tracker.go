package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"time"
)

var config struct {
	path  string
	delay time.Duration
}

var delayOpt uint

func init() {
	log.SetFlags(0)
	flag.StringVar(&config.path, "path", "", "path to log file dir")
	flag.UintVar(&delayOpt, "f", uint(100), "microsec delay between each log event")
}

//var publish chan interface{}

func main() {
	flag.Parse()
	config.delay = time.Duration(delayOpt) * time.Microsecond
	log.Printf("%d", delayOpt)
	log.Printf("%d", config.delay.Nanoseconds())
	if config.path == "" {
		log.Println("option -path is required.")
		flag.Usage()
		os.Exit(0)
	}

	log.Printf("trak %q", config.path)
	ctl, requests, reports := tracker()

	user := make(chan os.Signal, 1)
	signal.Notify(user, os.Interrupt, os.Kill)

	go track(*ctl, requests, reports, config.path, "*")

	// driver
	flag := true
	requests <- struct{}{}
	for flag {
		//		time.Sleep(time.Microsecond)
		select {
		case report := <-reports:
			for _, event := range report.events {
				log.Println(event.String())
			}
			// wait a bit before requesting next update
			time.Sleep(config.delay)
			requests <- struct{}{}
		case stat := <-ctl.stat:
			log.Printf("stat: %s", stat)
			flag = false
		case <-user:
			os.Exit(0)
		default:
		}
	}
	// driver

	log.Println("bye")
	os.Exit(0)
}

func tracker() (*control, chan struct{}, chan *trackreport) {
	r := make(chan struct{}, 1)
	c := make(chan *trackreport, 0)
	return procControl(), r, c
}

func procControl() *control {
	return &control{
		sig:  make(chan interface{}, 1),
		stat: make(chan interface{}, 1),
	}
}

type control struct {
	sig  chan interface{}
	stat chan interface{}
}

// ----------------------------------------------------------------------
// tracker task
// ----------------------------------------------------------------------
func track(ctl control, requests <-chan struct{}, out chan<- *trackreport, basepath string, pattern string) {
	defer recovery(ctl, "done")

	log.Println("traking..")

	// maintains snapshot view of tracker after each request - initially empty
	var snapshot map[string]os.FileInfo = make(map[string]os.FileInfo)

	for {
		select {
		case <-requests:

			file, e := os.Open(basepath)
			anomaly(e)
			filenames, e := file.Readdirnames(0)
			anomaly(e)
			e = file.Close()
			anomaly(e)

			workingset := make(map[string]os.FileInfo)

			var eventTime = time.Now()
			var eventType fileEventCode
			var events = make([]fileEvent, len(filenames)+len(snapshot))
			var eventNum = 0

			for _, basename := range filenames {
				// REVU: need os agnostic variant, not just for *nix
				if basename[0] == '.' {
					continue
				}

				filename := path.Join(basepath, basename)
				info, e := os.Stat(filename)
				if e != nil {
					// deleted under our nose
					// were we tracking it?
					if _, found := snapshot[basename]; found {
						events[eventNum] = fileEvent{eventTime, TrackEvent.DeletedFile, snapshot[basename]}
						eventNum++
						delete(snapshot, basename)
						continue
					}
				}

				workingset[basename] = info
				info0 := snapshot[basename]
				news := false
				// is it news?
				if info0 != nil {
					// compare
					if info0.Size() != info.Size() {
						// changed
						eventType = TrackEvent.ModifiedFile
						news = true
					}
				} else {
					eventType = TrackEvent.NewFile
					news = true
				}
				if news {
					events[eventNum] = fileEvent{eventTime, eventType, info}
					eventNum++
					snapshot[basename] = info
				}
			}
			// were we tracking anything that is no longer in the dir?
			for f, _ := range snapshot {
				if _, found := workingset[f]; !found {
					events[eventNum] = fileEvent{eventTime, TrackEvent.DeletedFile, snapshot[f]}
					eventNum++
					delete(snapshot, f)
				}
			}
			report := trackreport{basepath, events[:eventNum]}
			out <- &report
		}
	}
}

// ----------------------------------------------------------------------
// tracked file event
// ----------------------------------------------------------------------
type fileEventCode string

func (t fileEventCode) String() string { return string(t) }

// enum
var TrackEvent = struct {
	NewFile, ModifiedFile, DeletedFile fileEventCode
}{
	NewFile:      "TRK",
	ModifiedFile: "MOD",
	DeletedFile:  "DEL",
}

type fileEvent struct {
	timestamp_ns time.Time
	event        fileEventCode
	fileinfo     os.FileInfo
}

func (t *fileEvent) String() string {
	return fmt.Sprintf("%d %3s stat %s", t.timestamp_ns.UnixNano(), t.event.String(), fileStatString(t.fileinfo))
}

func fileStatString(f os.FileInfo) string {
	if f == nil {
		return "BUG - nil"
	}
	return fmt.Sprintf("%020d %s %020d %s", f.Size(), f.Mode(), f.ModTime().Unix(), f.Name())
}

// ----------------------------------------------------------------------
// tracker report
// ----------------------------------------------------------------------

// assumes track focuses on a specific (base)path
type trackreport struct {
	basepath string
	events   []fileEvent
}

// ----------------------------------------------------------------------
// temp | replace with anomaly.*
// ----------------------------------------------------------------------

func recovery(ctl control, ok interface{}) {
	log.Println("recovering error ..")
	p := recover()

	if p != nil {
		ctl.stat <- p
		return
	}
	ctl.stat <- ok
}
func anomaly(e error) {
	if e != nil {
		panic(e)
	}
}
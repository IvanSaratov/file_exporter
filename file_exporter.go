package main

import (
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	addr      = kingpin.Flag("listen", "Address on which to expose metrics").Short('l').Default(":9393").String()
	directory = kingpin.Flag("directory", "Path to directory for watch").Short('d').Default(".").String()

	directorySize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "file_exporter",
			Subsystem: "directory",
			Name:      "get_size",
			Help:      "Size of directory in bytes",
		},
	)
	directoryTime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "file_exporter",
			Subsystem: "directory",
			Name:      "get_last_update",
			Help:      "Get last update directory time in UNIX seconds",
		}, []string{"status"},
	)
)

func getDirectorySize(path string, info os.FileInfo) int64 {
	size := info.Size()
	if !info.IsDir() {
		return size
	}

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return size
	}
	defer dir.Close()

	fis, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	for _, fi := range fis {
		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}
		size += getDirectorySize(path+"/"+fi.Name(), fi)
	}

	return size
}

func setDirectorySize(path string) error {
	info, err := os.Lstat(*directory)
	if err != nil {
		return err
	}
	//Not optimize
	size := getDirectorySize(*directory, info)
	directorySize.Set(float64(size))

	return nil
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.Parse()

	if _, err := os.Stat(*directory); os.IsNotExist(err) {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Infof("Starting new directory watcher")

	//Set default size
	if err := setDirectorySize(*directory); err != nil {
		log.Fatal(err)
	}
	log.Infof("Set default directory size")

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if err := setDirectorySize(*directory); err != nil {
					return
				}
				directoryTime.With(prometheus.Labels{"status": event.Op.String()}).SetToCurrentTime()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Warnf("error: %v", err)
			}
		}
	}()

	err = watcher.Add(*directory)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Add %s directory to watcher", *directory)

	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Listen on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

package main

import (
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

type Conf struct {
	HttpListen    string `envconfig:"LISTEN_HTTP" default:"localhost:9393"`
	DirectoryName string `envconfig:"DIRECTORY_NAME"`
}

var (
	directorySize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "file_exporter",
			Subsystem: "directory",
			Name:      "get_directory_size",
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

func main() {
	var conf Conf
	if err := envconfig.Process("file_exporter", &conf); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(conf.DirectoryName); os.IsNotExist(err) {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Infof("Starting new directory watcher")

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				info, err := os.Lstat(conf.DirectoryName)
				if err != nil {
					return
				}
				//Not optimize
				size := getDirectorySize(conf.DirectoryName, info)
				directorySize.Set(float64(size))
				directoryTime.With(prometheus.Labels{"status": event.Op.String()}).SetToCurrentTime()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Warnf("error: %v", err)
			}
		}
	}()

	err = watcher.Add(conf.DirectoryName)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Listen on %s", conf.HttpListen)
	log.Fatal(http.ListenAndServe(conf.HttpListen, nil))
}

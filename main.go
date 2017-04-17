package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

// DefaultPort is the default TCP port to use for streaming
const DefaultPort = 2002

// BufferSize is the buffer size for reading / writing
const BufferSize = 8192

var monitorDir *string
var remoteHost *string
var remotePort *int
var remoteHostConn string

// StreamHeader is the JSON header sent (and terminated with \n) to the new streaming connection
type StreamHeader struct {
	FileName  string `json:"filename"`
	TimeStamp int64  `json:"timestamp"`
}

func streamFile(fi os.FileInfo) {
	log.Println("Streaming file:", fi.Name())
	f, err := os.Open(*monitorDir + "/" + fi.Name())
	if err != nil {
		log.Println(fi.Name(), err)
		return
	}
	defer f.Close()
	conn, err := net.Dial("tcp", remoteHostConn)
	if err != nil {
		log.Println(remoteHostConn, err)
		return
	}
	defer conn.Close()
	hdrBuf, _ := json.Marshal(StreamHeader{FileName: fi.Name(), TimeStamp: fi.ModTime().Unix()})
	n, err := conn.Write(hdrBuf)
	if n != len(hdrBuf) {
		log.Println("Sent partial header for", fi.Name())
		return
	}
	if err != nil {
		log.Println("Error sending file header", fi.Name())
		return
	}
	n, err = conn.Write([]byte{'\n'})
	if n != 1 || err != nil {
		log.Println("Error sending header delimiter")
		return
	}
	buf := make([]byte, BufferSize)
	lastReadTime := time.Now()
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			log.Println(fi.Name(), err)
			break
		}
		if n > 0 {
			lastReadTime = time.Now()
			written := 0
			for written < n {
				wn, err := conn.Write(buf[written:n])
				if err != nil {
					log.Println(fi.Name(), err)
					break
				}
				written += wn
			}
		} else {
			readSince := time.Since(lastReadTime)
			if readSince.Seconds() > 1 {
				log.Println("No new data on", fi.Name(), "since", lastReadTime, "-- closing")
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func main() {
	monitorDir = flag.String("dir", "", "Directory to monitor")
	remoteHost = flag.String("connect", "", "Remote host to connect to")
	remotePort = flag.Int("port", DefaultPort, "TCP port to use")
	flag.Parse()

	if *monitorDir == "" {
		log.Fatal("Expecting a directory to monitor")
	}

	if *remoteHost == "" {
		log.Fatal("Expecing a remote host to connect to")
	}

	remoteHostConn = *remoteHost + ":" + strconv.Itoa(*remotePort)

	oldFiles, err := ioutil.ReadDir(*monitorDir)
	if err != nil {
		log.Fatal("Error reading directory ", monitorDir, ":", err)
	}
	for {
		files, err := ioutil.ReadDir(*monitorDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, f2 := range files {
			found := false
			for _, f1 := range oldFiles {
				if f1.Name() == f2.Name() {
					found = true
					break
				}
			}
			if found {
				continue
			}
			go streamFile(f2)
		}
		oldFiles = files
		time.Sleep(50 * time.Millisecond)
	}
}

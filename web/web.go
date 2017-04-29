package web

import (
	"fmt"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

var mediaSequence int
var playlist atomic.Value

const (
	m3u8Header      = "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-ALLOW-CACHE:NO\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n"
	segmentTemplate = "#EXT-INF:2.00,\n/ts/seg-%d.ts\n"
	playlistLength  = 4
)

func init() {
	go live()
}

func live() {
	playlist.Store([][]byte{})
	for {
		p := playlist.Load().([][]byte)
		if len(p) > playlistLength {
			s := len(p) - playlistLength
			p = p[s:]
		}
		mediaSequence++
		ts := mediaSequence%playlistLength + 1
		p = append(p, []byte(fmt.Sprintf(segmentTemplate, ts)))
		if ts == playlistLength {
			p = append(p, []byte("#EXT-X-DISCONTINUITY\n"))
		}
		playlist.Store(p)
		time.Sleep(time.Second * 2)
	}
}

func New() *goji.Mux {
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/m3u8/playlist.m3u8"), func(w http.ResponseWriter, r *http.Request) {
		p := playlist.Load().([][]byte)
		m3u8 := []byte(fmt.Sprintf(m3u8Header, mediaSequence))
		for _, t := range p {
			m3u8 = append(m3u8, t...)
		}
		w.Header().Add("Content-Type", "application/x-mpegURL")
		w.Header().Add("Content-Length", strconv.Itoa(len(m3u8)))
		w.Write(m3u8)
	})

	mux.HandleFuncC(pat.Get("/ts/:name"), func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		p := pat.Param(ctx, "name")
		f, err := ioutil.ReadFile("data/" + p)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(f)))
		w.Header().Add("Content-Type", "application/x-mpegURL")
		w.Write(f)
	})

	return mux
}

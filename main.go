package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/netutil"
)

var (
	wait             time.Duration
	listenAddr       string
	maxConn          int
	logger           = log.New(os.Stdout, "", log.LstdFlags)
	disableAccessLog bool
)

// middleware is common function
func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !disableAccessLog {
			logger.Printf("id:%d %s %s %s %s", time.Now().UnixNano(), r.Method, r.RequestURI, r.RemoteAddr, r.UserAgent())
		}
		next.ServeHTTP(w, r)
	}
}

func main() {
	flag.DurationVar(&wait, "glaceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.IntVar(&maxConn, "maxconn", 100, "server limited listener connection")
	flag.StringVar(&listenAddr, "listen", ":8000", "server listen address")
	flag.BoolVar(&disableAccessLog, "disableaccesslog", false, "disable access log")
	flag.Parse()

	logger.Println("servre is starting ...")

	r := http.NewServeMux()
	r.HandleFunc("/", middleware(hello))
	r.HandleFunc("/healthz", healthz)

	srv := &http.Server{
		Addr:         listenAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalln(err)
	}

	limitedListener := netutil.LimitListener(ln, maxConn)

	// SIGINT (Ctrl + C) で正常なシャットダウンを行う
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		// タイムアウトを待ってから停止
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)

	}()

	logger.Println("server is ready to handle requests at", listenAddr)
	logger.Fatalln(srv.Serve(limitedListener))

}

// hello is handlerfunc
func hello(w http.ResponseWriter, r *http.Request) {
	// validate
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	time.Sleep(time.Millisecond * 500)
	dump, _ := httputil.DumpRequest(r, true)
	logger.Printf("Request dump\n%s", string(dump))
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"msg": "Hello world"}`)
}

// healthz is handlerfunc
func healthz(w http.ResponseWriter, r *http.Request) {
	// validate
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

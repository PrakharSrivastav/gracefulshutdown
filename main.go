package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type App struct {
	quit chan os.Signal
	done chan struct{}
	srv  *http.Server
}

func New(mux *http.ServeMux) *App {
	app := new(App)

	// setup server
	srv := &http.Server{
		Addr:    ":1234",
		Handler: mux,
	}

	app.srv = srv
	app.done = make(chan struct{})
	app.quit = make(chan os.Signal)
	signal.Notify(app.quit, os.Kill, os.Interrupt)
	return app
}

func main() {
	// setup application context
	ctx := context.Background()

	// setup handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ss", func(writer http.ResponseWriter, request *http.Request) {
		if _, err := fmt.Fprint(writer, "123"); err != nil {
			log.Print(err)
		}
	})

	app := New(mux)

	go app.Start()
	go app.HandleShutdownGracefully(ctx)
	go app.CheckGoNum()

	<-app.done
	log.Print("application shutdown")

}

func (app *App) CheckGoNum() {
	t := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-t.C:
			log.Printf("goroutine %+v", runtime.NumGoroutine())
			log.Printf("cpu %+v", runtime.NumCPU())
		}
	}
}

func (app *App) Start() {
	log.Print("starting server")
	if err := app.srv.ListenAndServe(); err != nil {
		log.Printf("error starting server :: %+v ", err)
	}
}

func (app *App) HandleShutdownGracefully(ctx context.Context) {
	<-app.quit
	log.Print("received shutdown signal")
	c, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := app.srv.Shutdown(c); err != nil {
		log.Printf("error shutting serve down :: %+v", err)
	}
	close(app.done)
}

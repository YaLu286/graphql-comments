package main

import (
	"graphql-comments/config"
	"graphql-comments/graph"
	"graphql-comments/migrations"
	"graphql-comments/storage"
	"syscall"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if cfg.StorageType == storage.StorageTypePostgres {
		err = migrations.RunDatabaseMigrations(cfg)
		if err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(store)}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: nil,
	}

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %s: %v\n", cfg.ServerPort, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGALRM)

	<-stop

	log.Println("уходим красиво, нажми Ctrl+C еще раз, если невтерпеж")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exiting")
}

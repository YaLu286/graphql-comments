package main

import (
	"graphql-comments/config"
	"graphql-comments/graph"
	"graphql-comments/migrations"
	"graphql-comments/storage"

	"log"
	"net/http"
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

	// queryCache := lru.New(100)

	// srv.SetQueryCache(queryCache)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

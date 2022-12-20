package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mondaydarknight/hpdrive/internal/infrastructure"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Printf("Error: %v", err)
		if e, ok := err.(*appError); ok {
			replyJSON(w, e, e.Code)
		} else {
			replyJSON(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		}
	}
}

// Register API endpoints to the router.
func SetupRoutes(r *mux.Router) {
	repo, err := infrastructure.NewFileRepository()
	if err != nil {
		log.Fatal(err)
	}
	c := &controller{repo}
	r.Methods("GET").PathPrefix("/file").Handler(appHandler(c.get))
	r.Methods("POST").PathPrefix("/file").Handler(appHandler(c.create))
	r.Methods("PATCH").PathPrefix("/file").Handler(appHandler(c.patch))
	r.Methods("DELETE").PathPrefix("/file").Handler(appHandler(c.delete))
}

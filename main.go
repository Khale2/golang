package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Item represents a simple resource in our API
type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var (
	items = make(map[string]Item)
	mu    sync.RWMutex
)

func main() {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/items", handleItems)
	mux.HandleFunc("/items/", handleItem) // Covers /items/{id}

	// Logging middleware
	loggedMux := loggingMiddleware(mux)

	port := ":8080"
	fmt.Printf("Starting simple API on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, loggedMux))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Printf("404 Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the Simple API for Wallarm Lab testing!",
		"status":  "running",
	})
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.RLock()
		defer mu.RUnlock()
		itemList := make([]Item, 0, len(items))
		for _, item := range items {
			itemList = append(itemList, item)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(itemList)

	case http.MethodPost:
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("%d", len(items)+1)
		}
		mu.Lock()
		items[item.ID] = item
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/items/"):]
	if id == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		mu.RLock()
		item, ok := items[id]
		mu.RUnlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		item.ID = id
		mu.Lock()
		items[id] = item
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)

	case http.MethodDelete:
		mu.Lock()
		delete(items, id)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

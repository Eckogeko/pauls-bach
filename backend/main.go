package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"pauls-bach/config"
	"pauls-bach/handlers"
	"pauls-bach/market"
	mw "pauls-bach/middleware"
	"pauls-bach/models"
	"pauls-bach/sse"
	"pauls-bach/store"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	s, err := store.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to init store: %v", err)
	}

	// Bootstrap admin account
	bootstrapAdmin(s, cfg.AdminPIN)

	engine := &market.Engine{Store: s}
	broker := sse.NewBroker(cfg.JWTSecret)

	authH := &handlers.AuthHandler{Store: s, JWTSecret: cfg.JWTSecret}
	eventH := &handlers.EventHandler{Store: s, Engine: engine}
	tradingH := &handlers.TradingHandler{Store: s, Engine: engine, Broker: broker}
	adminH := &handlers.AdminHandler{Store: s, Engine: engine, Broker: broker}
	leaderboardH := &handlers.LeaderboardHandler{Store: s}
	historyH := &handlers.HistoryHandler{Store: s}
	bingoH := &handlers.BingoHandler{Store: s}
	bingoAdminH := &handlers.BingoAdminHandler{Store: s, Broker: broker}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/register", authH.Register)
		r.Get("/stream", broker.ServeHTTP)

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.JWTSecret))
			r.Get("/auth/me", authH.Me)
			r.Get("/events", eventH.List)
			r.Get("/events/{id}", eventH.Get)
			r.Get("/events/{id}/odds-history", eventH.OddsHistory)
			r.Post("/events/{id}/buy", tradingH.Buy)
			r.Post("/events/{id}/sell", tradingH.Sell)
			r.Get("/leaderboard", leaderboardH.Get)
			r.Get("/users/{id}/history", historyH.Get)
			r.Get("/bingo/events", bingoH.ListBingoEvents)
			r.Get("/bingo/board", bingoH.GetBoard)
			r.Post("/bingo/board", bingoH.CreateBoard)
			r.Get("/bingo/winners", bingoH.ListWinners)
			r.Get("/bingo/boards", bingoH.ListBoards)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.JWTSecret))
			r.Use(mw.AdminOnly)
			r.Get("/admin/users", adminH.ListUsers)
			r.Post("/admin/users/{id}/bingo", adminH.SetBingo)
			r.Post("/admin/users/{id}/balance", adminH.SetBalance)
			r.Post("/admin/users/{id}/reset-bingo", adminH.ResetBingoBoard)
			r.Post("/admin/events", adminH.CreateEvent)
			r.Put("/admin/events/{id}", adminH.UpdateEvent)
			r.Delete("/admin/events/{id}", adminH.DeleteEvent)
			r.Post("/admin/events/{id}/resolve", adminH.ResolveEvent)
			r.Post("/admin/bingo/events", bingoAdminH.CreateBingoEvent)
			r.Put("/admin/bingo/events/{id}", bingoAdminH.UpdateBingoEvent)
			r.Post("/admin/bingo/events/{id}/resolve", bingoAdminH.ResolveBingoEvent)
	
		})
	})

	// Serve frontend in production
	distDir := cfg.FrontendDist
	if abs, err := filepath.Abs(distDir); err == nil {
		distDir = abs
	}
	if _, err := os.Stat(distDir); err == nil {
		fileServer := http.FileServer(http.Dir(distDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file; if not found serve index.html (SPA)
			path := filepath.Join(distDir, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	fmt.Printf("Server starting on :%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func bootstrapAdmin(s *store.Store, adminPIN string) {
	store.WriteLock()
	defer store.WriteUnlock()

	if admin, _ := s.Users.GetByUsername("admin"); admin != nil {
		return // already exists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPIN), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash admin pin: %v", err)
	}

	admin := &models.User{
		Username: "admin",
		PinHash:  string(hash),
		Balance:  0,
		IsAdmin:  true,
	}
	if err := s.Users.Create(admin); err != nil {
		log.Fatalf("failed to create admin: %v", err)
	}
	log.Println("Admin account created")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

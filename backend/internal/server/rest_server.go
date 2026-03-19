package server

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/snarf-dev/fsm/v2/internal/auth"
	"github.com/snarf-dev/fsm/v2/internal/config"
)

type RestServer struct {
	manager   *ServerManager
	fsmConfig *config.FSMConfig
}

func CreateRestServer(cfg *config.FSMConfig) *RestServer {
	server := RestServer{
		manager:   CreateManager(cfg),
		fsmConfig: cfg,
	}

	if len(cfg.Admins) == 0 {
		log.Println("No server admins, creating")
		envPass := strings.TrimSpace(os.Getenv("FSM_ADMIN_PASSWORD"))
		if envPass != "" {
			hashedSecret, _ := auth.HashPassword(envPass)
			cfg.Admins["admin"] = hashedSecret
			cfg.SaveToFile()
			log.Printf("admin user created. login with admin and your Unraid configured password")
		} else {
			if password, hashedPassword, err := auth.GenerateRandomPassword(8); err != nil {
				log.Panicf("unable to generate password: %v", err)
				os.Exit(1)
			} else {
				cfg.Admins["admin"] = hashedPassword
				cfg.SaveToFile()
				log.Printf("admin user created. login with admin/%s", password)
			}
		}
	}

	if cfg.Factorio.AutoStart {
		err := server.manager.Start()
		if err != nil {
			log.Printf("failed to start the server, %v\n", err)
		}
	}

	watchConfig(cfg.Path, func() {
		err, newCfg := config.Load(&cfg.Path)
		if err == nil {
			server.manager.cfg = newCfg
			server.fsmConfig = newCfg
			log.Println("Config reloaded")
		}
	})

	return &server
}

func (s *RestServer) Start() {
	r := mux.NewRouter()

	r.HandleFunc("/start", s.withAuth(s.startHandler)).Methods("GET")
	r.HandleFunc("/stop", s.withAuth(s.stopHandler)).Methods("GET")
	r.HandleFunc("/status", s.withAuth(s.statusHandler)).Methods("GET")
	r.HandleFunc("/mods", s.withAuth(s.modsHandler)).Methods("GET")
	r.HandleFunc("/mods/bookmarked", s.withAuth(s.bookmarkedModsHandler)).Methods("GET")
	r.HandleFunc("/mods/download/{mod}/{version}", s.withAuth(s.handleDownloadMod)).Methods("GET")
	r.HandleFunc("/mods/install/{mod}/{version}", s.withAuth(s.handleInstallMod)).Methods("PUT")
	r.HandleFunc("/mods/uninstall/{mod}/{version}", s.withAuth(s.handleUninstallMod)).Methods("DELETE")
	r.HandleFunc("/mods/{mod}/{version}", s.withAuth(s.handleDeleteMod)).Methods("DELETE")
	r.HandleFunc("/toggle-mod", s.withAuth(s.toggleModHandler)).Methods("POST")
	r.HandleFunc("/rcon", s.withAuth(s.rconHandler)).Methods("POST")
	r.HandleFunc("/ws/logs", s.handleLogStream)
	r.HandleFunc("/saves", s.withAuth(s.handleListSaves)).Methods("GET")
	r.HandleFunc("/saves/{name}", s.withAuth(s.handleDownloadSave)).Methods("GET")
	r.HandleFunc("/saves", s.withAuth(s.handleUploadSave)).Methods("POST")
	r.HandleFunc("/saves/{name}", s.withAuth(s.handleDeleteSave)).Methods("DELETE")
	r.HandleFunc("/settings", s.withAuth(s.handleGetSettings)).Methods("GET")
	r.HandleFunc("/settings/save", s.withAuth(s.handleUpdateSave)).Methods("POST")

	r.HandleFunc("/admins", s.withAuth(s.handleListAdmins)).Methods("GET")
	r.HandleFunc("/admins", s.withAuth(s.handleAddAdmin)).Methods("POST")
	r.HandleFunc("/admins/{user}", s.withAuth(s.handleUpdateAdmin)).Methods("POST")
	r.HandleFunc("/admins/{user}", s.withAuth(s.handleDeleteAdmin)).Methods("DELETE")

	r.HandleFunc("/factorio-admins", s.withAuth(s.handleListFactorioAdmins)).Methods("GET")
	r.HandleFunc("/factorio-admins", s.withAuth(s.handleAddFactorioAdmin)).Methods("POST")
	r.HandleFunc("/factorio-admins/{user}", s.withAuth(s.handleRemoveFactorioAdmin)).Methods("DELETE")

	r.HandleFunc("/factorio-bans", s.withAuth(s.handleListFactorioBans)).Methods("GET")
	r.HandleFunc("/factorio-bans", s.withAuth(s.handleAddFactorioBanUser)).Methods("POST")
	r.HandleFunc("/factorio-bans/{user}", s.withAuth(s.handleRemoveFactorioBanUser)).Methods("DELETE")

	r.HandleFunc("/factorio-whitelist", s.withAuth(s.handleListFactorioWhitelistUsers)).Methods("GET")
	r.HandleFunc("/factorio-whitelist", s.withAuth(s.handleAddFactorioWhitelistUser)).Methods("POST")
	r.HandleFunc("/factorio-whitelist/{user}", s.withAuth(s.handleRemoveFactorioWhitelistUser)).Methods("DELETE")

	r.HandleFunc("/factorio-settings", s.withAuth(s.handleGetServerSettings)).Methods("GET")
	r.HandleFunc("/factorio-settings", s.withAuth(s.handleUpdateServerSettings)).Methods("PUT")

	r.HandleFunc("/factorio-versions", s.withAuth(s.handleListFactorioVersions)).Methods("GET")
	r.HandleFunc("/factorio-versions/{branch}/{version}", s.withAuth(s.handleSelectFactorioVersion)).Methods("PUT")
	r.HandleFunc("/factorio-versions/{branch}/{version}", s.withAuth(s.handleUninstallFactorioVersion)).Methods("DELETE")
	r.HandleFunc("/factorio-versions/{branch}/{version}/download", s.withAuth(s.handleDownloadFactorioVersion)).Methods("GET")
	r.HandleFunc("/ws/download/{branch}/{version}", s.handleDownloadProgressStream).Methods("GET")

	r.HandleFunc("/factorio-user", s.withAuth(s.handleGetFactorioUserSettings)).Methods("GET")
	r.HandleFunc("/factorio-user", s.withAuth(s.handleUpdateFactorioUserSettings)).Methods("POST")

	fs := http.FileServer(http.Dir("./frontend/dist"))
	r.PathPrefix("/").Handler(fs)

	handler := cors.AllowAll().Handler(r)

	log.Printf("Server manager running at %s\n", s.fsmConfig.Server.Listen)
	http.ListenAndServe(s.fsmConfig.Server.Listen, handler)
}

func (s *RestServer) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || !auth.CheckPassword(s.fsmConfig.Admins[username], password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

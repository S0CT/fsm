# Factorio Server Manager

Factorio Server Manager (FSM) is a web-based application that allows you to manage a dedicated Factorio server with ease. It provides a reactive Vue 3 frontend and a Go backend to control server lifecycle, RCON commands, mod management, save games, server configuration, and version control — all from a modern browser interface.

> ⚠️ **Work in Progress**
>
> This project is actively being developed and is not yet stable.  
> Features may change or break at any time. 

---

## Features

- **Web-based Interface** — Manage your server from any device.
- **Start/Stop Server** — Control the server lifecycle with a single click.
- **Real-time Logs** — View the server logs as they stream in live.
- **Save Game Management** — Upload, download, delete, and switch between save games.
- **Mod Management** — Download, Install, Uninstall, Enable or disable installed mods through the UI.
- **Server Settings Editor** — Modify `server-settings.json` dynamically with tooltips and validation.
- **Admin Authentication** — Simple admin section using INI-based authentication with hashed passwords.
- **Version Management** — Download and switch between Factorio server versions from the official sources.
- **Auto-configuration** — Uses INI config with sensible defaults and support for hot-reload.

---

## Screenshots

[Screenshots](screenshots)

---

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js + npm (for building the frontend)
- [Factorio Account](https://www.factorio.com/profile) for downloading servers
- `xz-utils` installed (for extracting `.tar.xz` archives)

### Build and Run

```bash
# Build frontend
cd frontend
npm install
npm run build

# Build backend
cd ../
go build -o fsm
./fsm
```

Then open your browser to [http://localhost:8888](http://localhost:8888)

---

## Configuration

FSM loads configuration from `fsm.ini`, searched in the following locations (in order):

1. Supplied `--config` argument
2. `./fsm.ini`
3. `~/.config/fsm/fsm.ini`
4. `/etc/fsm/fsm.ini`

### Example `fsm.ini`

```ini
[factorio]
saves           = ./data/saves
mods            = ./data/mods
logs            = ./data/logs
config          = ./data/config
downloads       = ./data/downloads
server_versions = ./data/servers

[rcon]
bind        = 127.0.0.1:27015
password    = secret

[server]
listen = :8888
```

---

## Docker & Unraid
<!-- Unraid Optimization: Included support natively -->
FSM includes optimizations specifically for Unraid Docker environments. An `unraid-template.xml` is provided in the repository, and the Docker container is pre-configured to handle Unraid's default paths (`/data`, `/fsm.ini`) and WebUI port (`8888`).

### Building

`docker build -t fsm .`

### Running

#### Initial setup


* Use the setup script `curl -fsSL https://raw.githubusercontent.com/S0CT/fsm/refs/heads/main/setup-fsm.sh | bash`
* Edit docker-compose.yml if your port preferences are different
* Edit fsm.ini and change the RCon password the other settings shouldn't need to be changed and they can be changed in the UI.
* `docker compose up`
* On first run you will see the admin password which you can use to log in with 

---

## License

MIT

---

## Acknowledgements

- Inspired by Factorio's excellent headless server support.
- Uses [gorilla/websocket](https://github.com/gorilla/websocket) and `ini.v1`.
- UI built with Vue 3 + Tailwind CSS + PrimeVue.

## Contributing

## Contributing

> **Want to help improve this project?**
>
> Contributions are welcome! Whether it's reporting bugs, suggesting features, or submitting pull requests — every bit helps. 
>
> ### How to contribute:
> - Fork the repository
> - Create a new branch (`git checkout -b my-feature`)
> - Commit your changes
> - Push to your fork and open a Pull Request
>
> If you're unsure where to start, feel free to open an issue to discuss your ideas!

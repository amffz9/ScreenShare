# ScreenShare

A lightweight, self-contained screen sharing app over your local network — no accounts, no cloud, no installation required.

Broadcasters share their screen (or a capture device like an ATEM Mini Pro or webcam). Viewers connect from any browser on the same network and watch in real time via peer-to-peer WebRTC.

## How It Works

- Run the app on one machine — it acts as both the web server and the signaling server
- Open `/broadcaster` in your browser to start sharing your screen
- Viewers open `/viewer` on any device on the same network to watch
- Everything is peer-to-peer after the initial connection — video never goes through the server

---

## Installation

### Windows

1. Go to the [Releases](../../releases) page and download `ScreenShare-Windows.exe`
2. Double-click the file to run it

> **Windows Defender SmartScreen warning:** Because this app isn't code-signed, Windows may show a blue warning screen. Click **"More info"** then **"Run anyway"** to proceed.

3. A Windows Firewall prompt may appear — click **Allow** so viewers on your network can connect
4. Your browser will open automatically to the home page

To stop the server, close the terminal window that opened alongside it.

---

### Linux

1. Go to the [Releases](../../releases) page and download `ScreenShare-Linux`
2. Open a terminal and make it executable:

```bash
chmod +x ScreenShare-Linux
```

3. Run it:

```bash
./ScreenShare-Linux
```

4. Your browser should open automatically. If it doesn't, open `http://localhost:8080` manually.

**Firewall note:** If viewers on your network can't connect, allow the port through your firewall:

```bash
sudo ufw allow 8080/tcp
```

---

### macOS

1. Go to the [Releases](../../releases) page and download the correct file:
   - **Intel Mac:** `ScreenShare-Mac-Intel`
   - **Apple Silicon (M1/M2/M3):** `ScreenShare-Mac-AppleSilicon`
2. Open Terminal and make it executable:

```bash
chmod +x ScreenShare-Mac-Intel   # or ScreenShare-Mac-AppleSilicon
```

3. Run it:

```bash
./ScreenShare-Mac-Intel
```

> **macOS Gatekeeper warning:** On first run, macOS may block the app. Go to **System Settings → Privacy & Security**, scroll down, and click **"Open Anyway"** next to the ScreenShare entry.

---

## Quick Start

Once the server is running:

| URL | Purpose |
|-----|---------|
| `http://localhost:8080` | Home page |
| `http://localhost:8080/broadcaster` | Start sharing your screen |
| `http://localhost:8080/viewer` | Watch a broadcast |

Viewers on your network can connect using your machine's local IP address (shown in the terminal output), e.g. `http://192.168.1.100:8080/viewer`. A QR code is shown on the home page for easy mobile access.

---

## Configuration (Optional)

Place a `config.json` file in the same folder as the executable to customize settings:

```json
{
  "port": 8080,
  "stunServer": "stun:stun.l.google.com:19302"
}
```

You can also use command-line flags:

```
--port 9090           Use a different port
--no-open             Don't open the browser automatically
--stun stun:...       Use a custom STUN server
```

---

## Browser Compatibility

| Feature | Chrome | Edge | Firefox | Safari |
|---------|--------|------|---------|--------|
| Screen sharing | Yes | Yes | Yes | Yes |
| Capture devices (webcam/HDMI) | Yes | Yes | Yes | Limited |
| System audio | Yes | Yes | No | No |

**Recommended:** Google Chrome or Microsoft Edge for the best experience.

---

## Building from Source

Requires [Go 1.22+](https://go.dev/dl/).

```bash
git clone https://github.com/amffz9/screenshare.git
cd screenshare
make build        # Build for your current platform
make build-all    # Cross-compile for Windows, Linux, macOS
make run          # Run directly with go run (no build step)
```

Binaries are written to `dist/`.

---

## Troubleshooting

See [USER_GUIDE.md](USER_GUIDE.md) for detailed troubleshooting, configuration options, and capture device setup (ATEM Mini Pro, Elgato, webcams).

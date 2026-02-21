# WebRTC Screen Sharing - User Guide

## Getting Started

### Running the Server

Double-click the executable for your platform, or run it from a terminal:

```
./webrtc-screenshare            # Linux / macOS
webrtc-screenshare.exe          # Windows
```

The server will start and automatically open your browser to the home page. You'll see output like:

```
No config file found at config.json, using defaults (port=8080, stun=stun:stun.l.google.com:19302)
Server running at http://localhost:8080
Local Network Access: http://192.168.1.100:8080
  Broadcaster: /broadcaster
  Viewer:      /viewer
```

To prevent the browser from opening automatically, use:

```
./webrtc-screenshare --no-open
```

### Home Page

The home page (`http://localhost:8080`) shows two options:

- **Start Broadcasting** - Share your screen or capture device
- **Join as Viewer** - Watch a broadcast

A QR code is also displayed so viewers can quickly open the viewer page on their phone.

---

## Broadcasting

### Screen Share Mode

1. Go to the **Broadcaster** page
2. Enter a name for your broadcast (e.g., "Conference Room A")
3. Leave **Source** set to **Screen Share**
4. Check **Include audio** if you want to share system audio (browser support varies)
5. Click **Start Broadcasting**
6. Your browser will ask you to choose a screen, window, or tab to share
7. Once sharing, you'll see a preview and a list of connected viewers

### Capture Device Mode (ATEM Mini Pro, HDMI Capture Cards, Webcams)

This mode lets you broadcast from hardware capture devices like the Blackmagic ATEM Mini Pro, Elgato capture cards, or USB webcams.

1. **Connect your device** to the computer via USB/HDMI before starting
2. Go to the **Broadcaster** page
3. Enter a name for your broadcast
4. Change **Source** to **Capture Device (HDMI/USB)**
5. Your browser will ask for permission to access camera/microphone devices - **allow this**
6. Select your **Video Device** from the dropdown (e.g., "Blackmagic Design" or "USB Video")
7. Select your **Audio Device** from the dropdown, or leave it as "No Audio"
   - The ATEM Mini Pro typically appears as both a video and audio device
   - If your device embeds audio in the HDMI signal, select the matching audio input
8. Check **Include audio** to send audio to viewers
9. Click **Start Broadcasting**

**Tip:** If your device doesn't appear in the dropdown, make sure it's connected and recognized by your operating system first. Check your OS device manager/settings.

### Stopping a Broadcast

Click **Stop Broadcasting**, or close the browser tab. All connected viewers will be notified.

---

## Viewing

1. Open the **Viewer** page on any device on the same network
   - Navigate to `http://<server-ip>:8080/viewer`
   - Or scan the QR code from the home page
2. The page auto-connects to the signaling server
3. You'll see a list of active broadcasters - click **Connect** on one
4. The video (and audio, if available) will begin playing
5. Use **Toggle Fullscreen** for a full-screen view
6. Click **Disconnect** to stop watching

If a broadcaster disconnects, you'll be notified and can connect to another.

---

## Configuration

### config.json

Place a `config.json` file in the same directory as the executable to customize settings:

```json
{
  "port": 8080,
  "stunServer": "stun:stun.l.google.com:19302"
}
```

| Setting      | Default                          | Description                       |
|-------------|----------------------------------|-----------------------------------|
| `port`      | `8080`                           | HTTP/WebSocket server port        |
| `stunServer`| `stun:stun.l.google.com:19302`   | STUN server for WebRTC NAT traversal |

**The config file is optional.** If it's missing, the server uses the defaults shown above.

### Command-Line Flags

Flags override both defaults and config.json values:

```
--port 9090          # Use port 9090
--stun stun:stun.example.com:3478   # Custom STUN server
--config myconfig.json               # Use a different config file path
--no-open                            # Don't open browser on startup
```

### Priority Order

Settings are applied in this order (later overrides earlier):

1. Built-in defaults
2. `config.json` values
3. Command-line flags

---

## Network Requirements

- The **broadcaster and all viewers must be on the same local network** (or have network connectivity to each other)
- The server must be reachable by all clients on the configured port (default `8080`)
- **Firewall:** Ensure the server port is open for inbound TCP connections
  - Windows: You may see a Windows Firewall prompt on first run - allow it
  - Linux: `sudo ufw allow 8080/tcp` (if using ufw)
  - macOS: Allow the app in System Settings > Privacy & Security > Firewall

### STUN Server

The STUN server helps peers discover their network addresses for WebRTC connections. The default Google STUN server works for most setups. For fully offline/air-gapped networks, you can run your own STUN server (e.g., coturn) and point to it via config.

---

## Troubleshooting

### Broadcaster Issues

| Problem | Solution |
|---------|----------|
| "No video capture devices found" | Make sure your capture device is plugged in and recognized by the OS |
| "Device is in use by another application" | Close other apps using the device (OBS, Zoom, etc.) |
| "Screen sharing cancelled" | You need to select a screen/window when prompted |
| No audio from capture device | Select the correct audio input device in the dropdown; check that "Include audio" is checked |
| ATEM Mini Pro not listed | Ensure it's connected via USB and shows up in your OS device settings. Try a different USB port. On macOS, check System Settings > Privacy > Camera permissions for your browser |

### Viewer Issues

| Problem | Solution |
|---------|----------|
| "No active broadcasters" | Make sure a broadcaster is running and connected to the same server |
| Video plays but no audio | The broadcaster may not have included audio. Ask them to enable it |
| "Connection lost" | The broadcaster may have stopped. Try refreshing and reconnecting |
| Can't reach the server | Verify you're on the same network and the server IP/port are correct |

### General

| Problem | Solution |
|---------|----------|
| Browser shows permission errors | Use Chrome or Edge for best WebRTC and capture device support |
| Server won't start | Check if another application is using the same port. Try `--port 9090` |
| QR code doesn't work | Make sure your phone is on the same Wi-Fi network as the server |

---

## Browser Compatibility

| Feature | Chrome | Edge | Firefox | Safari |
|---------|--------|------|---------|--------|
| Screen sharing | Yes | Yes | Yes | Yes |
| Capture devices | Yes | Yes | Yes | Limited |
| System audio (screen share) | Yes | Yes | No | No |
| Cast to TV | Yes | Yes | No | No |

**Recommended:** Use Google Chrome or Microsoft Edge for the best experience, especially when using capture devices.

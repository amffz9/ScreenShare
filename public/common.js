// Common utilities
function escapeHtml(str) {
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
}

function showStatus(message, type, elementId = 'status') {
    const el = document.getElementById(elementId);
    if (el) {
        el.textContent = message;
        el.className = 'status ' + type;
        el.style.display = 'block';
    }
}

// Auto-detect WebSocket URL
function getWebSocketUrl() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}`;
}

// Fetch server IP and display it
function fetchAndDisplayIp(elementId = 'serverIp') {
    const el = document.getElementById(elementId);
    if (!el) return;

    fetch('/ip')
        .then(res => res.json())
        .then(data => {
            el.textContent = data.ip;
        })
        .catch(err => {
            console.error('Failed to fetch IP:', err);
            el.textContent = 'Unknown';
        });
}

// Fetch RTC config from server
async function fetchRtcConfig() {
    try {
        const res = await fetch('/config');
        const data = await res.json();
        return { iceServers: [{ urls: data.stunServer }] };
    } catch (err) {
        console.error('Failed to fetch config, using default STUN:', err);
        return { iceServers: [{ urls: 'stun:stun.l.google.com:19302' }] };
    }
}

// Presentation API / Cast Logic
function setupCastButton(buttonId, urlToCast) {
    const castBtn = document.getElementById(buttonId);
    if (!castBtn || !window.PresentationRequest) return;

    // Make button visible since API is supported
    castBtn.style.display = 'inline-block';

    const presentationRequest = new PresentationRequest([urlToCast]);

    castBtn.addEventListener('click', () => {
        presentationRequest.start()
            .then(connection => {
                console.log('Presentation connection started');
                showStatus('Casting to TV...', 'success');

                connection.onterminate = () => {
                    showStatus('Cast connection terminated', 'info');
                };
            })
            .catch(error => {
                console.error('Cast failed:', error);
                if (error.name === 'NotAllowedError') {
                    showStatus('Cast cancelled by user', 'info');
                } else {
                    showStatus('Cast failed: ' + error.message, 'error');
                }
            });
    });
}

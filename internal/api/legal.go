package api

import "net/http"

const legalStyle = `<style>
body{max-width:640px;margin:3rem auto;padding:0 1.5rem;font:16px/1.6 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#1a1a1a}
h1{font-size:1.5rem}
h2{font-size:1.1rem;margin-top:2rem}
a{color:#4f46e5}
</style>`

const termsHTML = `<!doctype html><html><head><meta charset="utf-8"><title>Juno &mdash; Terms of Service</title>` + legalStyle + `</head><body>
<h1>Juno &mdash; Terms of Service</h1>
<p>Last updated: August 1, 2026</p>

<h2>Overview</h2>
<p>Juno is a personal Discord bot providing two features: a public, Lanyard-compatible presence feed for a single Discord user, and join-to-create temporary voice channels within one Discord server. It is not a general-purpose service and is not intended for use outside the server it is deployed in.</p>

<h2>Acceptable use</h2>
<p>You may not use Juno's API to abuse, excessively scrape, or attempt to disrupt the service. Access may be revoked at the operator's discretion.</p>

<h2>No warranty</h2>
<p>Juno is provided "as is" without warranty of any kind. Uptime, accuracy of presence data, and availability of the temporary voice channel feature are not guaranteed.</p>

<h2>Changes</h2>
<p>These terms may change at any time. Continued use after changes constitutes acceptance.</p>

<h2>Contact</h2>
<p><a href="mailto:noc@ok8.sh">noc@ok8.sh</a></p>
</body></html>`

const privacyHTML = `<!doctype html><html><head><meta charset="utf-8"><title>Juno &mdash; Privacy Policy</title>` + legalStyle + `</head><body>
<h1>Juno &mdash; Privacy Policy</h1>
<p>Last updated: August 1, 2026</p>

<h2>What data Juno collects</h2>
<p><strong>Presence data:</strong> using the Presence and Server Members privileged Discord intents, Juno receives real-time status and activity updates (online status, activities, Spotify listening status) for one specific Discord user ID configured by the operator. This data is cached in memory only, is never written to disk, and is discarded on restart.</p>
<p><strong>Guild member events:</strong> Juno receives standard guild member join/update/leave events, as required by the Server Members intent, but does not persist member data beyond what's needed to track the presence-tracked user.</p>
<p><strong>Voice state events:</strong> Juno receives voice state updates to detect when a member joins the configured "hub" voice channel, so it can create a temporary voice channel and move that member into it.</p>
<p><strong>Temporary voice channel records:</strong> the Discord channel IDs and empty-since timestamps of channels Juno creates are stored in an embedded SQLite database, so cleanup of empty channels survives restarts. A record is deleted once its channel is deleted.</p>
<p>Juno does not read, store, or process message content.</p>

<h2>How data is used</h2>
<p>Presence data is served publicly via a REST endpoint (<code>GET /api/users/me</code>) and a WebSocket feed (<code>GET /api/socket</code>) &mdash; this is by design, as the presence feature is a public status feed rather than private data. Temporary voice channel records are used solely for channel lifecycle management and are not exposed via any API.</p>

<h2>Data sharing</h2>
<p>Juno does not sell or share collected data with third parties. Data is only exchanged with Discord's API as necessary to provide the bot's features.</p>

<h2>Data retention &amp; deletion</h2>
<p>Presence data is kept in memory only and cleared on restart or when the tracked user's presence changes. Temporary voice channel records are deleted when their corresponding channel is deleted. To request removal of your data &mdash; for example, if you were moved into a temporary voice channel &mdash; contact us below.</p>

<h2>Changes</h2>
<p>This policy may be updated; the "Last updated" date above reflects the most recent revision.</p>

<h2>Contact</h2>
<p><a href="mailto:noc@ok8.sh">noc@ok8.sh</a></p>
</body></html>`

func termsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(termsHTML))
}

func privacyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(privacyHTML))
}

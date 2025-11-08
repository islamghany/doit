// Package middlewares provides HTTP middleware functions for the API.
package middlewares

import (
	"net/http"

	"doit/internal/web"
)

// SecurityHeaders adds security-related HTTP headers to all responses
// Implements protections against OWASP A05:2021 (Security Misconfiguration)
func SecurityHeaders() web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		h := func(w http.ResponseWriter, r *http.Request) error {
			// Content Security Policy
			// Restricts sources from which content can be loaded
			// Prevents XSS and data injection attacks
			// Attacker injects malicious script into pages viewed by other users. Scripts run in victim’s browser with the page’s privileges (cookies/localStorage/access to DOM).
			// w.Header().Set("Content-Security-Policy",
			// 	"default-src 'self'; "+
			// 		"script-src 'self'; "+
			// 		"style-src 'self' 'unsafe-inline'; "+
			// 		"img-src 'self' data: https:; "+
			// 		"font-src 'self'; "+
			// 		"connect-src 'self'; "+
			// 		"frame-ancestors 'none'; "+
			// 		"base-uri 'self'; "+
			// 		"form-action 'self'")

			// Prevent clickjacking attacks
			// DENY: page cannot be displayed in a frame, regardless of origin
			// Attacker embeds a transparent or disguised iframe of your site into their page so the user unknowingly clicks UI elements on your site (e.g., “Approve”, “Buy”, or “Transfer”).
			w.Header().Set("X-Frame-Options", "DENY")

			// Prevent MIME type sniffing
			// Forces browsers to respect the Content-Type header
			// Prevents attackers from forcing browsers to interpret content as a different MIME type than specified in the Content-Type header.
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Referrer Policy
			// Controls how much referrer information is sent with requests
			// When a browser sends the full URL of the current page as the Referer header to other sites, it may accidentally leak sensitive data (tokens, private paths, query parameters).
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy (formerly Feature-Policy)
			// Disables browser features that aren't needed
			// Prevents attackers from using browser features to gain unauthorized access to the user's device or data.
			w.Header().Set("Permissions-Policy",
				"geolocation=(), "+
					"microphone=(), "+
					"camera=(), "+
					"payment=(), "+
					"usb=(), "+
					"magnetometer=(), "+
					"gyroscope=(), "+
					"accelerometer=()")

			// Enable XSS protection in legacy browsers
			// Modern browsers use CSP instead, but this provides defense in depth
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Strict Transport Security (HSTS)
			// Forces HTTPS connections for 1 year (31536000 seconds)
			// includeSubDomains: applies to all subdomains
			// preload: eligible for browser HSTS preload lists
			w.Header().Set("Strict-Transport-Security",
				"max-age=31536000; includeSubDomains; preload")

			// Remove server identification headers
			// Don't leak server implementation details to attackers
			w.Header().Del("Server")
			w.Header().Del("X-Powered-By")

			return handler(w, r)
		}
		return h
	}
}

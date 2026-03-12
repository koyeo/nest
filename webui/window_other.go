//go:build !darwin

package webui

// openWebview is a no-op on non-darwin platforms.
// The server.go code already falls back to openBrowser for non-macOS.
func openWebview(url, title string) {
	// no-op: webview_go is only available on macOS
}

package webui

import (
	webview "github.com/webview/webview_go"
)

// openWebview opens a native webview window pointing to the given URL.
// This is called on macOS where WebKit is built-in.
// The webview window blocks until closed, which signals server to shut down.
func openWebview(url, title string) {
	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Nest: " + title)
	w.SetSize(1000, 650, webview.HintNone)
	w.Navigate(url)
	w.Run()
}

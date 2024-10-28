# gocaptcha

A simple captcha service written in Go. It generates a random captcha image and stores it in memory for a configurable amount of time. The image is then served as a PNG file.

The captcha image is generated using the [comic.ttf](https://github.com/googlefonts/noto-emoji/blob/main/fonts/NotoColorEmoji.ttf) font.

Verfication is done by checking if the captcha image is stored in memory. If it is, the captcha is considered valid.

Set envvar ALLOWED_ORIGINS if you want to call `/getcaptcha` from browser. If you won't set it, header `Access-Control-Allow-Origin` will be set to `https://mozilla.org`.

## Deploy to Google Cloud Functions

```bash
echo '{ "ALLOWED_ORIGINS": "https://example.com,https://anotherdomain.com" }' > env.json

gcloud functions deploy CaptchaFunction \
--runtime go122 \
--trigger-http \
--env-vars-file env.json \
--allow-unauthenticated \
--region europe-north1 \
--entry-point CaptchaHandler
```

## Usage

### As a library

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/t0mk/gocaptcha"
)

func main() {
	// Set up routes
	http.HandleFunc("/getcaptcha", gocaptcha.CaptchaHandler)
	http.HandleFunc("/verify", gocaptcha.CaptchaHandler)

	// Start the HTTP server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
```

### As a CLI

```bash
go run cli/main.go
```

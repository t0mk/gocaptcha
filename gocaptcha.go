package gocaptcha

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/freetype"
)

// Embed the font file
//
//go:embed comic.ttf
var fontFile embed.FS

var (
	captchaMap      sync.Map
	captchaTTL      = time.Hour
	captchaWidth    = 230
	captchaHeight   = 60
	captchaFontSize = 36.
	allowedOrigins  = map[string]struct{}{}
)

func generateCode(length int) string {
	// no capital O and zero
	characters := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz123456789"
	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		sb.WriteByte(characters[rand.Intn(len(characters))])
	}
	return sb.String()
}

func createCaptchaImage(code string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw random lines for obfuscation
	for i := 0; i < 5; i++ {
		x1 := rand.Intn(captchaWidth)
		y1 := rand.Intn(captchaHeight)
		x2 := rand.Intn(captchaWidth)
		y2 := rand.Intn(captchaHeight)
		// Use a darker color for better visibility against a white background
		drawLine(img, x1, y1, x2, y2, color.RGBA{0, 0, 0, 255})
	}

	fontBytes, err := fontFile.ReadFile("comic.ttf")
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(captchaFontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black)

	pt := freetype.Pt(10, 5+int(c.PointToFixed(captchaFontSize)>>6))
	if _, err := c.DrawString(code, pt); err != nil {
		return nil, err
	}

	// Add noise artifacts
	addNoise(img, 100)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func addNoise(img *image.RGBA, numNoise int) {
	for i := 0; i < numNoise; i++ {
		x := rand.Intn(captchaWidth)
		y := rand.Intn(captchaHeight)
		img.Set(x, y, color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255})
	}
}

func init() {
	// Initialize the allowed origins map
	originsRaw := os.Getenv("ALLOWED_ORIGINS")
	if originsRaw != "" {
		fmt.Println("Allowed origins:", originsRaw)
		allowedOrigins = map[string]struct{}{}
		allowedOriginsSlice := strings.Split(originsRaw, ",")
		for _, origin := range allowedOriginsSlice {
			allowedOrigins[origin] = struct{}{}
		}
	}
}

func CaptchaHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/getcaptcha" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}

		code := generateCode(6)

		imgBytes, err := createCaptchaImage(code)
		if err != nil {
			http.Error(w, "Failed to create captcha image", http.StatusInternalServerError)
			return
		}

		captchaMap.Store(code, struct{}{})
		time.AfterFunc(captchaTTL, func() {
			captchaMap.Delete(code)
		})

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(imgBytes)
	} else if path == "/verify" {
		code := r.URL.Query().Get("code")
		response := map[string]bool{"valid": false}
		w.Header().Set("Content-Type", "application/json")
		if code != "" {
			w.WriteHeader(http.StatusOK)
			response["valid"] = false
			if _, exists := captchaMap.Load(code); exists {
				captchaMap.Delete(code)
				response["valid"] = true
			} 
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	} else {
		http.NotFound(w, r)
	}
}

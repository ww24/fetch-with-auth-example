package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

//go:embed web
var web embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: newHandler(),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatalf("failed to listen and serve: %+v", err)
		}
	}()

	log.Println("server started on port", port)
	<-ctx.Done()
	log.Println("shutting down...")

	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown: err=%+v", err)
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	selector := newSelector(generateImage(true), generateImage(false))
	mux.HandleFunc("/restricted/image.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin, Authorization")
		sameOrigin := true
		if origin := r.Header.Get("Origin"); origin != "" {
			sameOrigin = false
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "HEAD, GET")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Accept, Vary")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		switch r.Method {
		case http.MethodHead, http.MethodGet: // pass
			log.Println("request", r.Method)
		case http.MethodOptions:
			log.Println("preflight request")
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		key := strings.TrimPrefix(auth, "Bearer ")
		if key == "" {
			http.Error(w, "no api key", http.StatusUnauthorized)
			return
		}
		if !compareAPIKey(key) {
			http.Error(w, "invalid api key", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		if err := jpeg.Encode(w, selector(sameOrigin), &jpeg.Options{Quality: 95}); err != nil {
			log.Printf("failed to encode image: %+v", err)
		}
	})

	webFS, err := fs.Sub(web, "web")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))
	return mux
}

var compareAPIKey = func() func(string) bool {
	const apiKeyHashStr = "7d99d70818ebafa2e46ab347542347ad2f0cebba4bff84464cfda4f564de8cec"
	apiKeyHash, err := hex.DecodeString(apiKeyHashStr)
	if err != nil {
		panic(err)
	}
	return func(in string) bool {
		h := sha256.New()
		h.Write([]byte(in))
		return subtle.ConstantTimeCompare(h.Sum(nil), apiKeyHash) != 0
	}
}()

func newSelector[T any](a, b T) func(bool) T {
	return func(flag bool) T {
		if flag {
			return a
		}
		return b
	}
}

func generateImage(sameOrigin bool) image.Image {
	bg := color.RGBA{0xee, 0x55, 0x44, 0xff}
	textCol := color.RGBA{0xff, 0xff, 0xff, 0xff}
	rect := image.Rect(0, 0, 480, 270)
	img := image.NewRGBA(rect)
	draw.Draw(img, rect, image.NewUniform(bg), image.Point{0, 0}, draw.Src)
	face := inconsolata.Bold8x16
	const text1 = "RESTRICTED IMAGE"
	var text2 = "(same origin)"
	if !sameOrigin {
		text2 = "(cross origin)"
	}
	adv := font.MeasureString(face, text1)
	height := (rect.Dy() + face.Height - int(float64(face.Height)*2)) / 2
	drawFont(img, (rect.Dx()-adv.Round())/2, height, text1, textCol, face)
	adv = font.MeasureString(face, text2)
	drawFont(img, (rect.Dx()-adv.Round())/2, height+int(float64(face.Height)*1.5), text2, textCol, face)
	return img
}

func drawFont(img draw.Image, x, y int, text string, color color.Color, face font.Face) {
	drawer := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(text)
}

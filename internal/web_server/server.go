package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"stego/internal/stego_loader"
	"strconv"
)

// StegoServer encapsulates the steganography web server.
type StegoServer struct {
	FrontendPath string
	ListenAddr   string
}

// NewStegoServer creates a new instance of the web server.
func NewStegoServer(frontendPath, listenAddr string) *StegoServer {
	return &StegoServer{
		FrontendPath: frontendPath,
		ListenAddr:   listenAddr,
	}
}

// ErrorResponse represents a structured JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// DecodeResponse represents a successful decoding response.
type DecodeResponse struct {
	Message string `json:"message"`
}

// Start registers handlers and starts the HTTP server.
func (s *StegoServer) Start() error {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/", http.FileServer(http.Dir(s.FrontendPath)))

	// API routes
	mux.HandleFunc("/api/encode", s.handleEncode)
	mux.HandleFunc("/api/decode", s.handleDecode)

	fmt.Printf("Server starting on %s...\n", s.ListenAddr)
	return http.ListenAndServe(s.ListenAddr, mux)
}

func (s *StegoServer) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (s *StegoServer) handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.sendError(w, "File too large (limit 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		s.sendError(w, "No image provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	message := r.FormValue("message")
	pass := r.FormValue("pass")
	robustStr := r.FormValue("robust")
	robust, err := strconv.Atoi(robustStr)
	if err != nil || robust == 0 {
		robust = 20
	}

	if message == "" || pass == "" {
		s.sendError(w, "Missing required parameters (message, pass)", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "stego-web-")
	if err != nil {
		s.sendError(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, header.Filename)
	outPath := filepath.Join(tmpDir, "encoded_"+header.Filename)

	f, err := os.Create(inputPath)
	if err != nil {
		s.sendError(w, "Server error", http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
	f.Close()

	msgLen := len(message) * 32
	err = stego_loader.EncodeFile(inputPath, outPath, message, pass, msgLen, robust, -robust)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Encoding failed: %v", err), http.StatusInternalServerError)
		return
	}

	encodedFile, err := os.Open(outPath)
	if err != nil {
		s.sendError(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer encodedFile.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(outPath))
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, encodedFile)
}

func (s *StegoServer) handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.sendError(w, "File too large (limit 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		s.sendError(w, "No image provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	pass := r.FormValue("pass")
	msgLenStr := r.FormValue("msg_len")
	msgLen, err := strconv.Atoi(msgLenStr)
	if err != nil || msgLen <= 0 {
		s.sendError(w, "Invalid message length", http.StatusBadRequest)
		return
	}

	if pass == "" {
		s.sendError(w, "Missing password", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "stego-web-decode-")
	if err != nil {
		s.sendError(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, header.Filename)
	f, err := os.Create(inputPath)
	if err != nil {
		s.sendError(w, "Server error", http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
	f.Close()

	msg, err := stego_loader.DecodeFile(inputPath, pass, msgLen)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Decoding failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DecodeResponse{Message: msg})
}

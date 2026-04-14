package server

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func createTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestHandleEncode(t *testing.T) {
	s := NewStegoServer("", ":8080")
	imgData := createTestImage()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.png")
	part.Write(imgData)
	writer.WriteField("message", "hello")
	writer.WriteField("pass", "secret")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/encode", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	s.handleEncode(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	if rr.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected content type image/png, got %v", rr.Header().Get("Content-Type"))
	}
}

func TestHandleDecode(t *testing.T) {
	s := NewStegoServer("", ":8080")
	imgData := createTestImage()

	// First encode an image to decode later
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.png")
	part.Write(imgData)
	writer.WriteField("message", "hello")
	writer.WriteField("pass", "secret")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/encode", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	s.handleEncode(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("setup encode failed: %v", rr.Body.String())
	}

	encodedImg := rr.Body.Bytes()
	msgLen := len("hello") * 32

	// Now test decode
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("image", "encoded.png")
	part.Write(encodedImg)
	writer.WriteField("pass", "secret")
	writer.WriteField("msg_len", strconv.Itoa(msgLen))
	writer.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/decode", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr = httptest.NewRecorder()
	s.handleDecode(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v: %v", rr.Code, rr.Body.String())
	}

	var resp DecodeResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", resp.Message)
	}
}

func TestHandleErrors(t *testing.T) {
	s := NewStegoServer("", ":8080")

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/encode", nil)
		rr := httptest.NewRecorder()
		s.handleEncode(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %v", rr.Code)
		}
	})

	t.Run("missing params", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("image", "test.png")
		part.Write(createTestImage())
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/encode", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		s.handleEncode(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %v", rr.Code)
		}
	})
}

// Package telegram_bot provides a Telegram bot interface for steganography.
package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"stego/internal/stego_loader"
	"stego/pkg/video"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StegoBot encapsulates the Telegram bot logic.
type StegoBot struct {
	Token string
}

// Start runs the Telegram bot and listens for updates.
func (b *StegoBot) Start() error {
	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Photo != nil || update.Message.Document != nil || update.Message.Video != nil {
			b.handleMedia(bot, update.Message)
		} else if update.Message.IsCommand() {
			b.handleCommand(bot, update.Message)
		}
	}
	return nil
}

func (b *StegoBot) handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	var reply string
	switch msg.Command() {
	case "start", "help":
		reply = "Welcome to Go-Stego Bot!\n\n" +
			"How to use:\n" +
			"1. Send an image/video with the caption:\n" +
			"   `/encode <password> <message>`\n" +
			"2. Send an encoded image/video with the caption:\n" +
			"   `/decode <password> <length_in_bits>`\n\n" +
			"Example:\n" +
			"`/encode secret \"Hello world\"`"
	default:
		reply = "Unknown command. Use /help."
	}
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, reply))
}

func (b *StegoBot) handleMedia(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	caption := msg.Caption
	if caption == "" {
		return
	}

	parts := strings.SplitN(caption, " ", 3)
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Invalid caption. Use /encode or /decode."))
		return
	}

	cmd := parts[0]
	pass := parts[1]

	var fileID string
	var isVideo bool
	if msg.Photo != nil {
		fileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		fileID = msg.Video.FileID
		isVideo = true
	} else if msg.Document != nil {
		fileID = msg.Document.FileID
		// Check if it's an image or video by MIME type if needed
	}

	if fileID == "" {
		return
	}

	fileURL, _ := bot.GetFileDirectURL(fileID)
	tmpDir, _ := os.MkdirTemp("", "stego-bot-")
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input"+filepath.Ext(fileURL))
	err := downloadFile(inputPath, fileURL)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to download file."))
		return
	}

	switch cmd {
	case "/encode":
		if len(parts) < 3 {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Missing message for encoding."))
			return
		}
		text := parts[2]
		outPath := filepath.Join(tmpDir, "output"+filepath.Ext(fileURL))
		msgLen := len(text) * 32

		var encodeErr error
		if isVideo {
			encodeErr = video.EncodeVideo(inputPath, outPath, text, pass, msgLen, 20, -20)
		} else {
			encodeErr = stego_loader.EncodeFile(inputPath, outPath, text, pass, msgLen, 20, -20)
		}

		if encodeErr != nil {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Encoding failed: %v", encodeErr)))
			return
		}

		var upload tgbotapi.Chattable
		if isVideo {
			upload = tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FilePath(outPath))
		} else {
			upload = tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FilePath(outPath))
		}
		bot.Send(upload)
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Encoded successfully. Length: %d", msgLen)))

	case "/decode":
		if len(parts) < 3 {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Missing message length for decoding."))
			return
		}
		msgLen, _ := strconv.Atoi(parts[2])
		if msgLen == 0 {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Invalid length."))
			return
		}

		var decoded string
		var decodeErr error
		if isVideo {
			decoded, decodeErr = video.DecodeVideo(inputPath, pass, msgLen)
		} else {
			decoded, decodeErr = stego_loader.DecodeFile(inputPath, pass, msgLen)
		}

		if decodeErr != nil {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Decoding failed: %v", decodeErr)))
			return
		}
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Decoded message: %q", decoded)))

	default:
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Unknown command in caption. Use /encode or /decode."))
	}
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

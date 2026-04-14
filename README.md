# Go-Stego

Go-Stego is a steganography tool that allows you to hide messages in images and videos using the Discrete Cosine Transform (DCT) algorithm.

## Features

- **Image Support**: JPEG, PNG, BMP, and GIF (first frame).
- **Video Support**: MP4, AVI, MKV (requires `ffmpeg`).
- **Web Interface**: A built-in web server with a GUI.
- **CLI**: A powerful command-line interface with subcommands.
- **Docker Support**: Easily run the web server in a container.
- **Version**: 1.0.0

## Installation

### From Source

```bash
go build -o stego ./cmd/console/*.go
```

### With Docker

```bash
docker build -t stego-web .
docker run -p 8080:8080 stego-web
```

## Usage

### CLI

The CLI now uses a subcommand structure.

#### Encode a message into an image

```bash
./stego encode -i input.png -o output.png -m "Hello World" -p "mypassword"
```

#### Decode a message from an image

```bash
./stego decode -i output.png -p "mypassword" -l 352
```
Note: `-l` (length) is bits (len of string * 32).

#### Encode a message into a video

```bash
./stego encode -i input.mp4 -o output.mp4 -m "Hidden Message" -p "password" -v
```

### Telegram Bot

Start the bot interface:
```bash
export TELEGRAM_BOT_TOKEN="your_token_here"
./stego bot
```
The bot supports `/encode` and `/decode` commands via image/video captions.

### Web Server

Run the web server:
```bash
go run ./cmd/web/main.go
```
Then visit `http://localhost:8080`.

## Configuration

The application can be configured using environment variables or a configuration file at `config/app.env`.

Example `config/app.env`:
```env
TELEGRAM_BOT_TOKEN=your_token_here
PORT=8080
FRONTEND_PATH=./frontend
```

- **Robustness**: Use `-r` to set the strength of the hidden message (default 20).
- **Password**: Use `-p` to set a password.

## Documentation

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for details on the steganography process and project structure.

## License

MIT License. See [LICENSE](LICENSE) for details.

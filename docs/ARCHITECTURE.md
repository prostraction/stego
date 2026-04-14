# Go-Stego Architecture

This document describes the steganography algorithm and project structure of Go-Stego.

## Steganography Algorithm: DCT-based Embedding

Go-Stego uses a frequency-domain steganography technique based on the **Discrete Cosine Transform (DCT)**. This method is generally more robust to common image manipulations like compression compared to simple Least Significant Bit (LSB) embedding in the spatial domain.

### Process Overview

1.  **Image Segmentation**: The image is divided into non-overlapping 8x8 pixel blocks.
2.  **Frequency Transformation**: For each 8x8 block, a 2D DCT is applied to transform spatial pixel values into frequency coefficients.
3.  **Coefficient Selection**: Specific middle-frequency coefficients are selected for embedding. These coefficients are chosen because they offer a good balance:
    *   Low frequencies (DC component) are avoided as changes there are highly visible.
    *   High frequencies are avoided as they are often discarded during JPEG compression.
4.  **Bit Embedding**: Each bit of the message is embedded into a block by slightly modifying the relationship between three selected DCT coefficients (`pixelIndex1`, `pixelIndex2`, `pixelIndex3`).
    *   To embed a '1': The coefficients at `pixelIndex1` and `pixelIndex2` are set to a positive robustness value (`addMod`), while `pixelIndex3` is set to a negative value (`negMod`).
    *   To embed a '0': The opposite is done.
5.  **Jittering with Password**: The exact indices used within the `validIndexes` set are jittered based on a user-provided password. This adds a layer of security, as an attacker without the password won't know which coefficients to analyze.
6.  **Inverse Transformation**: After modification, the Inverse 2D DCT (IDCT) is applied to reconstruct the 8x8 pixel block.
7.  **Parity Decoding**: During extraction, the same indices are identified using the password, and the relationship between the coefficients is checked to determine if a '1' or '0' was stored. A majority-voting mechanism (across blocks) is used to increase robustness.

## Project Structure

- `cmd/`: Application entry points.
    - `console/`: The command-line interface (CLI).
    - `web/`: The web server.
- `internal/`: Private implementation details.
    - `stego_encoding/`: Core logic for bit embedding/extracting in pixel buffers.
    - `stego_loader/`: Logic for reading/writing various image formats and managing the pixel buffers.
    - `web_server/`: Implementation of the HTTP server and API handlers.
- `pkg/`: Publicly usable packages.
    - `dct/`: Pure mathematical implementation of 2D DCT and IDCT.
    - `video/`: Logic for video steganography via frame extraction and reconstruction using `ffmpeg`.
- `docs/`: Detailed documentation (like this file).
- `frontend/`: Static assets for the web interface (HTML, CSS, JS).
- `tests/`: Integration and end-to-end tests (or located alongside packages).

## Dependencies

- **Go 1.25+**: Language runtime.
- **ffmpeg**: Required for video steganography (must be in system PATH).
- **github.com/edwvee/exiffix**: Used for fixing image orientation during decoding.
- **golang.org/x/sync**: Used for concurrent processing in the CLI.

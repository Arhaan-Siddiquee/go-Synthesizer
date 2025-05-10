# Audio Equalizer Web Application

![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)
A web-based audio equalizer application built with Go that allows users to upload WAV files and apply basic equalization effects.

## Screenshot

![image](https://github.com/user-attachments/assets/f1e044e9-9bfa-4dfc-aa7c-079f497ba34d)


## Installation

### Prerequisites
- Go 1.21 or later
- Git

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/audio-equalizer.git
   cd audio-equalizer
   ```
2. Initialize the Go module:
   ```bash
   go mod init github.com/yourusername/audio-equalizer
   ```
3. Download dependencies:
   ```bash
   go mod tidy
   ```
4. Run the application:
   ```bash
   go run main.go
   ```
5. Access the application at:
   ```bash
   http://localhost:8080
   ```

## Usage Guide

### Upload Audio
- Click "Choose File" to select a WAV file  
- Click "Upload" to process the file

### Equalizer Controls
- **Bass (60Hz):** Adjust low frequencies  
- **Mid (1kHz):** Adjust mid-range frequencies  
- **Treble (10kHz):** Adjust high frequencies  

### Playback
- Original and processed audio players appear after upload  
- Use standard audio controls to play/pause  

### Download
- Right-click the processed audio player  
- Select "Save audio as..." to download  

## Technical Implementation

### Architecture
```bash
audio-equalizer/
├── main.go # Main application logic
├── go.mod # Dependency management
├── uploads/ # Original audio storage
├── processed/ # Processed audio storage
└── README.md # This file
```

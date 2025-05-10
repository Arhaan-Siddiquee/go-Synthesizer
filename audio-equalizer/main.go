package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const (
	uploadDir      = "./uploads"
	processedDir   = "./processed"
	maxUploadSize  = 50 * 1024 * 1024 // 50MB
)

func init() {
	// Create directories if they don't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(processedDir, 0755); err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/process", processHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
	http.Handle("/processed/", http.StripPrefix("/processed/", http.FileServer(http.Dir(processedDir))))

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Audio Equalizer</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
			.upload-form { margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
			.equalizer-form { padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
			.slider-container { margin: 15px 0; }
			.slider { width: 100%; }
			.audio-player { margin-top: 20px; width: 100%; }
			.button { padding: 10px 15px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
			.button:hover { background: #0056b3; }
		</style>
	</head>
	<body>
		<h1>Audio Equalizer</h1>
		
		<div class="upload-form">
			<h2>Upload Audio File</h2>
			<form action="/upload" method="post" enctype="multipart/form-data">
				<input type="file" name="audioFile" accept=".wav,.mp3" required>
				<button type="submit" class="button">Upload</button>
			</form>
		</div>

		<div class="equalizer-form">
			<h2>Equalizer Controls</h2>
			<form action="/process" method="post">
				<div class="slider-container">
					<label for="bass">Bass (60Hz):</label>
					<input type="range" id="bass" name="bass" min="0" max="200" value="100" class="slider">
					<span id="bassValue">100%</span>
				</div>
				
				<div class="slider-container">
					<label for="mid">Mid (1kHz):</label>
					<input type="range" id="mid" name="mid" min="0" max="200" value="100" class="slider">
					<span id="midValue">100%</span>
				</div>
				
				<div class="slider-container">
					<label for="treble">Treble (10kHz):</label>
					<input type="range" id="treble" name="treble" min="0" max="200" value="100" class="slider">
					<span id="trebleValue">100%</span>
				</div>
				
				<input type="hidden" id="filename" name="filename" value="">
				<button type="submit" class="button">Apply Equalization</button>
			</form>

			<div id="audioPlayer" class="audio-player">
				<!-- Audio player will be inserted here -->
			</div>
		</div>

		<script>
			// Update slider value displays
			document.getElementById('bass').addEventListener('input', function() {
				document.getElementById('bassValue').textContent = this.value + '%';
			});
			document.getElementById('mid').addEventListener('input', function() {
				document.getElementById('midValue').textContent = this.value + '%';
			});
			document.getElementById('treble').addEventListener('input', function() {
				document.getElementById('trebleValue').textContent = this.value + '%';
			});

			// Set filename after upload
			if (window.location.search.includes('file=')) {
				const urlParams = new URLSearchParams(window.location.search);
				const filename = urlParams.get('file');
				document.getElementById('filename').value = filename;
				
				// Create audio player
				const audioPlayer = document.createElement('audio');
				audioPlayer.controls = true;
				audioPlayer.src = '/uploads/' + filename;
				document.getElementById('audioPlayer').appendChild(audioPlayer);
				
				// Create processed audio player if exists
				const processedPlayer = document.createElement('audio');
				processedPlayer.controls = true;
				processedPlayer.src = '/processed/processed_' + filename;
				processedPlayer.onerror = function() {
					this.remove(); // Remove if file doesn't exist
				};
				document.getElementById('audioPlayer').appendChild(document.createElement('br'));
				document.getElementById('audioPlayer').appendChild(document.createTextNode('Processed:'));
				document.getElementById('audioPlayer').appendChild(document.createElement('br'));
				document.getElementById('audioPlayer').appendChild(processedPlayer);
			}
		</script>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("audioFile")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the upload file
	dstPath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error creating the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	// Redirect back to the main page with the filename
	http.Redirect(w, r, "/?file="+handler.Filename, http.StatusSeeOther)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get form values
	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "No filename provided", http.StatusBadRequest)
		return
	}

	bassStr := r.FormValue("bass")
	midStr := r.FormValue("mid")
	trebleStr := r.FormValue("treble")

	bass, err := strconv.Atoi(bassStr)
	if err != nil {
		http.Error(w, "Invalid bass value", http.StatusBadRequest)
		return
	}

	mid, err := strconv.Atoi(midStr)
	if err != nil {
		http.Error(w, "Invalid mid value", http.StatusBadRequest)
		return
	}

	treble, err := strconv.Atoi(trebleStr)
	if err != nil {
		http.Error(w, "Invalid treble value", http.StatusBadRequest)
		return
	}

	// Open the original file
	srcPath := filepath.Join(uploadDir, filename)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		http.Error(w, "Error opening the source file", http.StatusInternalServerError)
		return
	}
	defer srcFile.Close()

	// Process the audio
	processedAudio, err := applyEqualizer(srcFile, bass, mid, treble)
	if err != nil {
		http.Error(w, "Error processing audio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the processed file
	processedPath := filepath.Join(processedDir, "processed_"+filename)
	err = os.WriteFile(processedPath, processedAudio, 0644)
	if err != nil {
		http.Error(w, "Error saving processed file", http.StatusInternalServerError)
		return
	}

	// Redirect back to the main page with the filename
	http.Redirect(w, r, "/?file="+filename, http.StatusSeeOther)
}

func applyEqualizer(src io.Reader, bass, mid, treble int) ([]byte, error) {
	// We need to read the entire file into memory to work with it
	srcBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("error reading source: %v", err)
	}

	// Create a reader that implements io.ReadSeeker
	srcReader := bytes.NewReader(srcBytes)

	// Decode the WAV file
	decoder := wav.NewDecoder(srcReader)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("error decoding WAV: %v", err)
	}

	// Convert factors to float64 (100% = 1.0)
	bassFactor := float64(bass) / 100.0
	midFactor := float64(mid) / 100.0
	trebleFactor := float64(treble) / 100.0

	// Apply equalization
	applyBasicEQ(buf, bassFactor, midFactor, trebleFactor)

	// Create a buffer to hold the output
	var outBuf bytes.Buffer

	// Create encoder - using WAVE_FORMAT_PCM (1) as the default format
	encoder := wav.NewEncoder(&outBuf, 
		buf.Format.SampleRate, 
		int(decoder.BitDepth),  // Using decoder's bit depth
		buf.Format.NumChannels, 
		1) // Using PCM format

	if err := encoder.Write(buf); err != nil {
		return nil, fmt.Errorf("error encoding WAV: %v", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("error closing encoder: %v", err)
	}

	return outBuf.Bytes(), nil
}

func applyBasicEQ(buf *audio.IntBuffer, bassFactor, midFactor, trebleFactor float64) {
	// Remove unused sampleRate parameter since we're not using it
	numSamples := len(buf.Data)
	numChannels := buf.Format.NumChannels

	for i := 0; i < numSamples; i += numChannels {
		for ch := 0; ch < numChannels; ch++ {
			idx := i + ch
			if idx >= len(buf.Data) {
				continue
			}

			// Simple EQ implementation
			original := float64(buf.Data[idx])
			
			// Apply frequency-based gains (simplified)
			// In a real EQ, you would use proper filtering
			bassSample := original * bassFactor
			midSample := original * midFactor
			trebleSample := original * trebleFactor
			
			// Combine (very simplistic approach)
			combined := (bassSample + midSample + trebleSample) / 3
			
			// Clamp to 16-bit range
			if combined > 32767 {
				combined = 32767
			} else if combined < -32768 {
				combined = -32768
			}
			
			buf.Data[idx] = int(combined)
		}
	}
}
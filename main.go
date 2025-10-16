package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// --- KONFIGURASI YANG PERLU ANDA UBAH ---
const (
	// endpointURL = "https://sikomo.ragdalion.com:61524/api/app-user/attendance/verify-face"
	endpointURL = "https://facer.yusharwz.my.id/api/v1/recognizes"
	// endpointURL = "https://lppom-dev.ragdalion.com/api/face-verification-services/v1/recognizes"
	jwtToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJlZWY2ZjY1Ny03MTgwLTRkMzAtODk1Yi1mOWMxMjY2ZmY1ZWQifQ.G2SUvXLpsSZMHpYExQtI4qFJyg2NOne9twmEQ32T4Yc"
	imagePath = "2.jpg"
	// attendanceId    = "17354182-13e9-446b-8631-22735c7d2031"
	attendanceId    = "95e2a6d5-83de-44ed-a8ea-6dec5d32aa73"
	checkType       = "checkIn"
	concurrentUsers = 100
)

type Config struct {
	EndpointURL     string
	ImagePath       string
	JWTToken        string
	UserID          string
	CheckType       string
	ConcurrentUsers int
	RequestTimeout  time.Duration
}

func sendRequest(wg *sync.WaitGroup, client *http.Client, config Config, successCounter, failureCounter *uint64) {
	defer wg.Done()

	file, err := os.Open(config.ImagePath)
	if err != nil {
		fmt.Printf("[Gagal] Error membuka file '%s': %v\n", config.ImagePath, err)
		atomic.AddUint64(failureCounter, 1)
		return
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	_ = writer.WriteField("userId", config.UserID)
	_ = writer.WriteField("for", config.CheckType)

	part, err := writer.CreateFormFile("uploadImage", filepath.Base(config.ImagePath))
	if err != nil {
		fmt.Printf("[Gagal] Error membuat form file: %v\n", err)
		atomic.AddUint64(failureCounter, 1)
		return
	}
	if _, err = io.Copy(part, file); err != nil {
		fmt.Printf("[Gagal] Error menyalin file ke body: %v\n", err)
		atomic.AddUint64(failureCounter, 1)
		return
	}
	writer.Close()

	req, err := http.NewRequest("POST", config.EndpointURL, &requestBody)
	if err != nil {
		fmt.Printf("[Gagal] Error membuat request: %v\n", err)
		atomic.AddUint64(failureCounter, 1)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-USER-JWT", config.JWTToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Gagal] Error saat mengirim request: %v\n", err)
		atomic.AddUint64(failureCounter, 1)
		return
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		atomic.AddUint64(failureCounter, 1)
		fmt.Printf("[Gagal] Status: %s. Gagal membaca response body: %v\n", resp.Status, readErr)
		return
	}

	if resp.StatusCode == http.StatusOK {
		atomic.AddUint64(successCounter, 1)
		fmt.Printf("[Berhasil] Status: %s, Response: %s\n", resp.Status, string(bodyBytes))
	} else {
		atomic.AddUint64(failureCounter, 1)
		fmt.Printf("[Gagal] Status: %s, Response: %s\n", resp.Status, string(bodyBytes))
	}
}

func main() {
	config := Config{
		EndpointURL:     endpointURL,
		ImagePath:       imagePath,
		JWTToken:        jwtToken,
		UserID:          attendanceId,
		CheckType:       checkType,
		ConcurrentUsers: concurrentUsers,
		RequestTimeout:  60 * time.Minute,
	}

	client := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.ConcurrentUsers + 10,
			MaxIdleConnsPerHost: config.ConcurrentUsers + 10,
			IdleConnTimeout:     90 * time.Minute,
		},
	}

	fmt.Printf("🚀 Memulai load test ke: %s\n", config.EndpointURL)
	fmt.Printf("👥 Jumlah pengguna konkuren: %d\n\n", config.ConcurrentUsers)

	var wg sync.WaitGroup
	var successCounter uint64
	var failureCounter uint64

	startTime := time.Now()

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go sendRequest(&wg, client, config, &successCounter, &failureCounter)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("\n--- Hasil Load Test ---")
	fmt.Printf("✅ Request Berhasil: %d\n", successCounter)
	fmt.Printf("❌ Request Gagal   : %d\n", failureCounter)
	fmt.Printf("⏱️  Waktu Selesai   : %.2f detik\n", duration.Seconds())
	fmt.Println("-----------------------")
}

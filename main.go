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
	endpointURL     = "https://sikomo.ragdalion.com:61524/api/app-user/attendance/verify-face"
	jwtToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMGYzNDQwNDgtNDYwMS00MzZkLWIyMjMtMWUxZmRiZjQ0NzY1IiwibmlrIjoiYzAzNDc1YjItMDVmZi00Y2NhLWFhN2EtZTk2ZDk5ZTcxYjAxIiwiZW1haWwiOiJ0ZXN0aW5nQGdtYWlsLmNvbSIsImZ1bGxuYW1lIjoiVGVzdGluZyIsIm1hY0FkZHJlc3MiOm51bGwsInN0cnVjdHVyYWxQb3NpdGlvbiI6Ii0iLCJmdW5jdGlvbmFsUG9zaXRpb24iOiItIiwicGFzc3dvcmRMYXN0VXBkYXRlZEF0IjoiMjAyNS0wNy0wNlQxNjozMjowNi4wNDRaIiwiam9pbkRhdGUiOiIyMDIzLTAxLTAxVDAwOjAwOjAwLjAwMFoiLCJlbmREYXRlIjpudWxsLCJkZXBhcnRtZW50SWQiOiJjMDM0NzViMi0wNWZmLTRjY2EtYWE3YS1lOTZkOTllNzFiMDEiLCJqb2JUaXRsZUlkIjoiNzViNzliMTEtNzNhMC00YjRjLWFlYzItMWQ1OTYyZTE1OTEwIiwiZ3JhZGVJZCI6ImU4NzY1ODNjLTUwNTEtNDY3Yy04OWNmLWFkMDBiNmVkMWE5ZCIsImlzQWN0aXZlIjp0cnVlLCJ3aGl0ZUxpc3QiOnRydWUsImF2YXRhclBhdGgiOiJodHRwczovL3Npa29tby5yYWdkYWxpb24uY29tOjYxNTI0L2ltYWdlcy91c2VyLzBmMzQ0MDQ4LTQ2MDEtNDM2ZC1iMjIzLTFlMWZkYmY0NDc2NS9mcm9udF8xNzUxODE5NjcwMDYxLmpwZyIsInJvbGUiOnsiaWQiOiI3YTI4OGYzNy02ZGRjLTQxYjUtOWJkZS0wMWQzNDk4MjIzOGMiLCJuYW1lIjoiSFIgMyIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNjk4WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNjk4WiIsImRlbGV0ZWRBdCI6bnVsbH0sImRlcGFydG1lbnQiOnsiaWQiOiJjMDM0NzViMi0wNWZmLTRjY2EtYWE3YS1lOTZkOTllNzFiMDEiLCJiYXRpa0NvZGUiOiIxLjEyLjEiLCJuYW1lIjoiSFIgR0EgREVQVCIsImVuZERhdGUiOm51bGwsImJhdGlrVXBkYXRlQXQiOiIyMDI1LTA2LTA0VDA5OjU1OjM1LjIyMFoiLCJjcmVhdGVkQXQiOiIyMDI1LTAyLTIxVDEwOjU5OjEzLjczNFoiLCJ1cGRhdGVkQXQiOiIyMDI1LTA2LTA0VDA5OjU1OjM1LjIyMVoiLCJkZWxldGVkQXQiOm51bGx9LCJqb2JUaXRsZSI6eyJpZCI6Ijc1Yjc5YjExLTczYTAtNGI0Yy1hZWMyLTFkNTk2MmUxNTkxMCIsIm5hbWUiOiItIiwiY3JlYXRlZEF0IjoiMjAyNS0wMi0yMFQwMjo0ODowNC40ODRaIiwidXBkYXRlZEF0IjoiMjAyNS0wMi0yMFQwMjo0ODowNC40ODRaIiwiZGVsZXRlZEF0IjpudWxsfSwiZ3JhZGUiOnsibmFtZSI6IlYtQSIsImlkIjoiZTg3NjU4M2MtNTA1MS00NjdjLTg5Y2YtYWQwMGI2ZWQxYTlkIiwiYmF0aWtDb2RlIjoiNWY5YzNhODYtYTBkZC00NjUzLTgzODgtYTFlYmEyOWQzZGNkIiwiZW5kRGF0ZSI6bnVsbCwiYmF0aWtVcGRhdGVBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNjA0WiIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNjA0WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNjA0WiIsImRlbGV0ZWRBdCI6bnVsbH0sInJvbGVJZCI6IjdhMjg4ZjM3LTZkZGMtNDFiNS05YmRlLTAxZDM0OTgyMjM4YyIsImNyZWF0ZWRBdCI6IjIwMjUtMDctMDZUMTY6MzI6MDYuMDQ1WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDctMDdUMDM6MzY6MTkuNDk4WiIsImRlbGV0ZWRBdCI6bnVsbCwicGhvbmVOdW1iZXIiOiIwODEyMzQ1Njc4OSIsIndvcmtpbmdTdGF0dXMiOiJQS1dUVCIsImdlbmRlciI6Im1hbGUiLCJzdGFydERhdGUiOiIyMDIzLTAyLTAxVDAwOjAwOjAwLjAwMFoiLCJlZHVjYXRpb24iOiJCYWNoZWxvcidzIERlZ3JlZSIsImJhdGlrVXBkYXRlZEF0IjoiMjAyNS0wNy0wNlQxMToxNTowNC41ODdaIiwibWFwcGluZ0luZm9ybWF0aW9uIjpmYWxzZSwiaW1hZ2VzIjpbeyJpZCI6ImMwOGMzNWUxLTNiMTMtNDI0My05YjE4LWFjNDcwMjBhNTEyMiIsInVzZXJJZCI6IjBmMzQ0MDQ4LTQ2MDEtNDM2ZC1iMjIzLTFlMWZkYmY0NDc2NSIsImxlZnRTaWRlVXJsIjoiaHR0cHM6Ly9zaWtvbW8ucmFnZGFsaW9uLmNvbTo2MTUyNC9pbWFnZXMvdXNlci8wZjM0NDA0OC00NjAxLTQzNmQtYjIyMy0xZTFmZGJmNDQ3NjUvbGVmdF8xNzUxODE5NjY5OTEzLmpwZyIsImZyb250U2lkZVVybCI6Imh0dHBzOi8vc2lrb21vLnJhZ2RhbGlvbi5jb206NjE1MjQvaW1hZ2VzL3VzZXIvMGYzNDQwNDgtNDYwMS00MzZkLWIyMjMtMWUxZmRiZjQ0NzY1L2Zyb250XzE3NTE4MTk2NzAwNjEuanBnIiwicmlnaHRTaWRlVXJsIjoiaHR0cHM6Ly9zaWtvbW8ucmFnZGFsaW9uLmNvbTo2MTUyNC9pbWFnZXMvdXNlci8wZjM0NDA0OC00NjAxLTQzNmQtYjIyMy0xZTFmZGJmNDQ3NjUvcmlnaHRfMTc1MTgxOTY3MDA2NC5qcGciLCJjcmVhdGVkQXQiOiIyMDI1LTA3LTA2VDE2OjM0OjMwLjEwMloiLCJ1cGRhdGVkQXQiOiIyMDI1LTA3LTA2VDE2OjM0OjMwLjEwMloiLCJkZWxldGVkQXQiOm51bGx9XSwiZmNtVG9rZW4iOm51bGwsInBhc3N3b3JkTGFzdFVwZGF0ZWRBdFRvU3RyaW5nIjoiTGFzdCB1cGRhdGUgMTIgaG91cnMgYWdvIiwiam9iVGl0bGVOYW1lIjoiLSIsImxhc3RBY3RpdmVUb2tlbiI6bnVsbH0sInBsYXRmb3JtIjoiYXBwLXVzZXIiLCJpYXQiOjE3NTE4NjIzMDZ9.R6fHqgfPSkkmxJrPndn_T5E_gP51mJvAMitZ938xUP4"
	imagePath       = "1.jpg"
	attendanceId    = "c76be934-07bf-4a80-a0ab-e1700594f357"
	checkType       = "checkIn"
	concurrentUsers = 1000
)

type Config struct {
	EndpointURL     string
	ImagePath       string
	JWTToken        string
	AttendanceID    string
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
	_ = writer.WriteField("attendanceId", config.AttendanceID)
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
	req.Header.Set("Authorization", "Bearer "+config.JWTToken)
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
		AttendanceID:    attendanceId,
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

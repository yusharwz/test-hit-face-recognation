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
	jwtToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMGYzNDQwNDgtNDYwMS00MzZkLWIyMjMtMWUxZmRiZjQ0NzY1IiwibmlrIjoiYzAzNDc1YjItMDVmZi00Y2NhLWFhN2EtZTk2ZDk5ZTcxYjAxIiwiZW1haWwiOiJ0ZXN0aW5nQGdtYWlsLmNvbSIsImZ1bGxuYW1lIjoiVGVzdGluZyIsIm1hY0FkZHJlc3MiOm51bGwsInN0cnVjdHVyYWxQb3NpdGlvbiI6Ii0iLCJmdW5jdGlvbmFsUG9zaXRpb24iOiItIiwicGFzc3dvcmRMYXN0VXBkYXRlZEF0IjoiMjAyNS0wNy0wNlQxNjozMjowNi4wNDRaIiwiam9pbkRhdGUiOiIyMDIzLTAxLTAxVDAwOjAwOjAwLjAwMFoiLCJlbmREYXRlIjpudWxsLCJkZXBhcnRtZW50SWQiOiJjMDM0NzViMi0wNWZmLTRjY2EtYWE3YS1lOTZkOTllNzFiMDEiLCJqb2JUaXRsZUlkIjoiNzViNzliMTEtNzNhMC00YjRjLWFlYzItMWQ1OTYyZTE1OTEwIiwiZ3JhZGVJZCI6ImU4NzY1ODNjLTUwNTEtNDY3Yy04OWNmLWFkMDBiNmVkMWE5ZCIsImlzQWN0aXZlIjp0cnVlLCJ3aGl0ZUxpc3QiOmZhbHNlLCJhdmF0YXJQYXRoIjoiaHR0cHM6Ly9zaWtvbW8ucmFnZGFsaW9uLmNvbTo2MTUyNC9pbWFnZXMvdXNlci8wZjM0NDA0OC00NjAxLTQzNmQtYjIyMy0xZTFmZGJmNDQ3NjUvZnJvbnRfMTc1MTg3NzI4MzYwMS5qcGciLCJyb2xlIjp7ImlkIjoiN2EyODhmMzctNmRkYy00MWI1LTliZGUtMDFkMzQ5ODIyMzhjIiwibmFtZSI6IkhSIDMiLCJjcmVhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjY5OFoiLCJ1cGRhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjY5OFoiLCJkZWxldGVkQXQiOm51bGx9LCJkZXBhcnRtZW50Ijp7ImlkIjoiYzAzNDc1YjItMDVmZi00Y2NhLWFhN2EtZTk2ZDk5ZTcxYjAxIiwiYmF0aWtDb2RlIjoiMS4xMi4xIiwibmFtZSI6IkhSIEdBIERFUFQiLCJlbmREYXRlIjpudWxsLCJiYXRpa1VwZGF0ZUF0IjoiMjAyNS0wNi0wNFQwOTo1NTozNS4yMjBaIiwiY3JlYXRlZEF0IjoiMjAyNS0wMi0yMVQxMDo1OToxMy43MzRaIiwidXBkYXRlZEF0IjoiMjAyNS0wNi0wNFQwOTo1NTozNS4yMjFaIiwiZGVsZXRlZEF0IjpudWxsfSwiam9iVGl0bGUiOnsiaWQiOiI3NWI3OWIxMS03M2EwLTRiNGMtYWVjMi0xZDU5NjJlMTU5MTAiLCJuYW1lIjoiLSIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNDg0WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDItMjBUMDI6NDg6MDQuNDg0WiIsImRlbGV0ZWRBdCI6bnVsbH0sImdyYWRlIjp7Im5hbWUiOiJWLUEiLCJpZCI6ImU4NzY1ODNjLTUwNTEtNDY3Yy04OWNmLWFkMDBiNmVkMWE5ZCIsImJhdGlrQ29kZSI6IjVmOWMzYTg2LWEwZGQtNDY1My04Mzg4LWExZWJhMjlkM2RjZCIsImVuZERhdGUiOm51bGwsImJhdGlrVXBkYXRlQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjYwNFoiLCJjcmVhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjYwNFoiLCJ1cGRhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjYwNFoiLCJkZWxldGVkQXQiOm51bGx9LCJyb2xlSWQiOiI3YTI4OGYzNy02ZGRjLTQxYjUtOWJkZS0wMWQzNDk4MjIzOGMiLCJjcmVhdGVkQXQiOiIyMDI1LTA3LTA2VDE2OjMyOjA2LjA0NVoiLCJ1cGRhdGVkQXQiOiIyMDI1LTA3LTA3VDA5OjI1OjIzLjUzOFoiLCJkZWxldGVkQXQiOm51bGwsInBob25lTnVtYmVyIjoiMDgxMjM0NTY3ODkiLCJ3b3JraW5nU3RhdHVzIjoiUEtXVFQiLCJnZW5kZXIiOiJtYWxlIiwic3RhcnREYXRlIjoiMjAyMy0wMi0wMVQwMDowMDowMC4wMDBaIiwiZWR1Y2F0aW9uIjoiQmFjaGVsb3IncyBEZWdyZWUiLCJiYXRpa1VwZGF0ZWRBdCI6IjIwMjUtMDctMDZUMTE6MTU6MDQuNTg3WiIsIm1hcHBpbmdJbmZvcm1hdGlvbiI6ZmFsc2UsImltYWdlcyI6W3siaWQiOiIwYzg0ZmFiYS0xOTkxLTRlMDEtOWIxNy0zMDAwYWMwNmFhYTUiLCJ1c2VySWQiOiIwZjM0NDA0OC00NjAxLTQzNmQtYjIyMy0xZTFmZGJmNDQ3NjUiLCJsZWZ0U2lkZVVybCI6Imh0dHBzOi8vc2lrb21vLnJhZ2RhbGlvbi5jb206NjE1MjQvaW1hZ2VzL3VzZXIvMGYzNDQwNDgtNDYwMS00MzZkLWIyMjMtMWUxZmRiZjQ0NzY1L2xlZnRfMTc1MTg3NzI4MzU5Ny5qcGciLCJmcm9udFNpZGVVcmwiOiJodHRwczovL3Npa29tby5yYWdkYWxpb24uY29tOjYxNTI0L2ltYWdlcy91c2VyLzBmMzQ0MDQ4LTQ2MDEtNDM2ZC1iMjIzLTFlMWZkYmY0NDc2NS9mcm9udF8xNzUxODc3MjgzNjAxLmpwZyIsInJpZ2h0U2lkZVVybCI6Imh0dHBzOi8vc2lrb21vLnJhZ2RhbGlvbi5jb206NjE1MjQvaW1hZ2VzL3VzZXIvMGYzNDQwNDgtNDYwMS00MzZkLWIyMjMtMWUxZmRiZjQ0NzY1L3JpZ2h0XzE3NTE4NzcyODM2MDIuanBnIiwiY3JlYXRlZEF0IjoiMjAyNS0wNy0wN1QwODozNDo0My42MDZaIiwidXBkYXRlZEF0IjoiMjAyNS0wNy0wN1QwODozNDo0My42MDZaIiwiZGVsZXRlZEF0IjpudWxsfV0sImZjbVRva2VuIjpudWxsLCJwYXNzd29yZExhc3RVcGRhdGVkQXRUb1N0cmluZyI6Ikxhc3QgdXBkYXRlIDE3IGhvdXJzIGFnbyIsImpvYlRpdGxlTmFtZSI6Ii0iLCJsYXN0QWN0aXZlVG9rZW4iOm51bGx9LCJwbGF0Zm9ybSI6ImFwcC11c2VyIiwiaWF0IjoxNzUxODgxMjI3fQ.TiPCgf2itSFQHB1dw25DztKbdtxL1Z7sVdpvB6X-Gzc"
	imagePath       = "1.jpg"
	attendanceId    = "c76be934-07bf-4a80-a0ab-e1700594f357"
	checkType       = "checkIn"
	concurrentUsers = 100
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

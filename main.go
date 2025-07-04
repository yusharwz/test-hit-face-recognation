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
	jwtToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiNmI4YjQ1MzctZTU0YS00ZjZjLTkxMGQtYjIxODExZjBjMjE4IiwibmlrIjoiMTAwODAwNCIsImVtYWlsIjoid2luYXJkaXNheWlkQGdtYWlsLmNvbSIsImZ1bGxuYW1lIjoiV0lOQVJESSIsIm1hY0FkZHJlc3MiOm51bGwsInN0cnVjdHVyYWxQb3NpdGlvbiI6Ii0iLCJmdW5jdGlvbmFsUG9zaXRpb24iOiItIiwicGFzc3dvcmRMYXN0VXBkYXRlZEF0IjoiMjAyNS0wNi0zMFQwMzozMzoxNi45NTRaIiwiam9pbkRhdGUiOiIyMDEwLTA4LTE5VDAwOjAwOjAwLjAwMFoiLCJlbmREYXRlIjpudWxsLCJkZXBhcnRtZW50SWQiOiIyNWNmZDFkMS03ODgyLTRkYTQtYWY1MC0yMzBjODA0MTMxNWIiLCJqb2JUaXRsZUlkIjoiNzViNzliMTEtNzNhMC00YjRjLWFlYzItMWQ1OTYyZTE1OTEwIiwiZ3JhZGVJZCI6IjE3YzM1ZjU1LTJjMTAtNDZiOS1iYTMyLWY2NTM0ZDg0MzFhMSIsImlzQWN0aXZlIjp0cnVlLCJ3aGl0ZUxpc3QiOmZhbHNlLCJhdmF0YXJQYXRoIjoiaHR0cHM6Ly9zaWtvbW8ucmFnZGFsaW9uLmNvbTo2MTUyNC9pbWFnZXMvdXNlci82YjhiNDUzNy1lNTRhLTRmNmMtOTEwZC1iMjE4MTFmMGMyMTgvZnJvbnRfMTc1MTI1NTkzNjgwNS5KUEciLCJyb2xlIjp7ImlkIjoiYzMyZjMwZjgtODcwNC00NjEwLWJmNGEtMGM3YjJlMWMyODc4IiwibmFtZSI6Ik1hbmFnZW1lbnQiLCJjcmVhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjY5OFoiLCJ1cGRhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjY5OFoiLCJkZWxldGVkQXQiOm51bGx9LCJkZXBhcnRtZW50Ijp7ImlkIjoiMjVjZmQxZDEtNzg4Mi00ZGE0LWFmNTAtMjMwYzgwNDEzMTViIiwiYmF0aWtDb2RlIjoiMS4zLjMiLCJuYW1lIjoiQ09OVFJPTCBERVBUIiwiZW5kRGF0ZSI6bnVsbCwiYmF0aWtVcGRhdGVBdCI6IjIwMjUtMDYtMDRUMDk6NTU6MzguMTQ3WiIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjFUMTA6NTk6MTYuOTA0WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDYtMDRUMDk6NTU6MzguMTQ4WiIsImRlbGV0ZWRBdCI6bnVsbH0sImpvYlRpdGxlIjp7ImlkIjoiNzViNzliMTEtNzNhMC00YjRjLWFlYzItMWQ1OTYyZTE1OTEwIiwibmFtZSI6Ii0iLCJjcmVhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjQ4NFoiLCJ1cGRhdGVkQXQiOiIyMDI1LTAyLTIwVDAyOjQ4OjA0LjQ4NFoiLCJkZWxldGVkQXQiOm51bGx9LCJncmFkZSI6eyJuYW1lIjoiR09MIElWLUUxIiwiaWQiOiIxN2MzNWY1NS0yYzEwLTQ2YjktYmEzMi1mNjUzNGQ4NDMxYTEiLCJiYXRpa0NvZGUiOiJJVi1FMS0xIiwiZW5kRGF0ZSI6bnVsbCwiYmF0aWtVcGRhdGVBdCI6IjIwMjUtMDYtMDRUMDk6NTY6MDEuMjI4WiIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjJUMDQ6Mzc6NTguNDAxWiIsInVwZGF0ZWRBdCI6IjIwMjUtMDYtMDRUMDk6NTY6MDEuMjI5WiIsImRlbGV0ZWRBdCI6bnVsbH0sInJvbGVJZCI6ImMzMmYzMGY4LTg3MDQtNDYxMC1iZjRhLTBjN2IyZTFjMjg3OCIsImNyZWF0ZWRBdCI6IjIwMjUtMDItMjRUMDM6NTQ6MDEuODg2WiIsInVwZGF0ZWRBdCI6IjIwMjUtMDctMDNUMDc6MDk6MjcuNjQxWiIsImRlbGV0ZWRBdCI6bnVsbCwicGhvbmVOdW1iZXIiOiIwODE1MTEyNTg3MjciLCJ3b3JraW5nU3RhdHVzIjoiUEtXVFQiLCJnZW5kZXIiOiJtYWxlIiwic3RhcnREYXRlIjoiMjAxMC0wOC0xOVQwMDowMDowMC4wMDBaIiwiZWR1Y2F0aW9uIjoiQmFjaGVsb3IiLCJiYXRpa1VwZGF0ZWRBdCI6IjIwMjUtMDYtMjVUMDg6MDQ6MDcuMTIwWiIsIm1hcHBpbmdJbmZvcm1hdGlvbiI6ZmFsc2UsImltYWdlcyI6W3siaWQiOiIxNTY0Y2I4NC03MTgxLTQ5MTctOGI2Yi0xZGMxZjkxOGE2NTgiLCJ1c2VySWQiOiI2YjhiNDUzNy1lNTRhLTRmNmMtOTEwZC1iMjE4MTFmMGMyMTgiLCJsZWZ0U2lkZVVybCI6Imh0dHBzOi8vc2lrb21vLnJhZ2RhbGlvbi5jb206NjE1MjQvaW1hZ2VzL3VzZXIvNmI4YjQ1MzctZTU0YS00ZjZjLTkxMGQtYjIxODExZjBjMjE4L2xlZnRfMTc1MTI1NTkzNjgwMi5KUEciLCJmcm9udFNpZGVVcmwiOiJodHRwczovL3Npa29tby5yYWdkYWxpb24uY29tOjYxNTI0L2ltYWdlcy91c2VyLzZiOGI0NTM3LWU1NGEtNGY2Yy05MTBkLWIyMTgxMWYwYzIxOC9mcm9udF8xNzUxMjU1OTM2ODA1LkpQRyIsInJpZ2h0U2lkZVVybCI6Imh0dHBzOi8vc2lrb21vLnJhZ2RhbGlvbi5jb206NjE1MjQvaW1hZ2VzL3VzZXIvNmI4YjQ1MzctZTU0YS00ZjZjLTkxMGQtYjIxODExZjBjMjE4L3JpZ2h0XzE3NTEyNTU5MzY4MDcuSlBHIiwiY3JlYXRlZEF0IjoiMjAyNS0wNi0zMFQwMzo1ODo1Ni44MTVaIiwidXBkYXRlZEF0IjoiMjAyNS0wNi0zMFQwMzo1ODo1Ni44MTVaIiwiZGVsZXRlZEF0IjpudWxsfV0sImZjbVRva2VuIjpudWxsLCJwYXNzd29yZExhc3RVcGRhdGVkQXRUb1N0cmluZyI6Ikxhc3QgdXBkYXRlIDMgZGF5cyBhZ28iLCJqb2JUaXRsZU5hbWUiOiItIiwibGFzdEFjdGl2ZVRva2VuIjoiZXlKaGJHY2lPaUpJVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SjFjMlZ5SWpwN0ltbGtJam9pTm1JNFlqUTFNemN0WlRVMFlTMDBaalpqTFRreE1HUXRZakl4T0RFeFpqQmpNakU0SWl3aWJtbHJJam9pTVRBd09EQXdOQ0lzSW1WdFlXbHNJam9pZDJsdVlYSmthWE5oZVdsa1FHZHRZV2xzTG1OdmJTSXNJbVoxYkd4dVlXMWxJam9pVjBsT1FWSkVTU0lzSW0xaFkwRmtaSEpsYzNNaU9tNTFiR3dzSW5OMGNuVmpkSFZ5WVd4UWIzTnBkR2x2YmlJNklpMGlMQ0ptZFc1amRHbHZibUZzVUc5emFYUnBiMjRpT2lJdElpd2ljR0Z6YzNkdmNtUk1ZWE4wVlhCa1lYUmxaRUYwSWpvaU1qQXlOUzB3Tmkwek1GUXdNem96TXpveE5pNDVOVFJhSWl3aWFtOXBia1JoZEdVaU9pSXlNREV3TFRBNExURTVWREF3T2pBd09qQXdMakF3TUZvaUxDSmxibVJFWVhSbElqcHVkV3hzTENKa1pYQmhjblJ0Wlc1MFNXUWlPaUl5TldObVpERmtNUzAzT0RneUxUUmtZVFF0WVdZMU1DMHlNekJqT0RBME1UTXhOV0lpTENKcWIySlVhWFJzWlVsa0lqb2lOelZpTnpsaU1URXROek5oTUMwMFlqUmpMV0ZsWXpJdE1XUTFPVFl5WlRFMU9URXdJaXdpWjNKaFpHVkpaQ0k2SWpFM1l6TTFaalUxTFRKak1UQXRORFppT1MxaVlUTXlMV1kyTlRNMFpEZzBNekZoTVNJc0ltbHpRV04wYVhabElqcDBjblZsTENKM2FHbDBaVXhwYzNRaU9tWmhiSE5sTENKaGRtRjBZWEpRWVhSb0lqb2lhSFIwY0hNNkx5OXphV3R2Ylc4dWNtRm5aR0ZzYVc5dUxtTnZiVG8yTVRVeU5DOXBiV0ZuWlhNdmRYTmxjaTgyWWpoaU5EVXpOeTFsTlRSaExUUm1ObU10T1RFd1pDMWlNakU0TVRGbU1HTXlNVGd2Wm5KdmJuUmZNVGMxTVRJMU5Ua3pOamd3TlM1S1VFY2lMQ0p5YjJ4bElqcDdJbWxrSWpvaVl6TXlaak13WmpndE9EY3dOQzAwTmpFd0xXSm1OR0V0TUdNM1lqSmxNV015T0RjNElpd2libUZ0WlNJNklrMWhibUZuWlcxbGJuUWlMQ0pqY21WaGRHVmtRWFFpT2lJeU1ESTFMVEF5TFRJd1ZEQXlPalE0T2pBMExqWTVPRm9pTENKMWNHUmhkR1ZrUVhRaU9pSXlNREkxTFRBeUxUSXdWREF5T2pRNE9qQTBMalk1T0ZvaUxDSmtaV3hsZEdWa1FYUWlPbTUxYkd4OUxDSmtaWEJoY25SdFpXNTBJanA3SW1sa0lqb2lNalZqWm1ReFpERXROemc0TWkwMFpHRTBMV0ZtTlRBdE1qTXdZemd3TkRFek1UVmlJaXdpWW1GMGFXdERiMlJsSWpvaU1TNHpMak1pTENKdVlXMWxJam9pUTA5T1ZGSlBUQ0JFUlZCVUlpd2laVzVrUkdGMFpTSTZiblZzYkN3aVltRjBhV3RWY0dSaGRHVkJkQ0k2SWpJd01qVXRNRFl0TURSVU1EazZOVFU2TXpndU1UUTNXaUlzSW1OeVpXRjBaV1JCZENJNklqSXdNalV0TURJdE1qRlVNVEE2TlRrNk1UWXVPVEEwV2lJc0luVndaR0YwWldSQmRDSTZJakl3TWpVdE1EWXRNRFJVTURrNk5UVTZNemd1TVRRNFdpSXNJbVJsYkdWMFpXUkJkQ0k2Ym5Wc2JIMHNJbXB2WWxScGRHeGxJanA3SW1sa0lqb2lOelZpTnpsaU1URXROek5oTUMwMFlqUmpMV0ZsWXpJdE1XUTFPVFl5WlRFMU9URXdJaXdpYm1GdFpTSTZJbEJ5YjJabGMzTnBiMjVoYkNJc0ltTnlaV0YwWldSQmRDSTZJakl3TWpVdE1ESXRNakJVTURJNk5EZzZNRFF1TkRnMFdpSXNJblZ3WkdGMFpXUkJkQ0k2SWpJd01qVXRNREl0TWpCVU1ESTZORGc2TURRdU5EZzBXaUlzSW1SbGJHVjBaV1JCZENJNmJuVnNiSDBzSW1keVlXUmxJanA3SW01aGJXVWlPaUpIVDB3Z1NWWXRSVEVpTENKcFpDSTZJakUzWXpNMVpqVTFMVEpqTVRBdE5EWmlPUzFpWVRNeUxXWTJOVE0wWkRnME16RmhNU0lzSW1KaGRHbHJRMjlrWlNJNklrbFdMVVV4TFRFaUxDSmxibVJFWVhSbElqcHVkV3hzTENKaVlYUnBhMVZ3WkdGMFpVRjBJam9pTWpBeU5TMHdOaTB3TkZRd09UbzFOam93TVM0eU1qaGFJaXdpWTNKbFlYUmxaRUYwSWpvaU1qQXlOUzB3TWkweU1sUXdORG96TnpvMU9DNDBNREZhSWl3aWRYQmtZWFJsWkVGMElqb2lNakF5TlMwd05pMHdORlF3T1RvMU5qb3dNUzR5TWpsYUlpd2laR1ZzWlhSbFpFRjBJanB1ZFd4c2ZTd2ljbTlzWlVsa0lqb2lZek15WmpNd1pqZ3RPRGN3TkMwME5qRXdMV0ptTkdFdE1HTTNZakpsTVdNeU9EYzRJaXdpWTNKbFlYUmxaRUYwSWpvaU1qQXlOUzB3TWkweU5GUXdNem8xTkRvd01TNDRPRFphSWl3aWRYQmtZWFJsWkVGMElqb2lNakF5TlMwd055MHdNMVF3TXpvMU1Ub3pPUzR6TWpGYUlpd2laR1ZzWlhSbFpFRjBJanB1ZFd4c0xDSndhRzl1WlU1MWJXSmxjaUk2SWpBNE1UVXhNVEkxT0RjeU55SXNJbmR2Y210cGJtZFRkR0YwZFhNaU9pSlFTMWRVVkNJc0ltZGxibVJsY2lJNkltMWhiR1VpTENKemRHRnlkRVJoZEdVaU9pSXlNREV3TFRBNExURTVWREF3T2pBd09qQXdMakF3TUZvaUxDSmxaSFZqWVhScGIyNGlPaUpDWVdOb1pXeHZjaUlzSW1KaGRHbHJWWEJrWVhSbFpFRjBJam9pTWpBeU5TMHdOaTB5TlZRd09Eb3dORG93Tnk0eE1qQmFJaXdpYldGd2NHbHVaMGx1Wm05eWJXRjBhVzl1SWpwbVlXeHpaU3dpYVcxaFoyVnpJanBiZXlKcFpDSTZJakUxTmpSallqZzBMVGN4T0RFdE5Ea3hOeTA0WWpaaUxURmtZekZtT1RFNFlUWTFPQ0lzSW5WelpYSkpaQ0k2SWpaaU9HSTBOVE0zTFdVMU5HRXROR1kyWXkwNU1UQmtMV0l5TVRneE1XWXdZekl4T0NJc0lteGxablJUYVdSbFZYSnNJam9pYUhSMGNITTZMeTl6YVd0dmJXOHVjbUZuWkdGc2FXOXVMbU52YlRvMk1UVXlOQzlwYldGblpYTXZkWE5sY2k4MllqaGlORFV6TnkxbE5UUmhMVFJtTm1NdE9URXdaQzFpTWpFNE1URm1NR015TVRndmJHVm1kRjh4TnpVeE1qVTFPVE0yT0RBeUxrcFFSeUlzSW1aeWIyNTBVMmxrWlZWeWJDSTZJbWgwZEhCek9pOHZjMmxyYjIxdkxuSmhaMlJoYkdsdmJpNWpiMjA2TmpFMU1qUXZhVzFoWjJWekwzVnpaWEl2Tm1JNFlqUTFNemN0WlRVMFlTMDBaalpqTFRreE1HUXRZakl4T0RFeFpqQmpNakU0TDJaeWIyNTBYekUzTlRFeU5UVTVNelk0TURVdVNsQkhJaXdpY21sbmFIUlRhV1JsVlhKc0lqb2lhSFIwY0hNNkx5OXphV3R2Ylc4dWNtRm5aR0ZzYVc5dUxtTnZiVG8yTVRVeU5DOXBiV0ZuWlhNdmRYTmxjaTgyWWpoaU5EVXpOeTFsTlRSaExUUm1ObU10T1RFd1pDMWlNakU0TVRGbU1HTXlNVGd2Y21sbmFIUmZNVGMxTVRJMU5Ua3pOamd3Tnk1S1VFY2lMQ0pqY21WaGRHVmtRWFFpT2lJeU1ESTFMVEEyTFRNd1ZEQXpPalU0T2pVMkxqZ3hOVm9pTENKMWNHUmhkR1ZrUVhRaU9pSXlNREkxTFRBMkxUTXdWREF6T2pVNE9qVTJMamd4TlZvaUxDSmtaV3hsZEdWa1FYUWlPbTUxYkd4OVhTd2labU50Vkc5clpXNGlPbTUxYkd3c0luQmhjM04zYjNKa1RHRnpkRlZ3WkdGMFpXUkJkRlJ2VTNSeWFXNW5Jam9pVEdGemRDQjFjR1JoZEdVZ015QmtZWGx6SUdGbmJ5SXNJbXB2WWxScGRHeGxUbUZ0WlNJNklpMGlMQ0pzWVhOMFFXTjBhWFpsVkc5clpXNGlPbTUxYkd4OUxDSndiR0YwWm05eWJTSTZJbmRsWWkxaFpHMXBiaUlzSW1saGRDSTZNVGMxTVRVeE56UXlPSDAuc3M5NWo5SnFxQkdlUlVWcHQ1TGptM2IwaXpLY0RfMmVLLTNJN0xscXBCQSJ9LCJwbGF0Zm9ybSI6ImFwcC11c2VyIiwiaWF0IjoxNzUxNTMwOTgyfQ.4DwKPK1Zg54DxRNDGX-zyILUYWD7kR0TsYqqczbdf3Y"
	imagePath       = "1.jpg"
	attendanceId    = "c76be934-07bf-4a80-a0ab-e1700594f357"
	checkType       = "checkIn"
	concurrentUsers = 150
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

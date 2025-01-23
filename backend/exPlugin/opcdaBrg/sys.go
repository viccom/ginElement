package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/shirou/gopsutil/v3/cpu"
)

// GetHardwareID 生成一个基于CPU信息的唯一ID
func GetHardwareID() (string, error) {
	// 获取CPU信息
	info, err := cpu.Info()
	if err != nil {
		return "", fmt.Errorf("无法获取CPU信息: %v", err)
	}

	if len(info) == 0 {
		return "", fmt.Errorf("未找到CPU信息")
	}

	// 使用CPU的VendorID和ModelName生成哈希
	hash := sha256.New()
	hash.Write([]byte(info[0].VendorID + info[0].ModelName))
	hashSum := hash.Sum(nil)

	// 取哈希值的前8位作为ID
	id := hex.EncodeToString(hashSum)[:8]

	return id, nil
}

func startLocalApi() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sysinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		hwid, err := GetHardwareID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sysinfo := map[string]string{
			"hwid":    hwid,
			"version": version,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sysinfo)
	})

	mux.HandleFunc("/api/v1/sysstatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := map[string]bool{
			"simRunning": simRunning,
			"pubRunning": pubRunning,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	apiServer = &http.Server{
		Addr:    ":7780",
		Handler: mux,
	}

	go func() {
		log.Println("Starting local API server on :7780")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start API server: %v", err)
		}
	}()
}

func stopLocalApi() {
	if apiServer != nil {
		log.Println("Shutting down API server...")
		if err := apiServer.Close(); err != nil {
			log.Printf("Error closing API server: %v", err)
		}
	}
}

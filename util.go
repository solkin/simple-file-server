package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type SuccessResult struct {
	Status int         `json:"status"`
	Result interface{} `json:"result"`
}

type ErrorResult struct {
	Status      int    `json:"status"`
	Description string `json:"description"`
}

func Success(status int, data interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", GetServerName())
	w.WriteHeader(status)
	result := SuccessResult{
		Status: status,
		Result: data,
	}
	_ = json.NewEncoder(w).Encode(result)
}

func Error(status int, description string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", GetServerName())
	w.WriteHeader(status)
	result := ErrorResult{
		Status:      status,
		Description: description,
	}
	_ = json.NewEncoder(w).Encode(result)
}

func GetServerName() string {
	return "file-server/1.0 (Unix)"
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func ParseInt64(str string) int64 {
	value, _ := strconv.ParseInt(str, 10, 64)
	return value
}

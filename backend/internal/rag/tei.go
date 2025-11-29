package rag

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
)

type teiReq struct {
	Inputs []string `json:"inputs"`
}
type teiResp struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func EmbedBatch(texts []string) ([][]float32, error) {
	url := os.Getenv("EMBEDDING_URL")
	if url == "" {
		dim := 1024
		if v := os.Getenv("EMBED_DIM"); v != "" {
			// simple parse without errors fallback
			if n := atoi(v); n > 0 {
				dim = n
			}
		}
		out := make([][]float32, len(texts))
		for i, t := range texts {
			out[i] = localEmbed(t, dim)
		}
		return out, nil
	}
	req := teiReq{Inputs: texts}
	b, _ := json.Marshal(req)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r teiResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Embeddings, nil
}

func localEmbed(text string, dim int) []float32 {
	h := sha1.Sum([]byte(text))
	seed := binary.BigEndian.Uint32(h[0:4])
	// LCG parameters
	var a uint32 = 1664525
	var c uint32 = 1013904223
	m := uint32(1<<31 - 1)
	x := seed
	v := make([]float32, dim)
	for i := 0; i < dim; i++ {
		x = (a*x + c) & m
		// map to [-1,1]
		v[i] = float32(int32(x)) / float32(int32(m))
	}
	// simple L2 normalization
	var sum float64
	for i := 0; i < dim; i++ {
		sum += float64(v[i] * v[i])
	}
	if sum > 0 {
		inv := float32(1.0 / sqrt(sum))
		for i := 0; i < dim; i++ {
			v[i] *= inv
		}
	}
	return v
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func sqrt(x float64) float64 {
	// Newton's method
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 16; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}

package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type upsertReq struct {
	Vectors []struct {
		ID       string      `json:"id"`
		Values   []float32   `json:"values"`
		Metadata interface{} `json:"metadata"`
	} `json:"vectors"`
}

func PineconeUpsert(index string, vectors map[string][]float32, meta map[string]map[string]interface{}) error {
	apiKey := os.Getenv("PINECONE_API_KEY")
	host := os.Getenv("PINECONE_HOST")
	if apiKey == "" || host == "" {
		return nil
	}
	endpoint := fmt.Sprintf("%s/vectors/upsert", host)
	b := upsertReq{Vectors: make([]struct {
		ID       string      "json:\"id\""
		Values   []float32   "json:\"values\""
		Metadata interface{} "json:\"metadata\""
	}, 0, len(vectors))}
	for id, vec := range vectors {
		item := struct {
			ID       string      `json:"id"`
			Values   []float32   `json:"values"`
			Metadata interface{} `json:"metadata"`
		}{ID: id, Values: vec, Metadata: meta[id]}
		b.Vectors = append(b.Vectors, item)
	}
	payload, _ := json.Marshal(b)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type QueryReq struct {
	TopK            int         `json:"topK"`
	Vector          []float32   `json:"vector"`
	Namespace       string      `json:"namespace,omitempty"`
	Filter          interface{} `json:"filter,omitempty"`
	IncludeMetadata bool        `json:"includeMetadata"`
}

type QueryResp struct {
	Matches []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"matches"`
}

func PineconeQueryTopK(vec []float32, topK int) (*QueryResp, error) {
	host := os.Getenv("PINECONE_HOST")
	apiKey := os.Getenv("PINECONE_API_KEY")
	if host == "" || apiKey == "" || len(vec) == 0 {
		return nil, nil
	}
	reqBody := QueryReq{TopK: topK, Vector: vec, IncludeMetadata: true}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", host+"/query", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out QueryResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func PineconeDeleteAll() error {
	host := os.Getenv("PINECONE_HOST")
	apiKey := os.Getenv("PINECONE_API_KEY")
	if host == "" || apiKey == "" {
		return nil
	}
	body := map[string]interface{}{"deleteAll": true}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", host+"/vectors/delete", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func PineconeDeleteByIDs(ids []string) error {
	host := os.Getenv("PINECONE_HOST")
	apiKey := os.Getenv("PINECONE_API_KEY")
	if host == "" || apiKey == "" || len(ids) == 0 {
		return nil
	}
	body := map[string]interface{}{"ids": ids}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", host+"/vectors/delete", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

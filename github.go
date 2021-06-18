package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type service struct {
	Username string
	Password string
}

type RefResponse struct {
	Ref    string `json:"ref"`
	NodeID string `json:"node_id"`
	URL    string `json:"url"`
	Object struct {
		Type string `json:"type"`
		Sha  string `json:"sha"`
		URL  string `json:"url"`
	} `json:"object"`
}

type TagResponse struct {
	NodeID  string `json:"node_id"`
	Tag     string `json:"tag"`
	Sha     string `json:"sha"`
	URL     string `json:"url"`
	Message string `json:"message"`
	Tagger  struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Date  time.Time `json:"date"`
	} `json:"tagger"`
	Object struct {
		Type string `json:"type"`
		Sha  string `json:"sha"`
		URL  string `json:"url"`
	} `json:"object"`
	Verification struct {
		Verified  bool        `json:"verified"`
		Reason    string      `json:"reason"`
		Signature interface{} `json:"signature"`
		Payload   interface{} `json:"payload"`
	} `json:"verification"`
}

type Service interface {
	CreateRef(gitURL, branch, commitHash string) (*RefResponse, error)
	CreateTag(gitURL, branch, commitHash string) (*TagResponse, error)
}

func NewService(username, password string) Service {
	return &service{
		Username: username,
		Password: password,
	}
}

func (s *service) CreateRef(gitURL, branch, commitHash string) (*RefResponse, error) {
	str, _ := json.Marshal(map[string]string{"ref": "refs/tags/" + branch, "sha": commitHash})

	byteReq := strings.NewReader(string(str))

	req, err := http.NewRequest(http.MethodPost, gitURL, byteReq)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")

	req.SetBasicAuth(s.Username, s.Password)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	refResponse := new(RefResponse)

	if err := json.NewDecoder(res.Body).Decode(&refResponse); err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create ref failed with error %s", res.Status)
	}

	return refResponse, nil
}

func (s *service) CreateTag(gitURL, branch, commitHash string) (*TagResponse, error) {
	str, _ := json.Marshal(map[string]string{"tag": branch, "object": commitHash, "type": "commit", "message": "create tag " + branch})

	byteReq := strings.NewReader(string(str))

	req, err := http.NewRequest(http.MethodPost, gitURL, byteReq)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3+json")

	req.SetBasicAuth(s.Username, s.Password)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	tagResponse := new(TagResponse)

	if err := json.NewDecoder(res.Body).Decode(&tagResponse); err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create tag failed with error %s", res.Status)
	}

	return tagResponse, nil
}

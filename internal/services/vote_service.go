package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type poll struct {
	Title       string `json:"name"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type createPollResponseData struct {
	ID int `json:"id"`
}

type service struct {
	client  *http.Client
	baseURL string
}

type VoteService interface {
	CreatePoll(title, description string, dueDate time.Time) (int, error)
}

func NewVoteService(baseURL string) VoteService {
	return &service{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (s *service) CreatePoll(title, description string, dueDate time.Time) (int, error) {
	jsonData, err := json.Marshal(poll{
		Title:       title,
		Description: description,
		DueDate:     dueDate.Format("2006-01-02T15:04:05"),
	})
	if err != nil {
		return 0, err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.baseURL, "poll"), bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	request.Header.Add("Content-Type", "application/json; charset=utf-8")

	fmt.Println(request.URL, string(jsonData))
	response, err := s.client.Do(request)
	if err != nil {
		return 0, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	responseData := new(createPollResponseData)
	if err := json.Unmarshal(responseBody, responseData); err != nil {
		return 0, err
	}

	return responseData.ID, nil
}

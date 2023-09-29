package services

import (
	"access_governance_system/internal/db/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type poll struct {
	Title       string `json:"name"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type Vote struct {
	UserID int64  `json:"user_id"`
	Option string `json:"option"`
}

type service struct {
	client  *http.Client
	baseURL string
}

type VoteService interface {
	CreatePoll(title, description string, dueDate time.Time) (models.Poll, error)
	GetVotes(pollID int) ([]Vote, error)
}

func NewVoteService(baseURL string) VoteService {
	return &service{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (s *service) CreatePoll(title, description string, dueDate time.Time) (models.Poll, error) {
	jsonData, err := json.Marshal(poll{
		Title:       title,
		Description: description,
		DueDate:     dueDate.Format("2006-01-02T15:04:05"),
	})
	if err != nil {
		return models.Poll{}, err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.baseURL, "poll"), bytes.NewBuffer(jsonData))
	if err != nil {
		return models.Poll{}, err
	}

	request.Header.Add("Content-Type", "application/json; charset=utf-8")

	fmt.Println(request.URL, string(jsonData))
	response, err := s.client.Do(request)
	if err != nil {
		return models.Poll{}, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return models.Poll{}, err
	}

	responseData := new(models.Poll)
	if err := json.Unmarshal(responseBody, responseData); err != nil {
		return models.Poll{}, err
	}

	return *responseData, nil
}

func (s *service) GetVotes(pollID int) ([]Vote, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", s.baseURL, "vote"), nil)
	if err != nil {
		return nil, err
	}

	query := request.URL.Query()
	query.Add("poll_id", strconv.Itoa(pollID))
	request.URL.RawQuery = query.Encode()

	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	responseData := make([]Vote, 0)
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

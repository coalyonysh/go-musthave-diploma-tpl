package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"coalyonysh/go-musthave-diploma-tpl/internal/models"
)

type AccrualService struct {
	baseURL string
	client  *http.Client
}

func NewAccrualService(baseURL string) *AccrualService {
	return &AccrualService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *AccrualService) GetOrderAccrual(orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // Заказ не зарегистрирован
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accrual models.AccrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
		return nil, err
	}

	return &accrual, nil
}

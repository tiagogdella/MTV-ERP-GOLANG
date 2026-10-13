package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//Criando erros
var (
	ErrInvalidCEP  = errors.New("CEP inválido")
	ErrNotFound    = errors.New("CEP não encontrado")
	ErrTimeout     = errors.New("ViaCEP não respondeu a tempo")
	ErrUnavailable = errors.New("ViaCEP indisponível")
)

var cepFormat = regexp.MustCompile(`^[0-9]{8}$`)

type Address struct {
	CEP          string `json:"cep"`
	Street       string `json:"logradouro"`
	Neighborhood string `json:"bairro"`
	City         string `json:"localidade"`
	State        string `json:"uf"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

func (c *Client) Lookup(ctx context.Context, cep string) (*Address, error) {
	cep = strings.ReplaceAll(cep, "-", "")
	if !cepFormat.MatchString(cep) {
		return nil, ErrInvalidCEP
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	url := fmt.Sprintf("%s/ws/%s/json/", c.baseURL, cep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("montar requisição ao ViaCEP: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrInvalidCEP
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}

	var addr Address
	if err := json.NewDecoder(resp.Body).Decode(&addr); err != nil {
		return nil, fmt.Errorf("%w: resposta inválida: %v", ErrUnavailable, err)
	}
	if addr.CEP == "" {
		return nil, ErrNotFound
	}

	return &addr, nil
}

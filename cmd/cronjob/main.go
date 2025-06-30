package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/carinfinin/keeper/internal/clientcfg"
	"github.com/carinfinin/keeper/internal/fingerprint"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/models"
	"io"
	"net/http"
	"os"
)

type Worker struct {
	config  *clientcfg.Config
	service *service.KeeperService
}

func NewWorker() (*Worker, error) {
	cfg, err := clientcfg.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("Ошибка загрузки конфига: %v\n", err)
	}

	s, err := service.NewClientService(cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации: %v", err)
	}

	return &Worker{
		config:  cfg,
		service: s,
	}, nil
}

func main() {
	// Получаем локальные изменения
	w, err := NewWorker()
	if err != nil {
		fmt.Printf("error newworker: %v\n", err)
		os.Exit(1)
	}

	b, err := w.request(w.config, http.MethodGet, "/api/last_sync", nil)
	if err != nil {
		fmt.Printf("Ошибка запроса last_sync: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("lchange")
	fmt.Println(b)

	var lchange models.LastSync
	err = json.Unmarshal(b, &lchange)
	if err != nil {
		fmt.Printf("ошибка обработки last_sync: %v\n", err)
		os.Exit(1)
	}

	localChanges, err := w.service.GetLastChanges(context.Background(), lchange.Update)
	if err != nil {
		fmt.Printf("ошибка получения GetLastChanges: %v\n", err)
		os.Exit(1)
	}

	// Отправляем на сервер
	serverChanges, err := w.PushLastChanges(context.Background(), localChanges)
	if err != nil {
		fmt.Printf("ошибка отправки PushLastChanges: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("serverChanges")
	fmt.Println(serverChanges)

	//// Применяем серверные изменения
	err = w.service.MergeLastChanges(context.Background(), serverChanges)
	if err != nil {
		fmt.Printf("ошибка сохранения MergeLastChanges: %v\n", err)
		os.Exit(1)
	}
}

func (w *Worker) request(cfg *clientcfg.Config, methodHTTP, pathMethod string, body []byte) ([]byte, error) {

	tokens, err := w.service.GetTokens(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ошибка получения токенов: %v", err)
	}

	req, err := http.NewRequest(methodHTTP, cfg.BaseURL+pathMethod, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	fp := fingerprint.Get()
	deviceID := fp.GenerateHash()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.Access)
	req.Header.Add("User-Agent", fmt.Sprintf("Kepper/%s (*%s)", cfg.Version, deviceID))

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		newTokens, err := w.service.RefreshTokens(context.Background(), tokens.Refresh)
		if err != nil {
			return nil, fmt.Errorf("ошибка обновления токена: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+newTokens.Access)

		response, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ошибка повторного запроса: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("недействительный токен после обновления")
		}
	}

	if response.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("ошибка сервера: %s, тело ответа: %s", response.Status, string(bodyBytes))
	}
	return io.ReadAll(response.Body)
}

func (w *Worker) PushLastChanges(ctx context.Context, items []*models.Item) ([]*models.Item, error) {
	res := make([]*models.Item, 0)

	b, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	buf, err := w.request(w.config, http.MethodPost, "/api/item", b)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

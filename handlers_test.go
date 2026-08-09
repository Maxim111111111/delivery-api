package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestGetProductByID(t *testing.T) {
	variants := []struct {
		name     string
		id       string
		status   int
		wantName string
	}{
		{"существующая позиция", "1", 200, "Iphone 15 Pro Max"},
		{"несуществующая позиция", "999", 404, ""},
		{"невалидная позиция", "abc", 400, ""},
	}

	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/products/"+variant.id, nil)
			request.SetPathValue("id", variant.id)
			recorder := httptest.NewRecorder()

			GetProductByID(recorder, request)
			if recorder.Code != variant.status {
				t.Fatalf("Ошибка. Невалидный статус код: %d. Ожидался статус код: %d", recorder.Code, variant.status)
			}

			if variant.wantName == "" {
				return // ← ошибочный случай, тело не проверяем, выходим
			}

			var got Product
			err := json.NewDecoder(recorder.Body).Decode(&got)
			if err != nil {
				t.Fatalf("Ошибка при парсинге json: %v", err)
			}
			if got.Name != variant.wantName {
				t.Errorf("Ошибка. Невалидное поле Name: %s. Ожидалось значение поля %s", got.Name, variant.wantName)
			}
		})

	}
}

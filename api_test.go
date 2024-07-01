package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestPingRoutes(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestClientBalance(t *testing.T) {
	router := setupRouter()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error starting mock DB: %s", err)
	}
	models.SetDB(db)
	defer db.Close()

	args := map[string]any{"kiraninumber": "07003100000", "credit": float64(100)}
	mock.ExpectQuery("SELECT useralias,credit FROM cc_card WHERE useralias = ?").
		WithArgs(args["kiraninumber"]).
		WillReturnRows(sqlmock.NewRows([]string{"useralias", "credit"}).
			AddRow(args["kiraninumber"], args["credit"]))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/clientbalance?kiraninumber="+args["kiraninumber"].(string), nil)
	req.Header.Add("Authorization", env.Env.ApiKey)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := make(map[string]any)
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, body["credit"], args["credit"].(float64))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestClientRecharge(t *testing.T) {
	router := setupRouter()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error starting mock DB: %s", err)
	}
	models.SetDB(db)
	defer db.Close()

	args := map[string]any{
		"kiraninumber": "07003100000", "cardID": 1, "amount": 100, "paymentType": 0,
		"txRef": "test", "credit": float64(150), "now": AnyTime{},
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id,useralias,credit,vat FROM cc_card WHERE useralias = ?").
		WithArgs(args["kiraninumber"]).
		WillReturnRows(sqlmock.NewRows([]string{"id", "useralias", "credit", "vat"}).
			AddRow(args["cardID"], args["kiraninumber"], args["credit"].(float64)-float64(args["amount"].(int)), 0))
	mock.ExpectExec("INSERT INTO cc_logpayment").
		WithArgs(args["now"], args["amount"], args["cardID"],
			"Recharge API "+args["txRef"].(string), 1, args["paymentType"]).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO cc_logrefill").
		WithArgs(args["now"], float64(args["amount"].(int)), args["cardID"],
			"Recharge API "+args["txRef"].(string), args["paymentType"]).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE cc_card").
		WithArgs(args["credit"], args["cardID"]).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE cc_logpayment").
		WithArgs(1, args["cardID"]).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("/api/clientrecharge?kiraninumber=%s&amount=%d&txRef=%s",
			args["kiraninumber"], args["amount"].(int), args["txRef"]),
		nil)
	req.Header.Add("Authorization", env.Env.ApiKey)
	router.ServeHTTP(w, req)

	t.Log(w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
	body := make(map[string]any)
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, body["credit"], args["credit"].(float64))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

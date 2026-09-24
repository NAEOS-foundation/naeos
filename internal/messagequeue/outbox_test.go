// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package messagequeue

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresOutboxStoreRejectsNilDB(t *testing.T) {
	if _, err := NewPostgresOutboxStore(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func TestPostgresOutboxStoreMigrate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewPostgresOutboxStore(db)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS naeos_message_outbox").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOutboxStoreEnqueue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewPostgresOutboxStore(db)
	if err != nil {
		t.Fatal(err)
	}
	msg := &Message{ID: "msg-1", IdempotencyKey: "idem-1", Topic: "build", Payload: map[string]any{"spec": "demo"}}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO naeos_message_outbox")).WithArgs("msg-1", "idem-1", "build", "{\"spec\":\"demo\"}", sqlmock.AnyArg(), 0, 3).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.Enqueue(context.Background(), msg); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if msg.Timestamp.IsZero() {
		t.Fatal("enqueue should assign timestamp")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOutboxStoreClaim(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewPostgresOutboxStore(db)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "idempotency_key", "topic", "payload", "created_at", "retries", "max_retries"}).
		AddRow("msg-1", "idem-1", "build", "{\"spec\":\"demo\"}", time.Unix(100, 0).UTC(), 1, 3)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, idempotency_key, topic, payload, created_at, retries, max_retries")).WithArgs(5).WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE naeos_message_outbox")).WithArgs("worker-1", sqlmock.AnyArg(), "msg-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	messages, err := store.Claim(context.Background(), "worker-1", 5, time.Minute)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected one claimed message, got %d", len(messages))
	}
	if messages[0].ID != "msg-1" || messages[0].Topic != "build" {
		t.Fatalf("unexpected message: %#v", messages[0])
	}
	payload, ok := messages[0].Payload.(map[string]any)
	if !ok || payload["spec"] != "demo" {
		t.Fatalf("unexpected payload: %#v", messages[0].Payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOutboxStoreAckAndFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewPostgresOutboxStore(db)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE naeos_message_outbox")).WithArgs("msg-1", "worker-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.Ack(context.Background(), "msg-1", "worker-1"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	retryAt := time.Unix(200, 0).UTC()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE naeos_message_outbox")).WithArgs(retryAt, "temporary failure", "msg-1", "worker-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.Fail(context.Background(), "msg-1", "worker-1", retryAt, "temporary failure"); err != nil {
		t.Fatalf("fail: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresOutboxStoreValidation(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store, err := NewPostgresOutboxStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Enqueue(context.Background(), nil); err == nil {
		t.Fatal("expected nil message error")
	}
	if err := store.Enqueue(context.Background(), &Message{ID: "m"}); err == nil {
		t.Fatal("expected missing topic error")
	}
	if _, err := store.Claim(context.Background(), "", 1, time.Second); err == nil {
		t.Fatal("expected missing worker id error")
	}
	if err := store.Ack(context.Background(), "", ""); err == nil {
		t.Fatal("expected missing message id error")
	}
	if err := store.Fail(context.Background(), "", "", time.Now(), "x"); err == nil {
		t.Fatal("expected missing message id error")
	}
}

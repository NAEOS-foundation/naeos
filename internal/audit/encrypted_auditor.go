package audit

import (
	"encoding/json"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

type EncryptedAuditor struct {
	inner      Auditor
	passphrase string
	mu         sync.Mutex
}

func NewEncryptedAuditor(inner Auditor, passphrase string) *EncryptedAuditor {
	return &EncryptedAuditor{
		inner:      inner,
		passphrase: passphrase,
	}
}

func (e *EncryptedAuditor) Log(event AuditEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshal event")
	}

	encrypted, err := securityext.EncryptConfig(data, e.passphrase)
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "encrypt event")
	}

	encEvent := AuditEvent{
		ID:        event.ID,
		Timestamp: event.Timestamp,
		Action:    "audit.encrypted",
		Resource:  "audit",
		Status:    "success",
		Details:   encrypted,
		Metadata: map[string]string{
			"encrypted": "true",
			"algorithm": "AES-256-GCM",
			"original_action": event.Action,
			"original_resource": event.Resource,
		},
	}

	return e.inner.Log(encEvent)
}

type DecryptedReader struct {
	inner      *MemoryAuditor
	passphrase string
}

func NewDecryptedReader(inner *MemoryAuditor, passphrase string) *DecryptedReader {
	return &DecryptedReader{
		inner:      inner,
		passphrase: passphrase,
	}
}

func (d *DecryptedReader) Events() ([]AuditEvent, error) {
	raw := d.inner.Events()
	var result []AuditEvent

	for _, e := range raw {
		if e.Details != "" && e.Metadata["encrypted"] == "true" {
			plaintext, err := securityext.DecryptConfig(e.Details, d.passphrase)
			if err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "decrypt event %s", e.ID)
			}
			var decrypted AuditEvent
			if err := json.Unmarshal(plaintext, &decrypted); err != nil {
				return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "unmarshal decrypted event %s", e.ID)
			}
			result = append(result, decrypted)
		} else {
			result = append(result, e)
		}
	}

	return result, nil
}

func NewEncryptedFileAuditor(homeDir, passphrase string) (*EncryptedAuditor, error) {
	inner, err := NewFileAuditor(homeDir)
	if err != nil {
		return nil, err
	}
	return NewEncryptedAuditor(inner, passphrase), nil
}



package main

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
)

type Session struct {
	IP        string
	CreatedAt time.Time
	TTL       time.Duration
}

var sessions = sync.Map{}

func init() {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			var cleaned int

			sessions.Range(func(key, value any) bool {
				session := value.(*Session)
				if time.Since(session.CreatedAt) > session.TTL {
					sessions.Delete(key)
					cleaned++
				}
				return true
			})
			log.Printf("Cleaned %d sessions", cleaned)
		}
	}()
}

func ValidateSessionID(id string) error {
	if len(id) < 6 || len(id) > 10 {
		return errors.New("invalid session ID length")
	}

	for _, c := range id {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return errors.New("session ID must be alphanumeric")
		}
	}

	return nil
}

// GetSession gets or creates a session
// after validating the session ID and the lifetime, if any
func GetSession(parameters map[string]string) (*Session, error) {
	id, ok := parameters["session"]
	if !ok {
		return nil, nil
	}

	if err := ValidateSessionID(id); err != nil {
		return nil, err
	}

	session, ok := sessions.Load(id)
	if !ok {
		ttl, ok := parameters["lifetime"]
		if !ok {
			ttl = "10"
		}

		ttlInt, err := strconv.Atoi(ttl)
		if err != nil {
			return nil, err
		}

		if ttlInt < 0 || ttlInt > 60 {
			return nil, errors.New("invalid session lifetime")
		}

		return CreateSession(id, ttlInt)
	}

	return session.(*Session), nil
}

func CreateSession(id string, ttl int) (*Session, error) {
	ip, err := GenerateIPv6FromCIDR(cidr)
	if err != nil {
		return nil, err
	}

	session := Session{
		IP:        ip,
		CreatedAt: time.Now(),
		TTL:       time.Duration(ttl) * time.Minute,
	}

	sessions.Store(id, &session)

	return &session, nil
}

// Package cache fornece um cache em memória, thread-safe, com expiração por TTL.
// É adequado ao deploy de instância única (Render free): simples e sem
// dependências externas. Em escala horizontal a invalidação não se propaga
// entre instâncias — nesse cenário seria necessário um backend compartilhado.
package cache

import (
	"strings"
	"sync"
	"time"
)

type item struct {
	value     any
	expiresAt time.Time
}

// Store guarda valores arbitrários por chave, com expiração por TTL e
// invalidação por prefixo de chave (usado para flush por usuário).
type Store struct {
	mu   sync.RWMutex
	data map[string]item
}

// New cria um Store vazio e pronto para uso.
func New() *Store {
	return &Store{data: make(map[string]item)}
}

// Get retorna o valor associado à chave quando presente e não expirado.
// Expiração é preguiçosa: entradas vencidas são removidas ao serem lidas.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	it, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expiresAt) {
		s.mu.Lock()
		// Revalida sob lock de escrita: a entrada pode ter sido renovada entre
		// a leitura e a aquisição do lock.
		if cur, ok := s.data[key]; ok && time.Now().After(cur.expiresAt) {
			delete(s.data, key)
		}
		s.mu.Unlock()
		return nil, false
	}
	return it.value, true
}

// Set grava o valor sob a chave com o TTL informado.
func (s *Store) Set(key string, value any, ttl time.Duration) {
	s.mu.Lock()
	s.data[key] = item{value: value, expiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
}

// FlushPrefix remove todas as entradas cuja chave começa com o prefixo dado.
// Usado para invalidar o cache de um usuário específico (prefixo "userID|").
func (s *Store) FlushPrefix(prefix string) {
	s.mu.Lock()
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			delete(s.data, k)
		}
	}
	s.mu.Unlock()
}

// Flush remove todas as entradas do store.
func (s *Store) Flush() {
	s.mu.Lock()
	s.data = make(map[string]item)
	s.mu.Unlock()
}

package cache

import (
	"testing"
	"time"
)

func TestStore_GetHitAndMiss(t *testing.T) {
	s := New()
	if _, ok := s.Get("k"); ok {
		t.Fatal("esperado miss em store vazio")
	}
	s.Set("k", 42, time.Minute)
	v, ok := s.Get("k")
	if !ok || v.(int) != 42 {
		t.Fatalf("esperado hit com 42, obtido %v ok=%v", v, ok)
	}
}

func TestStore_Expiration(t *testing.T) {
	s := New()
	s.Set("k", "v", 10*time.Millisecond)
	if _, ok := s.Get("k"); !ok {
		t.Fatal("esperado hit antes de expirar")
	}
	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Get("k"); ok {
		t.Fatal("esperado miss após expirar por TTL")
	}
}

func TestStore_FlushPrefix(t *testing.T) {
	s := New()
	s.Set("user-1|a", 1, time.Minute)
	s.Set("user-1|b", 2, time.Minute)
	s.Set("user-2|a", 3, time.Minute)
	s.Set("|public", 4, time.Minute)

	s.FlushPrefix("user-1|")

	if _, ok := s.Get("user-1|a"); ok {
		t.Error("user-1|a deveria ter sido invalidado")
	}
	if _, ok := s.Get("user-1|b"); ok {
		t.Error("user-1|b deveria ter sido invalidado")
	}
	if _, ok := s.Get("user-2|a"); !ok {
		t.Error("user-2|a NÃO deveria ser afetado pelo flush de user-1")
	}
	if _, ok := s.Get("|public"); !ok {
		t.Error("entrada pública NÃO deveria ser afetada pelo flush de user-1")
	}
}

func TestStore_Flush(t *testing.T) {
	s := New()
	s.Set("a", 1, time.Minute)
	s.Set("b", 2, time.Minute)
	s.Flush()
	if _, ok := s.Get("a"); ok {
		t.Error("a deveria ter sido removido pelo Flush")
	}
	if _, ok := s.Get("b"); ok {
		t.Error("b deveria ter sido removido pelo Flush")
	}
}

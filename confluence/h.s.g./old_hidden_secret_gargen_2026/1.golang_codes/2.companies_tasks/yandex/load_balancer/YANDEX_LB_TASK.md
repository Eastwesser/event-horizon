# YANDEX LOAD BALANCER

```go
/*
ТАСКА 1
За backend интерфейсом реализация.
backend impl уже реализован
нужно реализовать балансировщик Balancer между миросервисами, в приоритете наименее нагруженный экземпляр
*/

package main

import (
  "context"
  "sync"
)

type Request interface{}

type Response interface{}

type BackendImpl struct {
  addr string
}

type Backend interface {
  Invoke(ctx context.Context, req Request) (Response, error)
}

func (b *BackendImpl) Invoke(ctx context.Context, req Request) (Response, error) {
  return nil, nil
}

var _ Backend = &BackendImpl{}

// addr содержит ip:port конкретного экземпляра
func NewBackend(addr string) *BackendImpl

type Balancer struct {
  Backs     []*BackendImpl
  req_count int
  mu        sync.Mutex
}

var _ Backend = &Balancer{}

// addrs содержат адреса всех балансируемых экземпляров
func NewBalancer(addrs []string) *Balancer {
  b := Balancer{Backs: make([]*BackendImpl, len(addrs))}
  for i, addr := range b.Backs {
    back := NewBackend(addrs[i])
    *addr = *back
  }
  return &b
}

func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
  b.mu.Lock()
  b.req_count++
  back_num := b.req_count % len(b.Backs)
  b.mu.Unlock()

  resp, err := b.Backs[back_num].Invoke(ctx, req)
  if err != nil {
    return nil, err
  }
  return resp, err
}

/*
    ТАСКА 2
    Есть экземпляры со сбоями, и если сбоят - исключение из балансировки.
    Если 3 три ретрая было,  мы к нему не обращаемся.
*/
```

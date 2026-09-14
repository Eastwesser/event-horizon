package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

/*
	https://ya.cc/t/N32jhpbz956niK

	Есть приложение с микросервисной архитектурой.
	Микросервис можно абстрагировать с помощью интерфейса Backend.
	Для доступа к одному экземпляру микросервиса можно использовать
	тип BackendImpl, который уже реализован.

	Для каждого микросервиса есть несколько десятков запущенных
	экземпляров, каждый из которых доступен по своему адресу addr.
	Однако отдельные экземпляры микросервиса ненадежны:
	они могут падать, быть недоступными либо перегруженными.

	Поэтому вам нужно реализовать тип Balancer, который также реализует
	интерфейс Backend и осуществляет client-side балансировку нагрузки
	между экземплярами микросервиса.
*/

type Request interface{}

type Response interface{}

type Backend interface {
	Invoke(ctx context.Context, req Request) (Response, error)
}

var _ Backend = &BackendImpl{}

// addr содержит ip:port конкретного экземпляра
func NewBackend(addr string) *BackendImpl {
	return &BackendImpl{}
}

type Balancer struct {
	curr       int64
	backends   []Backend
	idxBackend int
	mu         sync.Mutex
}

var _ Backend = &Balancer{}

// addrs содержат адреса всех балансируемых экземпляров
func NewBalancer(addrs []string) *Balancer {
	backends := make([]Backend, len(addrs))

	for i, addr range addrs {
		    backends[i] = NewBackend(addr)
		}

	return &Balancer{
		    backends: backends,
		    idxBackend: len(backends)
	}
}


func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
	b.mu.Lock()
	b := backends[b.curr]
	curr := b.curr
	b.curr = (b.curr++) % b.idxBackend
	b.mu.Unlock()

	var resp Response
	var err error
	const maxRetry = 2

	// ДЗ - STRATEGY паттерн изать, чтобы в балансировку мы возвращали функцию, linked-list strategy

	// нужен ли тут бесконечный цикл? Если retry больше двух , то выходим,  или по контексту выходим? Через next а не через  current (через связный список нужно)
	for range maxRetry {
		err = nil
		if countRetry == maxRetry {
			return nil, errors.New("ошибка")
		}
		ctxt, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err = b.backends[curr].Invoke(ctxt, req)
		if err == nil {
			break
		}

		curr = (b.curr++) % b.idxBackend
	}

	return resp, err
}
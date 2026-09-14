package main

import "fmt"

// Target: интерфейс, который ожидает клиент
type Notifier interface{ Notify(msg string) }

// Adaptee: несовместимый по интерфейсу существующий тип
type LegacyPrinter struct{}

func (p *LegacyPrinter) PrintBytes(b []byte) { fmt.Printf("[legacy] %s\n", string(b)) }

// Adapter: приводит LegacyPrinter к Notifier
type PrinterAdapter struct{ p *LegacyPrinter }

func (a *PrinterAdapter) Notify(msg string) { a.p.PrintBytes([]byte(msg)) }

// Client, которому нужен Notifier
func SendAlert(n Notifier, text string) { n.Notify(text) }

func main() {
	lp := &LegacyPrinter{}
	adapter := &PrinterAdapter{p: lp}
	SendAlert(adapter, "Adapter works")
}

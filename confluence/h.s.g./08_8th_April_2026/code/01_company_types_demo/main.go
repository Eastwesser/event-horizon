// HR-якорь: типы компаний + чеклист легенды (из 08_hr_meeting).
package main

import "fmt"

type CompanyKind int

const (
	ProductNonIT CompanyKind = iota
	ITProduct
	Outsourcing
	Outstaffing
)

func (k CompanyKind) String() string {
	return [...]string{"продуктовая (не IT)", "IT-продукт", "аутсорсинг", "аутстаффинг"}[k]
}

type LegendCheck struct {
	HasRealSite     bool
	OKVEDIsSoftware bool
	NotReorgOnly    bool
	LiquidationOK   bool // дата/статус проверены, если ликвидирована
}

func (c LegendCheck) OK() bool {
	return c.HasRealSite && c.OKVEDIsSoftware && c.NotReorgOnly && c.LiquidationOK
}

func main() {
	kinds := []CompanyKind{ProductNonIT, ITProduct, Outsourcing, Outstaffing}
	fmt.Println("типы компаний на HR:")
	for _, k := range kinds {
		fmt.Println(" -", k)
	}

	candidate := LegendCheck{
		HasRealSite: true, OKVEDIsSoftware: true, NotReorgOnly: true, LiquidationOK: true,
	}
	fmt.Println("легенда проходит чеклист:", candidate.OK())
}

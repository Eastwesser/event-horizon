

Разбор Java-кода с точки зрения Go-разработчика
1. Структура и поля
java
private String name;
private String surname;
public String address;
public String phone;
Что в Go:

go
type Person struct {
    name    string
    surname string
    Address string // публичное поле
    Phone   string // публичное поле
}
Проблемы:

Смешение приватных и публичных полей — плохая практика. В Go это решается регистром первой буквы, но смешивать name и Address в одном типе — странно.

Externalizable и Serializable — в Go нет прямой аналогии, но есть encoding/gob, encoding/json и encoding/xml. Для XML в Go используются теги:

go
type Person struct {
    Name    string `xml:"name"`
    Surname string `xml:"surname"`
    Address string `xml:"address"`
    Phone   string `xml:"phone"`
}
2. Конструктор
java
public Person(String name, String surname) {
    this.name = name;
    this.surname = surname;
}
В Go:

go
func NewPerson(name, surname string) *Person {
    return &Person{name: name, surname: surname}
}
Замечание: В Go принято возвращать указатель, если структура большая или должна изменяться. Здесь — ок.

3. hashCode() возвращает константу
java
@Override
public int hashCode() {
    return 1;
}
Это катастрофа.

В Java это нарушает контракт hashCode(): все объекты попадут в один бакет HashMap, что превратит HashMap в LinkedList — O(n) вместо O(1).

В Go аналог — использование структуры как ключа мапы. Если бы вы сделали так:

go
m := map[Person]int{}
Go использует все поля структуры для хеша. Если бы вы захотели "зафиксировать" хеш, это было бы невозможно без обертки.

Вывод: В Go такой проблемы нет, но если бы вы реализовывали свой Hasher — это было бы ошибкой.

4. equals() — корректный, но многословный
java
@Override
public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    Person person = (Person) o;
    if (name != null ? !name.equals(person.name) : person.name != null) return false;
    if (surname != null ? !surname.equals(person.surname) : person.surname != null) return false;
    return true;
}
В Go:

go
func (p Person) Equals(other Person) bool {
    return p.name == other.name && p.surname == other.surname
}
Замечание: В Go можно просто сравнить структуры через ==, если все поля сравнимы. Это короче и безопаснее.

5. clone() через super.clone()
java
@Override
public Person clone() {
    try {
        return (Person) super.clone();
    } catch (CloneNotSupportedException e) {
        throw new Error("Clone not supported.");
    }
}
В Go:

go
func (p Person) Clone() Person {
    return p // копирование значения
}

// или для указателя:
func (p *Person) Clone() *Person {
    if p == nil {
        return nil
    }
    clone := *p
    return &clone
}
Замечание: В Go нет CloneNotSupportedException, и клонирование обычно делается вручную или через copy(). Это безопаснее.

6. writeExternal / readExternal — порядок полей не совпадает!
java
@Override
public void writeExternal(ObjectOutput out) throws IOException {
    out.writeObject(name);
    out.writeObject(surname);
    out.writeObject(phone);   // phone
    out.writeObject(address); // address
}

@Override
public void readExternal(ObjectInput in) throws IOException, ClassNotFoundException {
    name = (String) in.readObject();
    surname = (String) in.readObject();
    address = (String) in.readObject(); // address
    phone = (String) in.readObject();   // phone
}
Это критическая ошибка!

При сериализации порядок: name, surname, phone, address.
При десериализации порядок: name, surname, address, phone.

Это приведет к тому, что address и phone поменяются местами. В Go такое невозможно, потому что сериализация через gob/json использует имена полей, а не порядок.

В Go:

go
type Person struct {
    Name    string `json:"name"`
    Surname string `json:"surname"`
    Address string `json:"address"`
    Phone   string `json:"phone"`
}

// Сериализация:
data, _ := json.Marshal(p)
// Десериализация:
var p2 Person
json.Unmarshal(data, &p2)
Порядок не важен — все работает по именам.

7. Externalizable vs Serializable
В Java Externalizable требует ручной реализации writeExternal/readExternal, что увеличивает риск ошибок (см. пункт 6).

В Go:

encoding/gob — автоматическая сериализация.

encoding/json — автоматическая.

encoding/xml — автоматическая.

Если нужна кастомная сериализация — реализуйте интерфейсы MarshalJSON/UnmarshalJSON или GobEncode/GobDecode, но это редко нужно.

8. XML-аннотации
java
@XmlAccessorType(XmlAccessType.PROPERTY)
@XmlType(name = "Person", propOrder = {
    "name",
    "surname",
    "address",
    "phone",
})
В Go:

go
type Person struct {
    XMLName xml.Name `xml:"Person"`
    Name    string   `xml:"name"`
    Surname string   `xml:"surname"`
    Address string   `xml:"address"`
    Phone   string   `xml:"phone"`
}
Использование XmlAccessType.PROPERTY означает, что XML строится по геттерам, а не по полям. В Go такого нет — только по полям или через кастомные маршалеры.

Итог: что бы я сказал на код-ревью как Go-разработчик
hashCode() == 1 — грубая ошибка, ломает производительность HashMap.

Порядок в writeExternal/readExternal не совпадает — критический баг, поля перепутаются.

Смешение private и public полей — плохой дизайн.

clone() через super.clone() — в Go это не нужно, копирование значений проще.

Externalizable — избыточен, в Go есть автоматическая сериализация.

equals() — в Go можно сравнить структуры через ==.

Как бы это выглядело на Go:

go
package main

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
)

type Person struct {
    XMLName xml.Name `xml:"Person"`
    Name    string   `xml:"name" json:"name"`
    Surname string   `xml:"surname" json:"surname"`
    Address string   `xml:"address" json:"address"`
    Phone   string   `xml:"phone" json:"phone"`
}

func NewPerson(name, surname string) *Person {
    return &Person{Name: name, Surname: surname}
}

func (p Person) Equals(other Person) bool {
    return p.Name == other.Name && p.Surname == other.Surname
}

func (p Person) Clone() Person {
    return p
}

func main() {
    p := NewPerson("John", "Doe")
    p.Address = "123 Main St"
    p.Phone = "555-1234"

    // JSON
    data, _ := json.Marshal(p)
    fmt.Println(string(data))

    // XML
    xmlData, _ := xml.Marshal(p)
    fmt.Println(string(xmlData))
}

====


Таблица PERSON

ID	NAME
1	alice
2	john
3	bill
4	mike
5	sara
6	helen
7	steven
8	garry
9	kile
10	liza
Таблица PHONE

ID	PERSON_ID	NUMBER
101	2	7776677
102	4	1111111
103	6	6666666
104	8	5553535
105	10	0000111
SQL-запрос:

sql
select
  NAME, NUMBER
from PERSON p
left join PHONE ph
  on p.ID=ph.PERSON_ID
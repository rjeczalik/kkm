kkm [![GoDoc](https://godoc.org/github.com/rjeczalik/kkm?status.svg)](https://godoc.org/github.com/rjeczalik/kkm)
---

Get your ticket payment history or reverse lookup your name by your student card ID.

### cmd/kkm-history [![GoDoc](https://godoc.org/github.com/rjeczalik/kkm/cmd/kkm-history?status.svg)](https://godoc.org/github.com/rjeczalik/kkm/cmd/kkm-history)

Get your ticket payment history. Outputs `[]kkm.Ticket` in JSON, sorted by `PurchasedAt` field in increasing order. The `kkm.Ticket` is defined as:

```go
type Ticket struct {
	PurchasedAt time.Time `json:"purchased_at"`
	ExpiredAt   time.Time `json:"expires_at"`
	Type        string    `json:"type"`
	Price       int       `json:"price"`
	StudentID   int       `json:"student_id"`
	KKMID       int       `json:"kkm_id"`
}
```

Times are always in UTC. The price is a decimal with scale=2.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/kkm/cmd/kkm-history
```

*Example*

```bash
~ $ kkm-history -card UJ -id 1234567
[
	{
		"purchased_at": "2014-12-10T07:04:00Z",
		"expires_at": "2015-01-09T00:00:00Z",
		"type": "Ulgowy | Wszystkie dni tygodnia",
		"price": 4900,
		"student_id": 123456722,
		"kkm_id": 1234567890
	},
	{
		"purchased_at": "2015-01-13T07:25:00Z",
		"expires_at": "2015-02-12T00:00:00Z",
		"type": "Ulgowy | Wszystkie dni tygodnia",
		"price": 4900,
		"student_id": 123456722,
		"kkm_id": 1234567890
	}
]
```

### cmd/kkm-detail [![GoDoc](https://godoc.org/github.com/rjeczalik/kkm/cmd/kkm-history?status.svg)](https://godoc.org/github.com/rjeczalik/kkm/cmd/kkm-history)

Reverse lookup your personal details with your student card ID (`(kkm.Ticket).StudentID`) and KKM card ID (`(kkm.Ticket).KKMID`). Outputs `*kkm.Detail` in JSON. The `kkm.Detail` is defined as:

```go
type Detail struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}
```

The `Email` and `Phone` may be empty.

*Installation*

```bash
~ $ go get -u github.com/rjeczalik/kkm/cmd/kkm-detail
```

*Example*

```bash
~ $ kkm-detail -id 123456722 -kkm 1234567890
{
	"first_name": "Twoja",
	"last_name": "Stara",
	"email": "Pierze",
	"phone": "W rzece"
}
```

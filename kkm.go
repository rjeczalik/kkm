package kkm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var CityCardType = map[string]int{
	"wszib": 20,
	"agh":   21,
	"uj":    22,
	"pk":    23,
	"ue":    24,
	"ur":    25,
	"pwst":  26,
	"am":    27,
	"wse":   28,
	"aik":   29,
	"up":    30,
	"wsh":   31,
	"ka":    32,
	"wsei":  33,
	"ifj":   34,
	"if":    35,
	"ikifp": 36,
}

const history = "http://www.mpk.krakow.pl/pl/sprawdz-waznosc-biletu/index,1.ht" +
	"ml?cityCardType=%d&dateValidity=1970-01-01&identityNumber=%d&sprawdz_kkm=" +
	"Sprawd%C5%BA"

const purchase = "2006-01-02 15:04"
const expires = "2006-01-02"

func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

type Ticket struct {
	PurchasedAt time.Time `json:"purchased_at"`
	ExpiredAt   time.Time `json:"expires_at"`
	Type        string    `json:"type"`
	Price       int       `json:"price"`
	StudentID   int       `json:"student_id"`
	KKMID       int       `json:"kkm_id"`
}

type ticketSlice []Ticket

func (t ticketSlice) Len() int           { return len(t) }
func (t ticketSlice) Less(i, j int) bool { return t[i].PurchasedAt.Before(t[j].PurchasedAt) }
func (t ticketSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func History(cardType string, studentID int) ([]Ticket, error) {
	typ, ok := CityCardType[strings.ToLower(cardType)]
	if !ok {
		return nil, errors.New("unknown card type")
	}
	res, err := http.Get(fmt.Sprintf(history, typ, studentID))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}
	return parse(res.Body)
}

func b(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("b").Text())
}

func parse(r io.Reader) (t []Ticket, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var n int
	doc.Find(`div[class='kkm-card'] > div`).Each(
		func(i int, s *goquery.Selection) {
			switch line := s.Text(); {
			case strings.Contains(line, "Rodzaj biletu:"):
				t = append(t, Ticket{Type: b(s)})
				n = len(t) - 1
			case strings.Contains(line, "Data i godzina zakupu:"):
				d, e := time.Parse(purchase, b(s))
				if e != nil {
					err = nonil(err, e)
					return
				}
				t[n].PurchasedAt = d
			case strings.Contains(line, "Numer legitymacji:"):
				x, e := strconv.Atoi(b(s))
				if e != nil {
					err = nonil(err, e)
					return
				}
				t[n].StudentID = x
			case strings.Contains(line, "Numer karty KKM:"):
				x, e := strconv.Atoi(b(s))
				if e != nil {
					err = nonil(err, e)
					return
				}
				t[n].KKMID = x
			case strings.Contains(line, "Cena:"):
				var m, l int
				x, e := fmt.Sscanf(b(s), "%d,%d zł", &m, &l)
				if e != nil || x != 2 {
					err = nonil(err, e, errors.New("invalid price string"))
					return
				}
				t[n].Price = m*100 + l
			case strings.Contains(line, "Data końca ważności:"):
				d, e := time.Parse(expires, b(s))
				if e != nil {
					err = nonil(err, e)
					return
				}
				t[n].ExpiredAt = d
			}

		},
	)
	sort.Sort(ticketSlice(t))
	return t, err
}

package kkm

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
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
	"Sprawd%%C5%%BA"

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
	if len(t) == 0 {
		return nil, errors.New("unable to find ticket history or no ticket history")
	}
	sort.Sort(ticketSlice(t))
	return t, err
}

type Detail struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

const host = "ebilet.kkm.krakow.pl"

const login = "https://ebilet.kkm.krakow.pl/ebilet/Logowanie"

const ticket = "https://ebilet.kkm.krakow.pl/ebilet/KupBilet"

func formLogin(studentID, kkmID int) url.Values {
	return url.Values{
		"CityCardTypeCode":  []string{"0"},
		"CustomerCodeStr":   []string{strconv.Itoa(studentID)},
		"CityCardNumberStr": []string{strconv.Itoa(kkmID)},
		"AcceptRegulamin":   []string{"true"},
		"AcceptDaneOsobowe": []string{"true"},
	}
}

var header = http.Header{
	"Pragma":        []string{"no-cache"},
	"Cache-Control": []string{"no-cache"},
	"User-Agent":    []string{"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.11 (KHTML like Gecko) Chrome/23.0.1271.95 Safari/537.11"},
}

var headerLogin = http.Header{
	"Content-Type":  []string{"application/x-www-form-urlencoded"},
	"Pragma":        []string{"no-cache"},
	"Cache-Control": []string{"no-cache"},
	"User-Agent":    []string{"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.11 (KHTML like Gecko) Chrome/23.0.1271.95 Safari/537.11"},
	"Origin":        []string{"https://ebilet.kkm.krakow.pl"},
	"Referer":       []string{"https://ebilet.kkm.krakow.pl/ebilet/Logowanie"},
}

var headerTicket = http.Header{
	"Pragma":        []string{"no-cache"},
	"Cache-Control": []string{"no-cache"},
	"User-Agent":    []string{"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.11 (KHTML like Gecko) Chrome/23.0.1271.95 Safari/537.11"},
	"Referer":       []string{"https://ebilet.kkm.krakow.pl/ebilet/"},
}

func client() *http.Client {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		panic(err)
	}
	// For mitmproxy + HTTPS.
	if os.Getenv("HTTP_PROXY") != "" {
		return &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				TLSHandshakeTimeout: 10 * time.Second,
			},
		}
	}
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
}

func Details(studentID, kkmID int) (*Detail, error) {
	c := client()
	req, err := http.NewRequest("GET", login, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	body := formLogin(studentID, kkmID).Encode()
	reqLogin, err := http.NewRequest("POST", login, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	reqLogin.Host = host
	reqLogin.Header = headerLogin
	reqLogin.ContentLength = int64(len(body))
	respLogin, err := c.Do(reqLogin)
	if err != nil {
		return nil, err
	}
	respLogin.Body.Close()
	if respLogin.StatusCode != 200 {
		return nil, errors.New(http.StatusText(respLogin.StatusCode))
	}
	reqTicket, err := http.NewRequest("GET", ticket, nil)
	if err != nil {
		return nil, err
	}
	reqTicket.Host = host
	reqTicket.Header = headerTicket
	respTicket, err := c.Do(reqTicket)
	if err != nil {
		return nil, err
	}
	defer respTicket.Body.Close()
	if respTicket.StatusCode != 200 {
		return nil, errors.New(http.StatusText(respTicket.StatusCode))
	}
	return parseTicket(respTicket.Body)
}

func parseTicket(r io.Reader) (*Detail, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	firstName, ok := doc.Find(`input[id='clientName']`).Attr("value")
	if !ok {
		return nil, errors.New("unable to read first name")
	}
	lastName, ok := doc.Find(`input[id='clientSurname']`).Attr("value")
	if !ok {
		return nil, errors.New("unable to read last name")
	}
	email, _ := doc.Find(`input[id='customerEmail']`).Attr("value")
	phone, _ := doc.Find(`input[id='customerPhoneNumber']`).Attr("value")
	d := &Detail{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
	}
	return d, nil
}

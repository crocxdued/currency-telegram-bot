package cbr

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html/charset"
)

type ValCurs struct {
	Valutes []struct {
		CharCode string `xml:"CharCode"`
		Value    string `xml:"Value"`
		Nominal  int    `xml:"Nominal"`
	} `xml:"Valute"`
}

type CBRClient struct {
	baseURL string
}

func New() *CBRClient {
	return &CBRClient{baseURL: "https://www.cbr.ru/scripts/XML_daily.asp"}
}

func (c *CBRClient) GetRate(ctx context.Context, from, to string) (float64, error) {
	if to != "RUB" && from != "RUB" {
		return 0, fmt.Errorf("CBR provider only supports RUB pairs")
	}

	resp, err := http.Get(c.baseURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	var data ValCurs
	if err := decoder.Decode(&data); err != nil {
		return 0, err
	}

	target := from
	if from == "RUB" {
		target = to
	}

	for _, v := range data.Valutes {
		if v.CharCode == target {
			valStr := strings.Replace(v.Value, ",", ".", 1)
			var rate float64
			fmt.Sscanf(valStr, "%f", &rate)
			res := rate / float64(v.Nominal)
			if from == "RUB" {
				return 1 / res, nil
			}
			return res, nil
		}
	}
	return 0, fmt.Errorf("currency %s not found", target)
}

func (c *CBRClient) GetName() string   { return "CBR" }
func (c *CBRClient) IsAvailable() bool { return true }

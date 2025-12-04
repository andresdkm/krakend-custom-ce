package transformers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"kraken-builder-plugins/pkg/hotels/common/models"
	"log"
	"net/http"
	"net/url"
	"path"
)

type IRequestFactory interface {
	BuildRequest() func(map[string]interface{}) func(interface{}) (interface{}, error)
}

func GetRequestFactory(provider models.Provider) (IRequestFactory, error) {
	switch provider {
	case models.ProviderTBO:
		return &TBORequestFactory{}, nil
	default:
		return nil, fmt.Errorf("no response factory found for provider: %s", provider)
	}
}

type TBORequestFactory struct {
}

func (rp *TBORequestFactory) BuildRequest() func(map[string]interface{}) func(interface{}) (interface{}, error) {
	return func(cfg map[string]interface{}) func(interface{}) (interface{}, error) {
		return func(input interface{}) (interface{}, error) {
			req, ok := input.(IRequestWrapper)
			if !ok {
				return nil, unknownTypeErr
			}
			r := modifier(req)
			checkin := req.Query().Get("checkin")
			checkout := req.Query().Get("checkout")
			enrichedRequest := map[string]interface{}{
				"CheckIn":          checkin,
				"CheckOut":         checkout,
				"HotelCodes":       "1402689,1405349,1405355,1407362,1413911,1414353,1415021,1415135,1415356,1415518,1415792,1416419,1416455,1416461,1416726,1440549,1440646,1440710,1440886,1440924,1441027,1441035,1441155,1441982,1442124,1443452,1443686,1447419,1448073,1450393,1450771,1450910,1450927,1450928,1451558,1452394,1452622,1452663,1453490,1457003,1457080,1457487,1457578,1457885,1458286,1458386,1458544,1458641,1459770,1463960,1463986,1464370,1464612,1465220,1465563,1465616,1465788,1466296,1466820,1466843,1467053,1467113,1468099,1469699,1469700,1469706,1472429,1474785,1475148,1475152,1479473,1479485,1482733,1482841,1482863,1483807,1484226,1485439,1487994,1490420,1491113,1491115,1491121,1491171,1491329,1491342,1491346,1491350,1491354,1491355,1491912,1492068,1492074,1492276,1492293,1492323,1493583,1493627,1493630,1493733",
				"GuestNationality": "US",
				"PaxRooms": []map[string]interface{}{
					{
						"Adults":       1,
						"Children":     0,
						"ChildrenAges": []int{},
					},
				},
				"ResponseTime":       18,
				"IsDetailedResponse": true,
				"Filters": map[string]interface{}{
					"Refundable": true,
					"NoOfRooms":  0,
					"MealType":   "All",
				},
			}
			newBody, err := json.Marshal(enrichedRequest)
			if err != nil {
				log.Printf("[TBO-PLUGIN] Error marshaling request: %v", err)
			}

			log.Printf("[TBO-PLUGIN] New body created, length: %d", len(newBody))

			r.body = ioutil.NopCloser(bytes.NewBuffer(newBody))
			r.method = "POST"
			r.query = url.Values{}
			r.headers = http.Header{
				"Content-Type": []string{"application/json"},
			}
			log.Printf("[TBO-PLUGIN] Sending request: %s", r.method)
			log.Printf("[TBO-PLUGIN] Request body: %s", string(newBody))
			return r, nil
		}
	}
}

func modifier(req IRequestWrapper) requestWrapper {
	return requestWrapper{
		params:  req.Params(),
		headers: req.Headers(),
		body:    req.Body(),
		method:  req.Method(),
		url:     req.URL(),
		query:   req.Query(),
		path:    path.Join(req.Path()),
	}
}

var unknownTypeErr = errors.New("unknow request type")

type requestWrapper struct {
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }

type IRequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

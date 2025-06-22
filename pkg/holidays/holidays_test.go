package holidays

import (
	"errors"
	"github.com/jame-developer/aeontrac/configuration"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadHolidays(t *testing.T) {
	//goland:noinspection GoErrorStringFormat
	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		breakUrl       bool
		expectedError  error
		expectedLength int
	}{
		{
			name: "Success",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte(`[{"id":"1","startDate":"2022-01-01","endDate":"2022-01-01","type":"Public","name":[{"language":"en","text":"New Year's Day"}],"nationwide":true,"subdivisions":[{"code":"US","shortName":"United States"}]}]`))
			},
			expectedError:  nil,
			expectedLength: 1,
		},
		{
			name: "InvalidStartDate",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte(`[{"id":"1","startDate":"2022-13-01","endDate":"2022-01-01","type":"Public","name":[{"language":"en","text":"New Year's Day"}],"nationwide":true,"subdivisions":[{"code":"US","shortName":"United States"}]}]`))
			},
			expectedError: &time.ParseError{
				Layout:     time.DateOnly,
				Value:      "2022-13-01",
				LayoutElem: "01",
				ValueElem:  "-01",
				Message:    ": month out of range",
			},
			expectedLength: 0,
		},
		{
			name: "InvalidEndDate",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte(`[{"id":"1","startDate":"2022-01-01","endDate":"2022-13-01","type":"Public","name":[{"language":"en","text":"New Year's Day"}],"nationwide":true,"subdivisions":[{"code":"US","shortName":"United States"}]}]`))
			},
			expectedError: &time.ParseError{
				Layout:     time.DateOnly,
				Value:      "2022-13-01",
				LayoutElem: "01",
				ValueElem:  "-01",
				Message:    ": month out of range",
			},
			expectedLength: 0,
		},
		{
			name: "HttpError",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusInternalServerError)
			},
			expectedError:  errors.New("failed to load public holidays"),
			expectedLength: 0,
		},
		{
			name: "UnmarshalError",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte(`not valid json`))
			},
			expectedError:  errors.New("invalid character 'o' in literal null (expecting 'u')"),
			expectedLength: 0,
		},
		{
			name: "EmptyResponse",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte(`[]`))
			},
			expectedError:  nil,
			expectedLength: 0,
		},
		{
			name: "HttpGetError",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			},
			expectedError:  errors.New("Get \"http://localhost:1234?countryIsoCode=US&validFrom=2024-01-01&validTo=2024-12-31\": dial tcp 127.0.0.1:1234: connect: connection refused"),
			breakUrl:       true,
			expectedLength: 0,
		},
		{
			name: "ReadAllError",
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("Content-Length", "1")
			},
			expectedError:  errors.New("unexpected EOF"),
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handlerFunc)
			defer server.Close()
			testUrl := server.URL
			if tt.breakUrl {
				testUrl = "http://localhost:1234"
			}
			config := configuration.PublicHolidaysConfig{
				Country: "US",
				APIURL:  testUrl,
			}

			holidays, err := LoadHolidays(config, 2024)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedLength, len(holidays))
		})
	}
}

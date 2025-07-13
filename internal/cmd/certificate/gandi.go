package certificate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-gandi/go-gandi/types"
	"github.com/peterhellberg/link"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Gandi is the handle used to interact with the Gandi API
type Gandi struct {
	token    string
	endpoint string
}

type CertificateType struct {
	ID           string            `json:"id"`
	CN           string            `json:"cn"`
	Dates        *CertificateDates `json:"dates"`
	Status       string            `json:"status"`
	State        string            `json:"state"`
	StateDetail  string            `json:"state_detail"`
	ErrorMsg     string            `json:"error_msg"`
	Cert         string            `json:"cert"`
	CSR          string            `json:"csr"`
	Intermediate string            `json:"intermediate"`
	Tags         []string          `json:"tags"`
	Renewable    bool              `json:"renewable"`
}

type CertificateDates struct {
	CreatedAt         time.Time `json:"created_at,omitempty"`
	EndsAt            time.Time `json:"ends_at,omitempty"`
	StartedAt         time.Time `json:"started_at,omitempty"`
	SubcriptionEndsAt time.Time `json:"subscription_ends_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

type RenewCertificateRequest struct {
	CSR       string `json:"csr"`
	DCVMethod string `json:"dcv_method"`
}

type AttachTagsRequest struct {
	Tags []string `json:"tags"`
}

type DomainValidationDetailsRequest struct {
	CSR       string `json:"csr"`
	DCVMethod string `json:"dcv_method"`
}

type DomainValidationDetailsResponse struct {
	RawMessages [][]string `json:"raw_messages"`
}

type ErrorResponse struct {
	Cause   string `json:"cause,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Object  string `json:"object,omitempty"`
}

// GandiClient Instantiate a new Gandi client
func GandiClient() (*Gandi, error) {
	token := os.Getenv("GANDI_TOKEN")
	if token == "" {
		return nil, errors.New("no token found. please set the environment variable GANDI_TOKEN")
	}

	endpoint := "https://api.gandi.net/v5/certificate/"
	return &Gandi{token: token, endpoint: endpoint}, nil
}

// Get issues a GET request. It takes a subpath rooted in the endpoint. Response data is written to the recipient.
// Returns the response headers and any error
func (g *Gandi) Get(path string, params, recipient interface{}) (http.Header, error) {
	return g.askGandi(http.MethodGet, path, params, recipient)
}

// GetCollection supports pagination on GET requests. It takes a subpath rooted in the endpoint. Response data is written to the recipient.
// Returns the response headers and any error
func (g *Gandi) GetCollection(path string, params interface{}) (http.Header, []json.RawMessage, error) {
	return g.askGandiCollection(http.MethodGet, path, params)
}

// Post issues a POST request. It takes a subpath rooted in the endpoint. Response data is written to the recipient.
// Returns the response headers and any error
func (g *Gandi) Post(path string, params, recipient interface{}) (http.Header, error) {
	return g.askGandi(http.MethodPost, path, params, recipient)
}

// Put issues a PUT request. It takes a subpath rooted in the endpoint. Response data is written to the recipient.
// Returns the response headers and any error
func (g *Gandi) Put(path string, params, recipient interface{}) (http.Header, error) {
	return g.askGandi(http.MethodPut, path, params, recipient)
}

func (g *Gandi) askGandi(method, path string, params, recipient interface{}) (http.Header, error) {
	header, body, err := g.doAskGandi(method, path, params, nil)
	if err != nil {
		return nil, err
	}
	if recipient == nil {
		return header, nil
	}

	return header, json.Unmarshal(body, &recipient)
}

// askGandiCollection gets a resource collection even if it is
// paginated: it sends queries until all elements have been retrieved.
// Note this method only works if the API returns a list of objects.
func (g *Gandi) askGandiCollection(method, path string, params interface{}) (http.Header, []json.RawMessage, error) {
	var elements []json.RawMessage
	var header http.Header
	for {
		var partial []json.RawMessage
		header, err := g.askGandi(method, path, params, &partial)
		if err != nil {
			return nil, nil, err
		}
		elements = append(elements, partial...)

		if header.Get("link") == "" {
			break
		} else {
			var next string
			for _, l := range link.Parse(header.Get("link")) {
				if l.Rel == "next" {
					next = l.URI
					break
				}
			}
			if next == "" {
				return nil, nil, fmt.Errorf("the next page has not been found in the link header")
			}
			path = strings.TrimPrefix(next, g.endpoint)
		}
	}
	return header, elements, nil
}

// GetBytes issues a GET request but does not attempt to parse any response into JSON.
// It returns the response headers, a byteslice of the response, and any error
func (g *Gandi) GetBytes(path string, params interface{}) (http.Header, []byte, error) {
	headers := [][2]string{{"Accept", "text/plain"}}
	return g.doAskGandi(http.MethodGet, path, params, headers)
}

func allElementsInArray(array1 []string, array2 []string) bool {
	elementMap := make(map[string]bool)

	for _, element := range array1 {
		elementMap[element] = true
	}

	for _, element2 := range array2 {
		if !elementMap[element2] {
			return false
		}
	}

	return true
}

func contains(s []interface{}, str interface{}) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// doAskGandi performs a call to the API. If the HTTP status code of
// the response is not success, the returned error is a RequestError
// (which contains the HTTP StatusCode).
func (g *Gandi) doAskGandi(method, path string, p interface{}, extraHeaders [][2]string) (http.Header, []byte, error) {
	var (
		err error
		req *http.Request
	)
	params, err := json.Marshal(p)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to json.Marshal request params (error '%w')", err)
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if params != nil && string(params) != "null" {
		req, err = http.NewRequest(method, g.endpoint+path, bytes.NewReader(params))
	} else {
		req, err = http.NewRequest(method, g.endpoint+path, nil)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create the request (error '%w')", err)
	}
	req.Header.Add("Authorization", "Bearer "+g.token)
	req.Header.Add("Content-Type", "application/json")

	for _, header := range extraHeaders {
		req.Header.Add(header[0], header[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to do the request (error '%w')", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to read the body (error '%w')", err)
	}

	// Delete queries can return a 204 code. In this case, the
	// body is empty. See for instance:
	// https://api.gandi.net/docs/simplehosting/#delete-v5-simplehosting-instances-instance_id
	if resp.StatusCode == 204 {
		return resp.Header, []byte("{}"), err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var message types.StandardResponse
		var ctype = resp.Header["Content-Type"]
		if ctype[0] != "application/json" {
			return nil, nil, fmt.Errorf("response body is not json for status %d", resp.StatusCode)
		}

		if err = json.Unmarshal(body, &message); err != nil {
			return nil, nil, fmt.Errorf("fail to decode the response body (error '%w')", err)
		}
		if message.Message != "" {
			err = fmt.Errorf("%d: %s", resp.StatusCode, message.Message)
		} else if len(message.Errors) > 0 {
			var errs []string
			for _, oneError := range message.Errors {
				errs = append(errs, fmt.Sprintf("%s: %s", oneError.Name, oneError.Description))
			}
			err = fmt.Errorf("%s", strings.Join(errs, ", "))
		} else {
			err = fmt.Errorf("%d", resp.StatusCode)
		}
		err = &types.RequestError{
			Err:        err,
			StatusCode: resp.StatusCode,
		}
	}
	return resp.Header, body, err
}

// ListCertificates requests the list of issued certificates
func (g *Gandi) ListCertificates() ([]CertificateType, error) {
	var certificates []CertificateType
	_, elements, err := g.GetCollection("issued-certs", nil)
	if err != nil {
		return nil, err
	}
	for _, element := range elements {
		var certificate CertificateType
		err := json.Unmarshal(element, &certificate)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certificate)
	}
	return certificates, nil
}

// GetCertificateBy requests the only issued certificates matching specific criteria
func (g *Gandi) GetCertificateBy(criteria map[string]interface{}) (*CertificateType, error) {
	log.Println("Get certificate by criteria: {" + mapToString(criteria) + "}")

	certificates, err := g.ListCertificates()
	if err != nil {
		return nil, err
	}

	var certs []CertificateType

	for i := 0; i < len(certificates); i++ {
		certValues := reflect.ValueOf(certificates[i])
		match := true

		for key, value := range criteria {
			field := certValues.FieldByName(key)

			if field.IsValid() {
				fieldType := field.Type()
				fieldValue := field.Interface()
				if fieldType.Kind() == reflect.Slice {
					if reflect.TypeOf(value).Kind() == reflect.Slice {
						match = match && allElementsInArray(fieldValue.([]string), value.([]string))
					} else {
						match = match && contains(fieldValue.([]interface{}), value.(string))
					}
				} else {
					match = match && fieldValue == value
				}
			} else {
				match = false
			}
		}

		if match {
			certs = append(certs, certificates[i])
		}
	}
	if len(certs) == 0 {
		return nil, errors.New("no certificate correspond to the criteria: {" + mapToString(criteria) + "}")
	}

	if len(certs) > 1 {
		return nil, errors.New("Too many certificate correspond to the criteria: {" + mapToString(criteria) + "}")
	}

	return &certs[0], nil
}

func mapToString(mapData map[string]interface{}) string {
	var result []string

	for key, value := range mapData {
		result = append(result, fmt.Sprintf("%s => %v", key, value))
	}

	return strings.Join(result, ", ")
}

func (g *Gandi) Renew(certificateID string, request RenewCertificateRequest) (ErrorResponse, error) {
	var response ErrorResponse
	_, err := g.Post("issued-certs/"+certificateID, request, &response)

	return response, err
}

// AskDomainValidation Ask to Gandi to validate the domain by checking dns records (DCV)
func (g *Gandi) AskDomainValidation(certificateID string) (ErrorResponse, error) {
	var response ErrorResponse
	_, err := g.Put("issued-certs/"+certificateID+"/dcv", nil, &response)
	return response, err
}

func (g *Gandi) GetDomainValidationDetails(certificateID string, request DomainValidationDetailsRequest) (DomainValidationDetailsResponse, error) {
	var response DomainValidationDetailsResponse
	_, err := g.Post("issued-certs/"+certificateID+"/dcv_params", request, &response)
	return response, err
}

// AttachTags attach tags to the certificate
func (g *Gandi) AttachTags(certificateID string, request AttachTagsRequest) (ErrorResponse, error) {
	var response ErrorResponse
	_, err := g.Put("issued-certs/"+certificateID+"/tags", request, &response)
	return response, err
}

// GetCertificate request details of an issued certificates
func (g *Gandi) GetCertificate(certificateID string) (CertificateType, error) {
	var certificate CertificateType
	_, err := g.Get("issued-certs/"+certificateID, nil, &certificate)
	return certificate, err
}

// GetIntermediate get intermediate certificate
func (g *Gandi) GetIntermediate(certificate *CertificateType) (string, error) {
	segments := strings.Split(certificate.Intermediate, "/")
	filename := segments[len(segments)-1]
	_, intermediate, err := g.GetBytes("pem/-/"+filename, nil)

	return string(intermediate), err
}

func certificateGandiCheck(cmd *cobra.Command, args []string) error {
	client, err := GandiClient()
	if err != nil {
		return err
	}
	var certs []*CertificateType
	_, err = client.Get("issued-certs", nil, &certs)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		fmt.Println(cert.CN, cert.Status)
	}
	return nil
}

var certificateGandiCheckCmd = &cobra.Command{
	Use:   "gandi-check",
	Short: "Check your Gandi access",
	RunE:  certificateGandiCheck,
	Args:  cobra.ExactArgs(0),
}

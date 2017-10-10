package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func GetClientErrorFromString(str string) error {
	return &Error{
		ErrorResponse: ErrorResponse{
			Code:    "AliyunGoClientFailure",
			Message: str,
		},
		StatusCode: -1,
	}
}

func GetClientError(err error) error {
	return GetClientErrorFromString(err.Error())
}

type Region string

type Client struct {
	AccessKeyId     string //Access Key Id
	AccessKeySecret string //Access Key Secret
	debug           bool
	httpClient      *http.Client
	endpoint        string
	version         string
	serviceCode     string
	regionID        Region
	businessInfo    string
	userAgent       string
}

//Init 初始化Client
func (client *Client) Init(endpoint, version, accessKeyId, accessKeySecret string) {
	client.AccessKeyId = accessKeyId
	client.AccessKeySecret = accessKeySecret + "&"
	client.debug = false
	client.httpClient = &http.Client{}
	client.endpoint = endpoint
	client.version = version
}

// SetEndpoint sets custom endpoint
func (client *Client) SetEndpoint(endpoint string) {
	client.endpoint = endpoint
}

// SetEndpoint sets custom version
func (client *Client) SetVersion(version string) {
	client.version = version
}

func (client *Client) SetRegionID(regionID Region) {
	client.regionID = regionID
}

//SetServiceCode sets serviceCode
func (client *Client) SetServiceCode(serviceCode string) {
	client.serviceCode = serviceCode
}

// SetAccessKeyId sets new AccessKeyId
func (client *Client) SetAccessKeyId(id string) {
	client.AccessKeyId = id
}

// SetAccessKeySecret sets new AccessKeySecret
func (client *Client) SetAccessKeySecret(secret string) {
	client.AccessKeySecret = secret + "&"
}

// SetDebug sets debug mode to log the request/response message
func (client *Client) SetDebug(debug bool) {
	client.debug = debug
}

func (client *Client) InvokeByAnyMethod(method, action, path string, args interface{}, response interface{}) error {

	request := Request{}
	request.init(client.version, action, client.AccessKeyId)

	data := ConvertToQueryValues(request)
	SetQueryValues(args, &data)

	// Sign request
	signature := CreateSignatureForRequest(method, &data, client.AccessKeySecret)

	data.Add("Signature", signature)
	// Generate the request URL
	var (
		httpReq *http.Request
		err     error
	)
	if method == http.MethodGet {
		requestURL := client.endpoint + path + "?" + data.Encode()
		//fmt.Println(requestURL)
		httpReq, err = http.NewRequest(method, requestURL, nil)
	} else {
		//fmt.Println(client.endpoint + path)
		httpReq, err = http.NewRequest(method, client.endpoint+path, strings.NewReader(data.Encode()))
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return GetClientError(err)
	}

	httpReq.Header.Set("User-Agent", httpReq.Header.Get("User-Agent")+" "+client.userAgent)

	t0 := time.Now()
	httpResp, err := client.httpClient.Do(httpReq)
	t1 := time.Now()
	if err != nil {
		return GetClientError(err)
	}
	statusCode := httpResp.StatusCode

	if client.debug {
		log.Printf("Invoke %s %s %d (%v) %v", ECSRequestMethod, client.endpoint, statusCode, t1.Sub(t0), data.Encode())
	}

	defer httpResp.Body.Close()
	body, err := ioutil.ReadAll(httpResp.Body)

	if err != nil {
		return GetClientError(err)
	}

	if client.debug {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "    ")
		log.Println(string(prettyJSON.Bytes()))
	}

	if statusCode >= 400 && statusCode <= 599 {
		errorResponse := ErrorResponse{}
		err = json.Unmarshal(body, &errorResponse)
		ecsError := &Error{
			ErrorResponse: errorResponse,
			StatusCode:    statusCode,
		}
		return ecsError
	}

	err = json.Unmarshal(body, response)
	//log.Printf("%++v", response)
	if err != nil {
		return GetClientError(err)
	}

	return nil
}

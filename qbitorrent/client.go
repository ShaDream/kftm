package qbitorrent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
)

const baseApi = "/api/v2"

var ErrBadResponse = errors.New("bad response")

type Client struct {
	BaseUrl string
	client  *http.Client
}

func NewClient(BaseUrl string) *Client {
	client := http.Client{}
	return &Client{
		BaseUrl: BaseUrl,
		client:  &client,
	}
}

func (c *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.BaseUrl, baseApi, endpoint)
	if err != nil {
		return nil, fmt.Errorf("join endpoint path: %w", err)
	}

	// add optional parameters that the user wants
	form := url.Values{}
	if opts != nil {
		for k, v := range opts {
			form.Add(k, v)
		}
	}

	req, err := http.NewRequest(http.MethodPost, endpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create post request: %w", err)
	}

	host := fmt.Sprintf("%s://%s", req.URL.Scheme, req.Host)

	req.Header.Set("Referer", host)
	req.Header.Set("Origin", host)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do post request: %w", err)
	}

	return resp, nil
}

func (c *Client) postMultipart(endpoint string, opts map[string]string) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.BaseUrl, baseApi, endpoint)
	if err != nil {
		return nil, fmt.Errorf("join endpoint path: %w", err)
	}

	body := new(bytes.Buffer)
	args := multipart.NewWriter(body)

	for key, value := range opts {
		args.WriteField(key, value)
	}

	args.Close()

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create post request: %w", err)
	}

	host := fmt.Sprintf("%s://%s", req.URL.Scheme, req.Host)

	req.Header.Set("Referer", host)
	req.Header.Set("Origin", host)

	req.Header.Set("Content-Type", args.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do post request: %w", err)
	}

	return resp, nil
}

func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.BaseUrl, baseApi, endpoint)
	if err != nil {
		return nil, fmt.Errorf("join endpoint path: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create get request: %w", err)
	}

	host := fmt.Sprintf("%s://%s", req.URL.Scheme, req.Host)

	req.Header.Set("Referer", host)
	req.Header.Set("Origin", host)

	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "kftm v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		query := req.URL.Query()
		for k, v := range opts {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do get request: %w", err)
	}

	return resp, nil
}

func (c *Client) Login(username, password string) error {
	credentials := make(map[string]string)
	credentials["username"] = username
	credentials["password"] = password

	resp, err := c.post("auth/login", credentials)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login response: %w", ErrBadResponse)
	}

	cookies := resp.Cookies()

	if len(cookies) > 0 {
		if c.client.Jar == nil {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			c.client.Jar = jar
		}
		c.client.Jar.SetCookies(resp.Request.URL, cookies)
	}

	return nil
}

func (c *Client) GetTorrentList(filters TorrentsInfoPostFormdataBody) ([]TorrentInfo, error) {
	var t []TorrentInfo

	args, err := StructToMap(filters)
	if err != nil {
		return nil, fmt.Errorf("convert struct to map: %w", err)
	}

	resp, err := c.post("torrents/info", args)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, wrapWrongStatusCode(resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return nil, fmt.Errorf("decode request body: %w", err)
	}

	return t, nil
}

func (c *Client) GetTorrentContent(hash string) ([]TorrentsFiles, error) {
	var t []TorrentsFiles

	args := map[string]string{
		"hash": hash,
	}

	resp, err := c.post("/torrents/files", args)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, wrapWrongStatusCode(resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return nil, fmt.Errorf("decode request body: %w", err)
	}

	return t, nil
}

func (c *Client) CreateTorrentFileUrl(addTorrentInfo AddTorrentsURLs) error {
	args, err := StructToMap(addTorrentInfo)
	if err != nil {
		return fmt.Errorf("convert struct to map: %w", err)
	}

	resp, err := c.postMultipart("/torrents/add", args)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return wrapWrongStatusCode(resp.StatusCode)
	}

	return nil
}

func (c *Client) RenameFile(data RenameTorrentFiles) error {
	args, err := StructToMap(data)
	if err != nil {
		return fmt.Errorf("convert struct to map: %w", err)
	}

	resp, err := c.post("/torrents/renameFile", args)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return wrapWrongStatusCode(resp.StatusCode)
	}

	return nil
}

func (c *Client) RenameFolder(data RenameTorrentFiles) error {
	args, err := StructToMap(data)
	if err != nil {
		return fmt.Errorf("convert struct to map: %w", err)
	}

	resp, err := c.post("/torrents/renameFolder", args)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return wrapWrongStatusCode(resp.StatusCode)
	}

	return nil
}

func (c *Client) GetAllCategories() (map[string]Category, error) {
	var result map[string]Category

	resp, err := c.get("/torrents/categories", nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, wrapWrongStatusCode(resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decode request body: %w", err)
	}

	return result, nil
}

func (c *Client) CreateCategory(category, savePath string) error {
	args := map[string]string{
		"category": category,
		"savePath": savePath,
	}

	resp, err := c.post("/torrents/renameFolder", args)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return wrapWrongStatusCode(resp.StatusCode)
	}

	return nil
}

func StructToMap(obj interface{}) (map[string]string, error) {
	result := make(map[string]string)
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get the JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		// Extract the JSON field name and check for "omitempty"
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		omitEmpty := strings.Contains(jsonTag, "omitempty")

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				if omitEmpty {
					continue // Skip nil pointers if omitempty is set
				}
				result[jsonFieldName] = "" // Add empty string for nil pointers
				continue
			}
			fieldValue = fieldValue.Elem() // Dereference the pointer
		}

		// Check if the field is a zero value and has "omitempty"
		if omitEmpty && isZero(fieldValue) {
			continue
		}

		// Convert the field value to a string
		var strValue string
		switch fieldValue.Kind() {
		case reflect.String:
			strValue = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strValue = fmt.Sprintf("%d", fieldValue.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strValue = fmt.Sprintf("%d", fieldValue.Uint())
		case reflect.Float32, reflect.Float64:
			strValue = fmt.Sprintf("%f", fieldValue.Float())
		case reflect.Bool:
			strValue = fmt.Sprintf("%t", fieldValue.Bool())
		case reflect.Pointer:
		default:
			// For other types, use JSON marshaling
			jsonBytes, err := json.Marshal(fieldValue.Interface())
			if err != nil {
				return nil, fmt.Errorf("failed to marshal field %s: %v", field.Name, err)
			}
			strValue = string(jsonBytes)
		}

		result[jsonFieldName] = strValue
	}

	return result, nil
}

func wrapWrongStatusCode(statusCode int) error {
	return fmt.Errorf("wrong status code %d: %w", statusCode, ErrBadResponse)
}

// isZero checks if a reflect.Value is the zero value for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

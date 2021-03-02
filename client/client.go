package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"memory-cache/msgtypes"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(serverAddr string) *Client {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	return &Client{
		url: "http://" + serverAddr,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		},
	}
}

func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	setTtl := msgtypes.Duration(ttl)
	setReq := msgtypes.SetReq{
		Key:   key,
		Value: value,
		Ttl:   setTtl,
	}

	jsonData, err := json.Marshal(setReq)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.url+"/set", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body error: %v", err)
	}

	if err := c.checkResponseStatus(resp, body); err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(key string) (interface{}, error) {
	url := fmt.Sprintf("%v/get/%v", c.url, key)
	return c.valueResponse(url)
}

func (c *Client) GetListElem(key string, index int) (interface{}, error) {
	url := fmt.Sprintf("%v/getListElem/%v/%v", c.url, key, index)
	return c.valueResponse(url)
}

func (c *Client) GetMapElemValue(key string, mapKey string) (interface{}, error) {
	url := fmt.Sprintf("%v/getMapElemValue/%v/%v", c.url, key, mapKey)
	return c.valueResponse(url)
}

func (c *Client) Keys() ([]string, error) {
	url := fmt.Sprintf("%v/keys", c.url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %v", err)
	}

	if err := c.checkResponseStatus(resp, body); err != nil {
		return nil, err
	}

	keysResp := &msgtypes.KeysResp{}
	err = json.Unmarshal(body, keysResp)
	if err != nil {
		return nil, err
	}

	return keysResp.Keys, nil
}

func (c *Client) Remove(key string) error {
	url := fmt.Sprintf("%v/remove/%v", c.url, key)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body error: %v", err)
	}

	if err := c.checkResponseStatus(resp, body); err != nil {
		return err
	}

	return nil
}

func (c *Client) valueResponse(url string) (interface{}, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %v", err)
	}

	if err := c.checkResponseStatus(resp, body); err != nil {
		return nil, err
	}

	valueResp := &msgtypes.ValueResp{}
	err = json.Unmarshal(body, valueResp)
	if err != nil {
		return nil, err
	}

	return valueResp.Value, nil
}

func (c *Client) checkResponseStatus(resp *http.Response, body []byte) error {
	if resp.StatusCode != http.StatusOK {
		errorResp := &msgtypes.ErrorResp{}
		err := json.Unmarshal(body, errorResp)
		if err != nil {
			return err
		}

		return fmt.Errorf("error responce '%v', status code: '%v'", errorResp.Error, resp.StatusCode)
	}

	return nil
}

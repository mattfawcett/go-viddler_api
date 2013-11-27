package viddler_api

import (
  "net/http"
  "net/url"
  "log"
  "errors"
  "io/ioutil"
  "github.com/bitly/go-simplejson"
)

type Client struct {
  ApiKey string
  SessionId string
  BaseDomain string
}

func NewClient(apiKey string) *Client {
  c := &Client{ ApiKey: apiKey, BaseDomain: "http://api.viddler.com" }
  return c
}

func (c *Client) mapToUrlValues(paramsMap map[string]string) (values url.Values) {
  values = url.Values{}
  values.Add("key", c.ApiKey)
  if c.SessionId != "" {
    values.Add("sessionid", c.SessionId)
  }
  for k, v := range paramsMap{
    values.Add(k, v)
  }
  return values
}

func (c *Client) generateBaseUrl(method string) (baseUrl string) {
  baseUrl  = c.BaseDomain + "/api/v2/"
  baseUrl += method
  baseUrl += ".json?"
  return baseUrl
}

func (c *Client) handleResponse(res http.Response, responseError error) (json *simplejson.Json, err error) {
  if err != nil {
    log.Fatal(err)
  }

  body, err := ioutil.ReadAll(res.Body)
  res.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  json, err = simplejson.NewJson(body)
  if err != nil {
    log.Fatal("error parsing json")
  }

  errorObj, containsErrors := json.CheckGet("error")
  if containsErrors {
    errorString, _ := errorObj.Get("description").String()
    err = errors.New(errorString)
  }

  return json, err
}

func (c *Client) Get(method string, paramsMap map[string]string) (json *simplejson.Json, err error) {
  params := c.mapToUrlValues(paramsMap)

  fullUrl := c.generateBaseUrl(method)
  fullUrl += params.Encode()

  res, err := http.Get(fullUrl)
  json, err = c.handleResponse(*res, err)
  return json, err
}


func (c *Client) Post(method string, paramsMap map[string]string) (json *simplejson.Json, err error) {
  urlValues := c.mapToUrlValues(paramsMap)
  fullUrl   := c.generateBaseUrl(method)

  res, err := http.PostForm(fullUrl, urlValues)
  json, err = c.handleResponse(*res, err)
  return json, err
}

func (c *Client) Authenticate(username string, password string) (success bool) {
  c.SessionId = ""

  params := map[string]string{
    "username": username,
    "password": password,
  }

  response, err := c.Get("viddler.users.auth", params)
  if err == nil {
    c.SessionId, _ = response.Get("auth").Get("sessionid").String()
  }

  return c.SessionId != ""
}

// Copyright 2011-2013 Numrotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
//
// Developed at www.stathat.com by Patrick Crosby
// Contact us on twitter with any questions:  twitter.com/stat_hat

// amzses is a Go package to send emails using Amazon's Simple Email Service.
package amzses // import "stathat.com/c/amzses"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"stathat.com/c/jconfig"
)

const (
	endpoint = "https://email.us-east-1.amazonaws.com"
)

type Client struct {
	accessKey string
	secretKey string
}

func NewClient(sesAccessKey, sesSecretKey string) *Client {
	client := &Client{accessKey: sesAccessKey, secretKey: sesSecretKey}
	return client
}

func (c *Client) SendMail(from, to, subject, body string) (string, error) {
	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", from)
	data.Add("Destination.ToAddresses.member.1", to)
	data.Add("Message.Subject.Data", subject)
	data.Add("Message.Body.Text.Data", body)
	data.Add("AWSAccessKeyId", c.accessKey)

	return c.sesPost(data)
}

func (c *Client) SendMailHTML(from, to, subject, bodyText, bodyHTML string) (string, error) {
	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", from)
	data.Add("Destination.ToAddresses.member.1", to)
	data.Add("Message.Subject.Data", subject)
	data.Add("Message.Body.Text.Data", bodyText)
	data.Add("Message.Body.Html.Data", bodyHTML)
	data.Add("AWSAccessKeyId", c.accessKey)

	return c.sesPost(data)
}

func (c *Client) authorizationHeader(date string) []string {
	h := hmac.New(sha256.New, []uint8(c.secretKey))
	h.Write([]uint8(date))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	auth := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", c.accessKey, signature)
	return []string{auth}
}

func (c *Client) sesGet(data url.Values) (string, error) {
	urlstr := fmt.Sprintf("%s?%s", endpoint, data.Encode())
	endpointURL, _ := url.Parse(urlstr)
	headers := map[string][]string{}

	now := time.Now().UTC()
	// date format: "Tue, 25 May 2010 21:20:27 +0000"
	date := now.Format("Mon, 02 Jan 2006 15:04:05 -0700")
	headers["Date"] = []string{date}

	h := hmac.New(sha256.New, []uint8(c.secretKey))
	h.Write([]uint8(date))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	auth := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", c.accessKey, signature)
	headers["X-Amzn-Authorization"] = []string{auth}

	req := http.Request{
		URL:        endpointURL,
		Method:     "GET",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
		Header:     headers,
	}

	r, err := http.DefaultClient.Do(&req)
	if err != nil {
		log.Printf("http error: %s", err)
		return "", err
	}

	resultbody, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if r.StatusCode != 200 {
		log.Printf("error, status = %d", r.StatusCode)

		log.Printf("error response: %s", resultbody)
		return "", errors.New(string(resultbody))
	}

	return string(resultbody), nil
}

func (c *Client) sesPost(data url.Values) (string, error) {
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	now := time.Now().UTC()
	// date format: "Tue, 25 May 2010 21:20:27 +0000"
	date := now.Format("Mon, 02 Jan 2006 15:04:05 -0700")
	req.Header.Set("Date", date)

	h := hmac.New(sha256.New, []uint8(c.secretKey))
	h.Write([]uint8(date))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	auth := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", c.accessKey, signature)
	req.Header.Set("X-Amzn-Authorization", auth)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("http error: %s", err)
		return "", err
	}

	resultbody, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if r.StatusCode != 200 {
		log.Printf("error, status = %d", r.StatusCode)

		log.Printf("error response: %s", resultbody)
		return "", errors.New(fmt.Sprintf("error code %d. response: %s", r.StatusCode, resultbody))
	}

	return string(resultbody), nil
}

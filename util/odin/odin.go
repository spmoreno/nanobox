// Package odin ...
package odin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/nanobox/models"
	"github.com/nanobox-io/nanobox/util/config"
	"github.com/nanobox-io/nanobox/util/data"
)

// Auth ...
func Auth(username, password string) (string, error) {

	//
	params := url.Values{}
	params.Set("password", password)

	//
	resBody := map[string]string{}

	//
	if err := doRequest("GET", fmt.Sprintf("users/%s/auth_token", username), params, nil, &resBody); err != nil {
		return "", err
	}

	return resBody["authentication_token"], nil
}

// App ...
func App(slug string) (models.App, error) {
	app := models.App{}

	return app, doRequest("GET", "apps/"+slug, nil, nil, &app)
}

// Deploy ...
func Deploy(appID, id, boxfile, message string) error {

	//
	body := map[string]map[string]string{
		"deploy": map[string]string{
			"boxfile_content": boxfile,
			"build_id":        id,
			"commit_message":  message,
		},
	}

	return doRequest("POST", fmt.Sprintf("apps/%s/deploys", appID), nil, body, nil)
}

// EstablishTunnel ...
func EstablishTunnel(appID, id string) (string, string, string, error) {
	// TODO: do somethign else here
	r := map[string]string{}
	err := doRequest("GET", fmt.Sprintf("apps/%s/tunnels/%s", appID, id), nil, nil, &r)

	return r["token"], r["url"], r["container"], err
}

// EstablishConsole ...
func EstablishConsole(appID, id string) (string, string, string, error) {
	// TODO: do somethign else here
	r := map[string]string{}
	err := doRequest("GET", fmt.Sprintf("apps/%s/consoles/%s", appID, id), nil, nil, &r)

	return r["token"], r["url"], r["container"], err
}

// GetWarehouse ...
func GetWarehouse(appID string) (string, string, error) {
	// TODO: do somethign else here
	r := map[string]string{}
	err := doRequest("GET", fmt.Sprintf("apps/%s/services/warehouse", appID), nil, nil, &r)

	return r["token"], r["url"], err
}

func GetPreviousBuild(appID string) (string, error) {
	r := []map[string]string{}
	err := doRequest("GET", fmt.Sprintf("apps/%s/deploys", appID), nil, nil, &r)
	if err != nil {
		return "", err
	}

	if len(r) > 0 {
		return r[0]["build_id"], nil
	}

	return "", nil
}

// doRequest ...
func doRequest(method, path string, params url.Values, requestBody, responseBody interface{}) error {

	var rbodyReader io.Reader

	//
	if requestBody != nil {
		jsonBytes, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		rbodyReader = bytes.NewBuffer(jsonBytes)
	}

	productionUrl := config.Viper().GetString("production_url")
	auth := models.Auth{}
	data.Get("global", "user", &auth)

	if params == nil {
		params = url.Values{}
	}
	params.Set("auth_token", auth.Key)

	//
	lumber.Debug("%s%s?%s\n", productionUrl, path, params.Encode())
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s?%s", productionUrl, path, params.Encode()), rbodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	lumber.Debug("req: %+v\n", req)

	//
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	lumber.Debug("res: %+v\n", res)

	//
	if responseBody != nil {
		b, err := ioutil.ReadAll(res.Body)
		lumber.Debug("response body: '%s'\n", b)
		err = json.Unmarshal(b, responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}

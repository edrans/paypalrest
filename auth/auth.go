package auth

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "time"
    "fmt"
)

type AuthRequest struct {
    ClientId     string
    ClientSecret string
}

type PayPalAuth struct {
    Endpoint     string
    ClientId     string
    ClientSecret string
    Scope        string
    Access_token string
    Token_type   string
    App_id       string
    Expires_in   int32
    Expires_on   time.Time
}

func NewAuth(endpoint string, clientId string, clientSecret string) (response *PayPalAuth, err error) {
    var ok bool
    var req *http.Request
    var resp *http.Response
    endpoints := map[string]string{"sandbox": "https://api.sandbox.paypal.com", "live": "https://api.paypal.com"}
    response = &PayPalAuth{}
    response.ClientId = clientId
    response.ClientSecret = clientSecret
    response.Endpoint, ok = endpoints[endpoint]
    if ok == false {
        err = errors.New("Invalid connection type. It must be sandbox or live")
        return
    }

    client := &http.Client{}
    values := url.Values{"grant_type": {"client_credentials"}}

    req, err = http.NewRequest("POST", response.Endpoint+"/v1/oauth2/token", strings.NewReader(values.Encode()))
    if err == nil {
        req.Header.Add("Accept", "application/json")
        req.Header.Add("Accept-Language", "en_US")
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        req.SetBasicAuth(clientId, clientSecret)
        resp, err = client.Do(req)
        if err == nil {
            defer resp.Body.Close()
            body, e := ioutil.ReadAll(resp.Body)
            if e == nil {
                e = json.Unmarshal(body, &response)
                if e == nil {
                    response.Expires_on = time.Now().UTC().Add(time.Duration(response.Expires_in) * time.Second).UTC()
                } else {
                    err = errors.New(fmt.Sprintf("Cannot parse PayPal Auth response body: %v", err))
                }
            } else {
                err = errors.New(fmt.Sprintf("Invalid PayPal API response: %v", err))
            }
        }
    }
    return
}

func (this *PayPalAuth) GetToken() (token string, err error) {
    var auth *PayPalAuth
    if this.Expires_on.Before(time.Now().UTC()) {
        auth, err = NewAuth(this.Endpoint, this.ClientId, this.ClientSecret)
        if err == nil {
            this.App_id = auth.App_id
            this.Token_type = auth.Token_type
            this.Access_token = auth.Access_token
            this.Expires_in = auth.Expires_in
            this.Expires_on = auth.Expires_on
        }
    }
    return this.Access_token, err
}

func (this *PayPalAuth) SetEndpoint(endpoint string) {
    this.Endpoint = endpoint
}

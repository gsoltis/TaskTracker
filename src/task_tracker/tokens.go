package task_tracker

import (
	"encoding/json"
	"crypto/rsa"
	"math/big"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func make_pk(n_str string, e_str string) (*rsa.PublicKey, error) {
	n_bytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(n_str)
	if err != nil {
		return nil, err
	}
	n := new(big.Int)
	n.SetBytes(n_bytes)
	e_bytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(e_str)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 8)
	diff := 8 - len(e_bytes)
	for i := 0; i < diff; i++ {
		buf[i] = 0
	}
	for i := 0; i < len(e_bytes); i++ {
		buf[diff + i] = e_bytes[i]
	}
	e := binary.BigEndian.Uint64(buf)
	key := rsa.PublicKey{
		N: n,
		E: int(e),
	}
	return &key, nil
}

func parse_keys(untyped interface{}) (map[string]*rsa.PublicKey, error) {
	keys := untyped.([]interface{})
	key_map := make(map[string]*rsa.PublicKey)
	for _, key_interface := range keys {
		key_data := key_interface.(map[string]interface{})
		key, err := make_pk(key_data["n"].(string), key_data["e"].(string))
		if err != nil {
			return nil, err
		}
		key_map[key_data["kid"].(string)] = key
	}
	return key_map, nil
}

func KeysFromString(in []byte) (map[string]*rsa.PublicKey, error) {
	var dat map[string]interface{};
	json.Unmarshal(in, &dat)
	return parse_keys(dat["keys"])
}

const keysURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"

type KeyCache struct {
	Keys map[string]*rsa.PublicKey
	Refreshed time.Time
}

func NewKeyCache(ctx appengine.Context) (*KeyCache, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(keysURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	key_map, err := KeysFromString(body)
	if err != nil {
		return nil, err
	}
	kc := KeyCache{Keys: key_map, Refreshed: time.Now()}
	return &kc, nil
}

func InitKeyCache(ctx appengine.Context) error {
	kc, err := NewKeyCache(ctx)
	if err != nil {
		return err
	}
	keyCache = kc
	return nil
}

var keyCache *KeyCache = nil

func (kc *KeyCache) GetKey(kid string) *rsa.PublicKey {
	return kc.Keys[kid]
}

func (kc *KeyCache) NeedsRefresh() bool {
	return kc == nil || kc.Refreshed.Before(time.Now().Sub(time.Hour))
}

func authRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		defer req.Body.Close()
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if keyCache.NeedsRefresh() {
			err = InitKeyCache(appengine.NewContext(req))
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		}
		token, err := jwt.Parse(string(body), func(token *jwt.Token) (interface{}, error) {
			kid := token.Header["kid"].(string)
			return keyCache.GetKey(kid), nil
		})
		_, err = NewSession(token, w, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}


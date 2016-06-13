package task_tracker

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"net/http"
	"appengine"
	"encoding/gob"
)

type TaskTrackerUser struct {
	Email string
}

func init() {
	gob.Register(&TaskTrackerUser{})
}

const userKey = "u"
var cookieStore = sessions.NewCookieStore([]byte("fcaea089c13791a33efa429979e26c2030fdf1d2df9c08389225361695d5623c"))
const sessionName = "_s"

func UserForToken(ctx appengine.Context, token *jwt.Token) (*TaskTrackerUser, error) {
	ctx.Debugf("Stuff: %v", token.Claims)
	return &TaskTrackerUser{
		Email: token.Claims["email"].(string),
	}, nil
}

func NewSession(token *jwt.Token, w http.ResponseWriter, req *http.Request) (*sessions.Session, error) {
	ctx := appengine.NewContext(req)
	user, err := UserForToken(ctx, token)
	if err != nil {
		return nil, err
	}
	session, err := cookieStore.New(req, sessionName)
	if err != nil {
		return nil, err
	}
	session.Values[userKey] = user
	err = session.Save(req, w)
	if err != nil {
		return nil, err
	} else {
		return session, nil
	}
}

func UserForRequest(req *http.Request) (*TaskTrackerUser, error) {
	session, err := cookieStore.Get(req, sessionName)
	if err != nil {
		return nil, err
	}
	val := session.Values[userKey]
	return val.(*TaskTrackerUser), nil
}

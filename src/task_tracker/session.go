package task_tracker

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"net/http"
	"appengine"
	"encoding/gob"
	"appengine/datastore"
	"time"
)

type TaskTrackerUser struct {
	Email string
	UserId string
	LastAggregation time.Time
}

func (u *TaskTrackerUser) Key(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "User", u.UserId, 0, nil)
}

func init() {
	gob.Register(&TaskTrackerUser{})
}

const userKey = "u"
var cookieStore = sessions.NewCookieStore([]byte("fcaea089c13791a33efa429979e26c2030fdf1d2df9c08389225361695d5623c"))
const sessionName = "_s"

func UserForToken(ctx appengine.Context, token *jwt.Token) (*TaskTrackerUser, error) {
	ctx.Debugf("Stuff: %v", token.Claims)
	claims := token.Claims.(jwt.MapClaims)
	user_id := claims["user_id"].(string)
	user_key := datastore.NewKey(ctx, "User", user_id, 0, nil)
	var user = &TaskTrackerUser{}
	err := datastore.Get(ctx, user_key, user)
	if err == datastore.ErrNoSuchEntity {
		user = &TaskTrackerUser{
			Email: claims["email"].(string),
			UserId: user_id,
			LastAggregation: time.Now(),
		}
		_, err := datastore.Put(ctx, user_key, user)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		token_email := claims["email"].(string)
		if user.Email != token_email {
			ctx.Debugf("User email case")
			user.Email = token_email
			_, err = datastore.Put(ctx, user_key, user)
			if err != nil {
				return nil, err
			}
		}
	}
	return user, nil
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
	session.Values[userKey] = *user
	err = session.Save(req, w)
	if err != nil {
		return nil, err
	} else {
		return session, nil
	}
}

func UserForRequest(req *http.Request) *TaskTrackerUser {
	session, err := cookieStore.Get(req, sessionName)
	if err != nil {
		return nil
	}
	val := session.Values[userKey]
	user, ok := val.(*TaskTrackerUser)
	if !ok {
		return nil
	}
	return user
}

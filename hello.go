package HelloAppEngine

import (
	"net/http"

	"appengine"
	"appengine/user"

	"html/template"
	"time"
	"appengine/datastore"
)

type Greeting struct {
	Author string
	Content string
	Date time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/sign", sign)
}


func guestBookParentKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func internalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

//var signTemplate = template.Must(template.New("sign").Parse(signTemplateHtml))
func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if u := user.Current(c); u != nil {
		g := Greeting{
			Content: r.FormValue("content"),
			Author: u.String(),
			Date: time.Now(),
		}
		key := datastore.NewIncompleteKey(c, "Greeting", guestBookParentKey(c))
		_, err := datastore.Put(c, key, &g)
		if err != nil {
			internalServerError(w, err)
			return
		} else {
			c.Infof("stuff")
		}
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		redirectToLogin(w, r, c)
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, c appengine.Context) {
	url, err := user.LoginURL(c, r.URL.String())
	if err != nil {
		internalServerError(w, err)
		return
	}
	//w.Header().Set("Location", url)
	//w.WriteHeader(http.StatusFound)
	http.Redirect(w, r, url, http.StatusFound)
}


const guestBookHtml = `
<html>
<head>
  <title>Go Guestbook</title>
</head>
<body>
  {{range .}}
    <p><b>{{.Author}}</b> wrote</p>
    <pre>{{.Content}}</pre>
  {{end}}
  <form action="/sign" method="post">
    <div><textarea name="content" rows="3" cols="60"></textarea></div>
    <div><input type="submit" value="Sign Guestbook" /></div>
  </form>
</body>
</html>
`

var guestBookTemplate = template.Must(template.New("Book").Parse(guestBookHtml))

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)

	if u == nil {
		redirectToLogin(w, r, c)
		return
	} else {
		query := datastore.NewQuery("Greeting").Ancestor(guestBookParentKey(c)).Order("-Date").Limit(10)
		greetings := make([]Greeting, 0, 10)
		if _, err := query.GetAll(c, &greetings); err != nil {
			internalServerError(w, err)
			return
		} else {
			c.Infof("Found %v", greetings)
			if err := guestBookTemplate.Execute(w, greetings); err != nil {
				internalServerError(w, err)
			}
		}
	}
}

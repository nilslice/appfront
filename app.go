package appfront

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	client "github.com/ponzu-cms/go-client"
)

func Router() *mux.Router {
	ponzu := client.New(client.Config{
		Host:         "http://localhost:8080",
		DisableCache: true,
	})

	router := mux.NewRouter()

	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("content-type", "text/html")
		fmt.Fprint(res, `<p><a href="/events">See All Events</a></p>`)
	})

	router.HandleFunc("/about", func(res http.ResponseWriter, req *http.Request) {
		about, err := ponzu.Content("About", 1)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := `
			<h1>{{ .title }}</h1>
			<div>{{ .content | richtext}}</div>
		`

		t := template.New("aboutTmpl")
		t.Funcs(template.FuncMap{
			"richtext": func(data interface{}) template.HTML {
				return template.HTML(data.(string))
			},
		})

		t, err = t.Parse(tmpl)
		if err != nil {
			http.Error(res, "Error", http.StatusInternalServerError)
			return
		}

		fmt.Println(t.Execute(res, about.Data[0]))
	})

	router.HandleFunc("/events", func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			events, err := ponzu.Contents("Event", client.QueryOptions{Count: -1})
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			if req.URL.Query().Get("format") == "json" {
				res.Header().Set("content-type", "application/json")

				res.Write(events.JSON)
				return
			}

			tmpl := `
				<ul>
					{{range .}}
					<li>
						<h3><a href="/event/{{ .id }}">{{ .title }}</a></h3>
						<ol>
						{{range .details}}
							<li>{{ . }}</li>
						{{end}}
						</ol>
						<a href="{{ .ticket_link }}">Buy Tickets</a>
					</li>
					{{end}}
				</ul>

				<h2>Submit new event:</h2>
				<form action="/events" method="post" enctype="multipart/form-data">
					<label>Title <br/><input name="title"/></label><br/>
					<label>Details <br/><input name="details.0"/><br/><input name="details.1"/><br/><input name="details.2"/></label><br/>
					<label>Ticket Link <br/><input name="ticket_link"/></label><br/>
					<input type="submit" value="Submit"/>
				</form>
				`

			t := template.New("eventsList")

			t, err = t.Parse(tmpl)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			res.Header().Set("content-type", "text/html")
			t.Execute(res, events.Data)

		case http.MethodPost:
			err := req.ParseMultipartForm(1024 * 1000 * 4) // 4MB
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			resp, err := ponzu.Create("Event", req.PostForm, nil)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println(resp.Data)

			id := resp.Data[0]["id"].(float64)

			http.Redirect(res, req, fmt.Sprintf("/event/%.0f", id), http.StatusFound)
		}
	})

	router.HandleFunc("/event/{id}", func(res http.ResponseWriter, req *http.Request) {
		sid := mux.Vars(req)["id"]
		id, err := strconv.Atoi(sid)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		event, err := ponzu.Content("Event", id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl := `
		<p><a href="/events">&larr; All Events</a></p>
		<h3>{{ .title }}</h3>
		<ol>
			{{range .details}}
			<li>{{ . }}</li>
		{{end}}
		</ol>
		<a href="{{ .ticket_link }}">Buy Tickets</a>
		`
		t := template.New("eventId")

		t, err = t.Parse(tmpl)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		res.Header().Set("content-type", "text/html")
		t.Execute(res, event.Data[0])
	})

	return router
}

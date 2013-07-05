package speedtest

import (
    "fmt"
    "net/http"
    "os"
    "io"
    "html/template"
    "appengine"
    "appengine/blobstore"
    "appengine/urlfetch"
)

func init() {
    http.HandleFunc("/local/0.png", local)
    http.HandleFunc("/blob/0.png", blob)
    http.HandleFunc("/remote/0.png", remote)
    http.HandleFunc("/upload", upload)
    http.HandleFunc("/", handler)
}

// Upload to blobstore

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Header().Set("Content-Type", "text/plain")
        io.WriteString(w, "Internal Server Error")
        c.Errorf("%v", err)
}

var rootTemplate = template.Must(template.New("root").Parse(rootTemplateHTML))

const rootTemplateHTML = `
<html><body>
<form action="{{.}}" method="POST" enctype="multipart/form-data">
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`

func upload(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        uploadURL, err := blobstore.UploadURL(c, "/upload", nil)
        if err != nil {
                serveError(c, w, err)
                return
        }
        w.Header().Set("Content-Type", "text/html")
        err = rootTemplate.Execute(w, uploadURL)
        if err != nil {
                c.Errorf("%v", err)
        }
}

// Speed test

func local(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "image/png")
    f, err := os.Open("static/0.png")
    if err != nil { panic(err) }
    buf := make([]byte, 1024*1024)
    n, err := f.Read(buf)
    if err != nil { panic(err) }
    w.Write(buf[:n])
    f.Close()
}

func blob(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    w.Header().Set("Content-Type", "image/png")
    f := blobstore.NewReader(c, appengine.BlobKey("AMIfv97HhOdzO1aYQEe0QBrzbWSjSgWr2-JUxFJh_KnwxAhEdAqqK76TeE7vm5eDJW0ZoMwFVwur0Ub3t1kD_KzP3yJi4LIG6A-dCdJrJYafoJgH7SITCBum4MF9CY-C7na5fBulmKwQXd2mEYMyfk_RDgeQN1SZug"))
    buf := make([]byte, 1024*1024)
    n, err := f.Read(buf)
    if err != nil { panic(err) }
    w.Write(buf[:n])
}

func remote(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    client := urlfetch.Client(c)
    resp, err := client.Get("http://commondatastorage.googleapis.com/gombtiles/small/0/0/0.png")
    if err != nil { panic(err) }
    w.Header().Set("Content-Type", "image/png")
    buf := make([]byte, 1024*1024)
    n, err := resp.Body.Read(buf)
    if err != nil { panic(err) }
    w.Write(buf[:n])
}

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, "<h2>Speed test of a tile delivery</h2>")
    fmt.Fprint(w, "<a href='/static/0.png'>Static (file in the code repository)</a><br/>")
    fmt.Fprint(w, "<a href='/blob/0.png'>Blob (file stored in the blobstore)</a><br/>")
    // fmt.Fprint(w, "<a href='/local/0.png'>Local open/read/write by a golang function (file part of the code repository)</a><p/>")
    fmt.Fprint(w, "<a href='http://commondatastorage.googleapis.com/gombtiles/small/0/0/0.png'>Direct Cloud Storage (file stored in the Cloud Storage)</a><br/>")
    fmt.Fprint(w, "<a href='/remote/0.png'>Remote urlfetch (file stored in the Cloud Storage)</a><br/>")
}
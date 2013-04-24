package main

import (
    "fmt"
    //"os"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type Thread struct {
    Posts    []*Post
    Board    string
    Time_rcv time.Time
}

type Post struct {
    Tim int64
    Ext string
}

func readURL(url string) []byte {
    resp, err := http.Get(url)

    if err != nil {
        log.Fatal(err)
    }

    body, err := ioutil.ReadAll(resp.Body)

    defer resp.Body.Close()
    if err != nil {
        log.Fatal(err)
    }

    return body
}

//Parse JSON into Go Thread struct
func parseJSON(jsonObj []byte) *Thread {

    fmt.Println("Parsing JSON")
    var t Thread
    err := json.Unmarshal(jsonObj, &t)

    if err != nil {
            log.Fatal(err)
    }
    fmt.Println("Done parsing JSON\n")

    return &t

}

//Download post image
func downloadImage(p *Post, board string) {

    url := strings.Join([]string{"http://images.4chan.org/", board, "/src/", strconv.FormatInt(p.Tim, 10), p.Ext}, "")
    fmt.Println("Downloading image from", url)

    img := readURL(url)

    err := ioutil.WriteFile((strconv.FormatInt(p.Tim, 10) + p.Ext), img, 0755)
    if err != nil {
         log.Fatal(err)
    }

    fmt.Println("Done downloading image from", url)

}

func main() {
	url := "http://boards.4chan.org/a/res/84158572"
	tmp := strings.Split(url, "/")
	fmt.Println(tmp)
	boardName := tmp[3]
	
	json = strings.Join( []string{"http://api.4chan.org/", boardName, "/res/", tmp[5], ".json"}, "" )
    resp := readURL(json)

    t := parseJSON(resp)
	t.Board = boardName
    fmt.Println("Board is:", t.Board)
    
    t.Time_rcv = time.Now()

    fmt.Println("Thread received at:", t.Time_rcv)

    fmt.Println("Starting image download")
    for _, p := range t.Posts {

        if !(p.Tim == 0) {
            fmt.Println(strconv.FormatInt(p.Tim, 10), p.Ext)
            //downloadImage(p, t.Board)
        }

    }
	
	fmt.Println("Done downloading images")

}
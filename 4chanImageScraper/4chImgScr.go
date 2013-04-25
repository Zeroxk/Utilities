package main

import (
    "fmt"
    "os"
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
	Cooldown int64
	LastPost *Post
}

type Post struct {
	No int64
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

    fmt.Println("\nParsing JSON")
	
	if len(jsonObj) == 0 {  
		return nil
	}
	
    var t Thread
    err := json.Unmarshal(jsonObj, &t)

    if err != nil {
            log.Fatal(err)
    }
    fmt.Println("Done parsing JSON\n")

    return &t

}

//Download thread images
func downloadImages(posts []*Post, board string) (lastPost *Post) {

	fmt.Println("Starting image downloads")
    for i, p := range(posts) {
		
		if i == len(posts)-1 {
			lastPost = p
		}
		
        if !(p.Tim == 0) {
            fmt.Println(strconv.FormatInt(p.Tim, 10), p.Ext)
            url := strings.Join([]string{"http://images.4chan.org/", board, "/src/", strconv.FormatInt(p.Tim, 10), p.Ext}, "")
			fmt.Println("Downloading image from", url)

			img := readURL(url)

			err := ioutil.WriteFile((strconv.FormatInt(p.Tim, 10) + p.Ext), img, 0755)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Done downloading image from", url)
        }

    }
	fmt.Println("Done downloading images from thread")

	return

}

func main() {
	url := "http://boards.4chan.org/g/res/33294249"
	fmt.Println("Thread is:", url)
	tmp := strings.Split(url, "/")
	boardName := tmp[3]
	
	json := strings.Join( []string{"http://api.4chan.org/", boardName, "/res/", tmp[5], ".json"}, "" )
    resp := readURL(json)

    t := parseJSON(resp)
	if t == nil {
		fmt.Println("Thread has died, stopping")
		os.Exit(0)
	}
	t.Board = boardName
	t.Cooldown = 10
    fmt.Println("Board is:", t.Board)
    
    t.Time_rcv = time.Now()

    fmt.Println("Thread received at:", t.Time_rcv)

    t.LastPost = downloadImages(t.Posts, t.Board)
	
	fmt.Println("Last post is:", strconv.FormatInt(t.LastPost.No, 10))
	
	//t := time.Now()
	//url := "http://www.cs.bell-labs.com/who/dmr/" 
	
	fmt.Println("Sleeping")
	time.Sleep(time.Duration(t.Cooldown)*time.Second)
	fmt.Println("Woke up")
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("If-Modified-Since", t.Time_rcv.Format("Mon, 2 Jan 2006 15:04:05 GMT"))
	//fmt.Println(req.Header.Get("If-Modified-Since"))
	
	r, err := http.DefaultClient.Do(req)
	
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Status code:", r.StatusCode)
	fmt.Println("Last modified:", r.Header.Get("Last-Modified"))
	if sc := r.StatusCode; sc == 304 {
		fmt.Println("Nothing happened")
	}else {
		fmt.Println("Something happened")
	}
	
	//fmt.Println("Done downloading images")

}
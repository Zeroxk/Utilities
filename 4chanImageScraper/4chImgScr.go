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
	"bufio"
)

type Thread struct {
    Posts    []*Post
    Board    string
    Time_rcv string
	Cooldown int64
	LastPost int
}

type Post struct {
	No int64
    Tim int64
    Ext string
}

//Reads url, returns url body as byte slice and Last-Modified header
func readURL(url string) ([]byte, string) {

    resp, err := http.Get(url)

    if err != nil {
        log.Fatal(err)
    }

    body, err := ioutil.ReadAll(resp.Body)

    defer resp.Body.Close()
    if err != nil {
        log.Fatal(err)
    }
		
	return body, resp.Header.Get("Last-Modified")
}

//Parse JSON into partial Go Thread structure
func parseJSON(jsonObj []byte, t *Thread) *Thread {

	fmt.Println("Parsing JSON")
	
    err := json.Unmarshal(jsonObj, t)

    if err != nil {
            log.Fatal(err)
    }
	t.LastPost = len(t.Posts)-1
    fmt.Println("Done parsing JSON\n")
	
	return t

}

//Download thread images
func downloadImages(posts []*Post, board string) {

	fmt.Println("Starting image downloads")
    for _, p := range(posts) {
		
        if !(p.Tim == 0) {
            //fmt.Println(strconv.FormatInt(p.Tim, 10), p.Ext)
            url := strings.Join([]string{"http://images.4chan.org/", board, "/src/", strconv.FormatInt(p.Tim, 10), p.Ext}, "")
			fmt.Println("Downloading image from", url)

			img,_ := readURL(url)

			err := ioutil.WriteFile((strconv.FormatInt(p.Tim, 10) + p.Ext), img, 0755)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Done downloading image from", url)
        }

    }
	fmt.Println("Done downloading images from thread\n")

}

//Creates complete Go Thread structure by parsing JSON object from url and adding more info
func get_Thread(url string) (thread *Thread, json string) {

	tmp := strings.Split(url, "/")
	
	thread = new(Thread)	
	thread.Board = tmp[3]
	
	thread.Cooldown = 10
	
	json = strings.Join( []string{"http://api.4chan.org/", strings.Join(tmp[3:], "/"), ".json"}, "" )
	
	fmt.Println("Reading url")
	jsonObj, lastMod := readURL(json)
	fmt.Println("Board is:", thread.Board)
	fmt.Println("Done reading url\n")
	thread.Time_rcv = lastMod
	
    thread = parseJSON(jsonObj, thread)

    fmt.Println("Thread last modified:", thread.Time_rcv)
	fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")
	
	return

}

//Updates a thread
func update(json string, thread *Thread) {

	jsonObj, lastMod := readURL(json)
	thread.Time_rcv = lastMod
			
	t := parseJSON(jsonObj, thread)

	postsDelta := t.Posts[thread.LastPost+1:]
	thread.Posts = append(thread.Posts, postsDelta...)
			
	thread.LastPost = t.LastPost

	fmt.Println("Thread last modified:", thread.Time_rcv)

	downloadImages(postsDelta, thread.Board)
	
	fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")
	
}

func main() {
	
	input := bufio.NewReader(os.Stdin)
	fmt.Printf("Url: ")

	var url, dir string
	fmt.Scanf("%s\n", &url)
	//url,_ = input.ReadString('\n')
	//fmt.Println(url)
	
	fmt.Printf("Directory: ")
	dir,_ = input.ReadString('\n')
	//fmt.Scanf("%s\n\n", &dir)
	dir = strings.Trim(dir, "\n")
	dir = strings.TrimSpace(dir)
	
	//fmt.Println(dir)
	
	err := os.Chdir(dir)
	
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Url is:", url)
	fmt.Println("Changed directory to:", dir, "\n")
	
	thread, json := get_Thread(url)
	
    downloadImages(thread.Posts, thread.Board)
		
	for {
	
		fmt.Println("Sleeping")
		time.Sleep(time.Duration(thread.Cooldown)*time.Second)
		fmt.Println("Woke up")
	
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("If-Modified-Since", thread.Time_rcv)
	
		r, err := http.DefaultClient.Do(req)
	
		if err != nil {
			log.Fatal(err)
		}
	
		switch sc := r.StatusCode; sc {
			case 404: fmt.Println("Thread has died, fun is over"); os.Exit(0)
			
			case 304: fmt.Println("Nothing new"); thread.Cooldown *=2
			
			default: fmt.Println("Thread has been updated")
					 thread.Cooldown = 10			
					 update(json, thread)
		}

		fmt.Println("Cooldown is:", thread.Cooldown, "\n")
	
	}
	

}
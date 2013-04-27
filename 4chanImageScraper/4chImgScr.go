package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

const (
    MAX_COOLDOWN     = 1800
    DEFAULT_COOLDOWN = 30
)

type Thread struct {
    Posts    []*Post
    Board    string
    Time_rcv string
    Cooldown int64
    LastPost int
    Dir      string
    Id       int
}

type Post struct {
    No  int64
    Tim int64
    Ext string
}

//Reads url, returns url body as byte slice and Last-Modified header
func readURL(url string) ([]byte, string) {

    resp, err := http.Get(url)

    if err != nil {
        log.Fatal(err)
    }

    body := make([]byte, 0)

    if resp.StatusCode != 404 {
        body, err = ioutil.ReadAll(resp.Body)

        defer resp.Body.Close()
        if err != nil {
            log.Fatal(err)
        }
    }

    return body, resp.Header.Get("Last-Modified")
}

//Parse JSON and fill into Go Thread struct
func parseJSON(jsonObj []byte, t *Thread) {

    fmt.Println("Parsing JSON")

    err := json.Unmarshal(jsonObj, t)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Done parsing JSON\n")

}

//Download thread images into its specified directory
func downloadImages(t *Thread) {

    fmt.Println("Starting image downloads")
    for i := t.LastPost; i < len(t.Posts); i++ {

        p := t.Posts[i]
        if !(p.Tim == 0) {
            //fmt.Println(strconv.FormatInt(p.Tim, 10), p.Ext)
            url := strings.Join([]string{"http://images.4chan.org/", t.Board, "/src/", strconv.FormatInt(p.Tim, 10), p.Ext}, "")
            fmt.Println("Downloading image from", url)

            img, _ := readURL(url)

            if len(img) != 0 {

                path := strings.Join([]string{t.Dir, `\`, strconv.FormatInt(p.Tim, 10), p.Ext}, "")
                err := ioutil.WriteFile(path, img, 0644)
                if err != nil {
                    log.Fatal(err)
                }
                fmt.Println("Done downloading image from", url)
            } else {
                fmt.Println("Image location has 404'd")
            }

        }

    }

    fmt.Println("Done downloading images from thread", t.Id, "\n")

}

//Creates complete Go Thread structure by parsing JSON object from url and adding more info
func get_Thread(url string) (thread *Thread, json string) {

    tmp := strings.Split(url, "/")

    thread = new(Thread)
    thread.Id, _ = strconv.Atoi(tmp[5])
    thread.Board = tmp[3]
    thread.Cooldown = DEFAULT_COOLDOWN

    json = strings.Join([]string{"http://api.4chan.org/", strings.Join(tmp[3:], "/"), ".json"}, "")

    fmt.Println("Reading url")
    jsonObj, lastMod := readURL(json)
    fmt.Println("Board is:", thread.Board)
    fmt.Println("Done reading url\n")
    thread.Time_rcv = lastMod

    parseJSON(jsonObj, thread)

    fmt.Println("Thread last modified:", thread.Time_rcv)

    return

}

//Updates a thread
func update(json string, thread *Thread) {

    jsonObj, lastMod := readURL(json)
    thread.Time_rcv = lastMod

    t := new(Thread)
    parseJSON(jsonObj, t)
    t.LastPost = len(t.Posts) - 1

    postsDelta := t.Posts[thread.LastPost+1:]
    fmt.Println(len(postsDelta), "new posts\n")
    thread.Posts = append(thread.Posts, postsDelta...)

    fmt.Println("Thread last modified:", thread.Time_rcv)

    downloadImages(thread)

    thread.LastPost = t.LastPost
    fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")

}

//Checks for duplicates in the thread's folder, calls external Java program ImgDupDel.jar
func checkDupes(dir string) {

    cmd := exec.Command("java", "-jar", "ImgDupDel.jar", dir)
    fmt.Println("Checking for dupes")
    str, err := cmd.Output()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(str))

}

func main() {

    input := bufio.NewReader(os.Stdin)
    var wg sync.WaitGroup

    for {

        fmt.Println("Leave inputs empty to signal end of input")
        fmt.Printf("Url: ")
        var url, dir string
        fmt.Scanf("%s\n", &url)

        fmt.Printf("Directory: ")
        dir, _ = input.ReadString('\n')
        dir = strings.Trim(dir, "\n")
        dir = strings.TrimSpace(dir)

        if url == "" && dir == "" {
            fmt.Println("Empty inputs, stopping program")
            break
        }

        go func(url, dir string) {

            wg.Add(1)
            dead := false

            fmt.Println("Url is:", url)
            fmt.Println("Directory is:", dir, "\n")

            thread, json := get_Thread(url)
            thread.Dir = dir

            downloadImages(thread)

            thread.LastPost = len(thread.Posts) - 1
            fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")

            for {

                fmt.Println("Thread", thread.Id, "is sleeping for", thread.Cooldown, "seconds\n")
                time.Sleep(time.Duration(thread.Cooldown) * time.Second)
                fmt.Println("Thread", thread.Id, "woke up")

                req, err := http.NewRequest("GET", url, nil)
                if err != nil {
                    log.Fatal(err)
                }
                req.Header.Add("If-Modified-Since", thread.Time_rcv)

                r, err := http.DefaultClient.Do(req)

                if err != nil {
                    log.Fatal(err)
                }

                fmt.Println("Status code for request response:", r.StatusCode)
                switch sc := r.StatusCode; sc {
                case 404:
                    fmt.Println("Thread", thread.Id, "died at time: ", time.Now())
                    checkDupes(thread.Dir)
                    wg.Done()
                    dead = true

                case 304:
                    fmt.Println("Nothing new for thread", thread.Id)
                    thread.Time_rcv = r.Header.Get("Last-Modified")
                    if tc := thread.Cooldown * 2; tc > MAX_COOLDOWN {
                        thread.Cooldown = MAX_COOLDOWN
                    } else {
                        thread.Cooldown = tc
                    }

                default:
                    fmt.Println("Thread", thread.Id, "has been updated")
                    thread.Cooldown = DEFAULT_COOLDOWN
                    update(json, thread)
                }

                if dead {
                    break
                }

            }

            fmt.Println("Goodbye thread", thread.Id)

        }(url, dir)

    }

    fmt.Println("Waiting on all threads to finish")
    wg.Wait()
    fmt.Println("All threads done, terminated gracefully")

}

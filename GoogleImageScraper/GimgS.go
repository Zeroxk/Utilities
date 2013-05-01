package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
)

const dir = "imgsGo\\"

type RespData struct {
    ResponseData Results
}

type Results struct {
    Results []Image
}

type Image struct {
    Url string
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

func main() {

    var keyword string
    var numDls int

    fmt.Print("Keyword:")
    fmt.Scanf("%s\n", &keyword)

    fmt.Print("# of images to dl:")
    fmt.Scanf("%d\n\n", &numDls)

    fmt.Println("Your keyword is:", keyword)
    fmt.Println("Downloading", strconv.Itoa(numDls), "images\n")

    var wg sync.WaitGroup

    for i := 0; i < numDls; i += 4 {

        wg.Add(1)

        start := strconv.Itoa(i)

        go func(start, keyword string) {

            query := "https://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&start=" + start + "&q=" + keyword
            fmt.Println(query)

            contents := readURL(query)

            fmt.Println("Parsing JSON")

            rpD := new(RespData)
            err := json.Unmarshal(contents, &rpD)

            if err != nil {
                log.Fatal(err)
            }
            fmt.Println("Done parsing JSON\n")

            if _, err := os.Stat(dir); os.IsNotExist(err) {
                os.Mkdir(dir, 0755)
            }

            fmt.Println("Starting image download")

            res := rpD.ResponseData.Results
            for _, img := range res {

                name := img.Url[(strings.LastIndex(img.Url, "/"))+1:]
                fmt.Println(name)
                image := readURL(img.Url)

                err = ioutil.WriteFile((dir + name), image, 0755)
                if err != nil {
                    log.Fatal(err)
                }
            }
            wg.Done()

        }(start, keyword)

    }

    wg.Wait()

    fmt.Println("Finished")
}

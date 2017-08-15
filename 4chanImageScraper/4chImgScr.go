package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//MaxCooldown Maximum seconds a thread will sleep
	MaxCooldown = 3600
	//DefaultCooldown Seconds a thread will sleep by default
	DefaultCooldown = 60
	//APIBaseThreadURL Usage: https://a.4cdn/[board]/thread/[threadnumber].json
	APIBaseThreadURL = "https://a.4cdn.org/"
	//ImageBaseURL Usage: https://i.4cdn.org/[board]/[filename]
	ImageBaseURL = "https://i.4cdn.org/"
)

type Thread struct {
	Posts       []*Post
	Board       string
	TimeRcv     string
	Cooldown    int64
	LastPost    int
	Dir         string
	ID          int
	SemanticURL string `json"semantic_url"`
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
func downloadImages(t *Thread, start int) {

	fmt.Println("Starting image downloads")
	baseURL := strings.Join([]string{ImageBaseURL, t.Board, "/src/"}, "")

	for i := start; i < len(t.Posts); i++ {

		p := t.Posts[i]
		if p.Tim != 0 {
			//fmt.Println(strconv.FormatInt(p.Tim, 10), p.Ext)
			url := strings.Join([]string{baseURL, strconv.FormatInt(p.Tim, 10), p.Ext}, "")
			fmt.Println("Downloading image from", url)

			img, _ := readURL(url)

			if len(img) != 0 {

				path := strings.Join([]string{t.Dir, strconv.FormatInt(p.Tim, 10), p.Ext}, "")
				err := ioutil.WriteFile(path, img, 0644)

				if err != nil {
					fmt.Println("Error while saving image")
					log.Println(err)
					continue
				}

				fmt.Println("Done downloading image from", url)

			} else {
				fmt.Println("Image location has 404'd")
			}

		}

	}

	fmt.Println("Done downloading images from thread", t.ID, "\n")

}

//Creates complete Go Thread structure by parsing JSON object from url and adding more info
func getThread(url string) (thread *Thread, json string) {

	tmp := strings.Split(url, "/")

	thread = new(Thread)
	thread.ID, _ = strconv.Atoi(tmp[5])
	thread.Board = tmp[3]
	thread.Cooldown = DefaultCooldown

	json = strings.Join([]string{APIBaseThreadURL, strings.Join(tmp[3:6], "/"), ".json"}, "")
	fmt.Println("Json url", json)
	fmt.Println("Reading url")
	jsonObj, lastMod := readURL(json)
	fmt.Println("Done reading url\n")
	thread.TimeRcv = lastMod

	parseJSON(jsonObj, thread)

	fmt.Println("Thread last modified:", thread.TimeRcv)

	return

}

//Updates a thread
func update(json string, thread *Thread) {

	jsonObj, lastMod := readURL(json)

	if len(jsonObj) != 0 {

		thread.TimeRcv = lastMod

		t := new(Thread)
		parseJSON(jsonObj, t)
		t.LastPost = len(t.Posts) - 1
		fmt.Println("Index old lastpost:", thread.LastPost)
		fmt.Println("Index new lastpost:", t.LastPost)

		numDelPosts := 0
		if len(t.Posts) < len(thread.Posts) {
			fmt.Println("Finding deleted posts")
			numDelPosts = findNumDelPosts(thread.Posts, t.Posts)
			fmt.Println("# of deleted posts:", numDelPosts, "\n")
		} else if t.Posts[t.LastPost].No == thread.Posts[thread.LastPost].No {
			fmt.Println("Threads are identical, update indicator was wrong")
			thread.Cooldown = MaxCooldown
			return
		}

		postsDelta := len(t.Posts) - (thread.LastPost + 1) - numDelPosts
		fmt.Println(postsDelta, "new posts\n")
		thread.Posts = t.Posts

		fmt.Println("Thread last modified:", thread.TimeRcv)

		downloadImages(thread, (thread.LastPost - numDelPosts))

		thread.LastPost = t.LastPost
		fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")

	} else {
		fmt.Println("Thread died while fetching update")
	}

}

//Checks for duplicates in the thread's folder, calls external Java program ImgDupDel.jar
func checkDupes(dir string) {

	cmd := exec.Command("java", "-jar", "ImgDupDel.jar", dir)
	fmt.Println("Checking for dupes")
	str, err := cmd.Output()

	fmt.Println(string(str))

	if err != nil {
		log.Println("Something went wrong, see java output")
	}

	fmt.Println("Done!")

}

func findNumDelPosts(old, new []*Post) int {

	num := 0
	if len(old) == 0 || len(new) == 0 {
		return 0
	}
	for i, p := range new {
		if p.No != old[i].No {
			if i == len(new)-1 {
				num++
			} else {
				num += findNumDelPosts(old[i+1:], new[i+1:])
			}
			break
		}
	}

	return num

}

func validURL(url string) bool {
	return strings.HasPrefix(url, "http://boards.4chan.org/") || strings.HasPrefix(url, "https://boards.4chan.org")
}

func validPath(dir string) bool {
	return filepath.IsAbs(dir)
}

func main() {

	input := bufio.NewReader(os.Stdin)
	var wg sync.WaitGroup
	wantDupes := false

	if len(os.Args) == 2 {
		if arg := os.Args[1]; arg == "--dupes" {
			wantDupes = true
		}
	}
	fmt.Println("Dupe checking is:", wantDupes)

	for {

		fmt.Println("Leave both of the inputs empty to signal end of input")
		fmt.Printf("Url: ")
		var url, dir string
		fmt.Scanf("%s\n", &url)

		fmt.Printf("Directory: ")
		dir, _ = input.ReadString('\n')
		dir = strings.Trim(dir, "\n")
		dir = strings.TrimSpace(dir)
		dir = filepath.FromSlash(dir)
		if !strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir += string(os.PathSeparator)
		}

		if url == "" && dir == "" {
			fmt.Println("Empty inputs, stopping program")
			break
		}

		if validURL(url) && validPath(dir) {

			go func(url, dir string) {

				wg.Add(1)
				dead := false

				fmt.Println("Url is:", url)
				fmt.Println("Directory is:", dir, "\n")

				thread, json := getThread(url)
				thread.Dir = dir

				downloadImages(thread, 0)

				thread.LastPost = len(thread.Posts) - 1
				fmt.Println("Last post is:", strconv.FormatInt(thread.Posts[thread.LastPost].No, 10), "\n")

				for !dead {

					fmt.Println("Thread", thread.SemanticURL, "is sleeping for", thread.Cooldown, "seconds\n")
					time.Sleep(time.Duration(thread.Cooldown) * time.Second)
					fmt.Println("Thread", thread.SemanticURL, "woke up")

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						log.Println(err)
						continue
					}
					req.Header.Add("If-Modified-Since", thread.TimeRcv)

					r, err := http.DefaultClient.Do(req)

					if err != nil {
						log.Println(err)
						continue
					}

					fmt.Println("Status code for request response:", r.StatusCode)
					switch sc := r.StatusCode; sc {
					case 404:
						fmt.Println("Thread", thread.SemanticURL, "died at time: ", time.Now())

						if wantDupes {
							checkDupes(thread.Dir)
						}

						wg.Done()
						dead = true

					case 304:
						fmt.Println("Nothing new for thread", thread.SemanticURL)
						thread.TimeRcv = r.Header.Get("Last-Modified")

						if tc := thread.Cooldown * 2; tc > MaxCooldown {
							fmt.Println("Time is now:", time.Now())
							thread.Cooldown = MaxCooldown
						} else {
							thread.Cooldown = tc
						}

					case 200:
						fmt.Println("Thread", thread.SemanticURL, "has been updated")
						thread.Cooldown = DefaultCooldown
						update(json, thread)

					default:
						fmt.Println("Status code other than 304, 404 and 200, continuing")
					}

				}

				fmt.Println("Goodbye thread", thread.SemanticURL, "\n")

			}(url, dir)

		} else {
			fmt.Println("Not valid url or directory")
		}

	}

	fmt.Println("Waiting on all threads to finish")
	wg.Wait()
	fmt.Println("All threads done, terminated gracefully")

}

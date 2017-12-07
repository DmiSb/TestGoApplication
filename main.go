package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Stat struct {
	Pos  int
	Last time.Time
}

var (
	stats map[string]Stat
	conf  Conf
	conn  redis.Conn
)

func homeBidder(w http.ResponseWriter, r *http.Request) {
	var rec Data
	err := json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		fmt.Fprintf(w, "Error: "+err.Error())
	} else {
		if last, err := redis.Int64(conn.Do("GET", "time:"+rec.Device.Ifa)); err == nil {
			delta := float64(time.Now().Unix() - last)
			if delta > conf.AppConf.SeriaMax {
				conn.Do("SET", "pos:"+rec.Device.Ifa, 0)
				fmt.Fprintf(w, "{\"pos:\"%d}\n", 0)
			} else {
				if delta > conf.AppConf.SeriaDelay {
					if n, err := conn.Do("INCR", "pos:"+rec.Device.Ifa); err == nil {
						fmt.Fprintf(w, "{\"pos:\"%d}\n", n)
					}
				} else {
					if n, err := redis.Int(conn.Do("GET", "pos:"+rec.Device.Ifa)); err == nil {
						fmt.Fprintf(w, "{\"pos:\"%d}\n", n)
					}
				}
			}
			conn.Do("SET", "time:"+rec.Device.Ifa, time.Now().Unix())
		} else {
			conn.Do("SET", "time:"+rec.Device.Ifa, time.Now().Unix())
			conn.Do("SET", "pos:"+rec.Device.Ifa, 0)
			fmt.Fprintf(w, "{\"pos:\"%d}\n", 0)
		}

		/* Save data to map
		if s, ok := stats[rec.Device.Ifa]; ok {
			delta := time.Now().Sub(s.Last)
			if delta.Seconds() > conf.AppConf.SeriaMax {
				s.Pos = 0
				s.Last = time.Now()
			} else {
				if delta.Seconds() > conf.AppConf.SeriaDelay {
					s.Pos++
					s.Last = time.Now()
				}
			}
			stats[rec.Device.Ifa] = s
			fmt.Fprintf(w, "{\"pos:\"%d}\n", s.Pos)
		} else {
			stats[rec.Device.Ifa] = Stat{
				0,
				time.Now(),
			}

			fmt.Fprintf(w, "{\"pos:\"%d}\n", 0)
		}*/

		key := fmt.Sprintf("stat:%s:%s:%s",
			rec.Device.Geo.Country,
			rec.Device.Os,
			rec.App.Bundle,
		)
		_, err = redis.Int(conn.Do("INCR", key))
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func statsBidder(w http.ResponseWriter, r *http.Request) {
	var ss []string
	var country string
	var os string
	var app string
	var count int
	var result = ""
	var p = ""

	keys, err := redis.Strings(conn.Do("KEYS", "stat:*"))
	if err == nil {
		for _, key := range keys {
			count, err = redis.Int(conn.Do("GET", key))
			if count > 0 {
				country = ""
				os = ""
				app = ""
				ss = strings.Split(key, ":")

				if len(ss) > 1 {
					country = ss[1]
				}
				if len(ss) > 2 {
					os = ss[2]
				}
				if len(ss) > 3 {
					app = ss[3]
				}

				r, err := json.Marshal(&struct {
					Country string `json:"country"`
					Os      string `json:"os"`
					App     string `json:"app"`
					Count   int    `json:"count"`
				}{
					Country: country,
					Os:      os,
					App:     app,
					Count:   count,
				})
				if err == nil {
					result += p + string(r)
					p = ","
				}
			}
		}
	}
	fmt.Fprintf(w, "{[%s]}\n", result)
}

func handleRequest() {
	http.HandleFunc("/", homeBidder)
	http.HandleFunc("/stats", statsBidder)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initConf() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&conf)
	if err != nil {
		panic("Error reading config file")
	}
}

func initDb() {
	var err error
	url := fmt.Sprintf("redis://user:%s@%s:%s",
		conf.DbConf.User,
		conf.DbConf.Host,
		conf.DbConf.Port,
	)
	conn, err = redis.DialURL(url)
	if err != nil {
		panic("Error connect to db")
	} else {
		log.Printf("Sucess connect to %s \n", url)
	}
}

func main() {
	stats = make(map[string]Stat)
	initConf()
	initDb()
	handleRequest()
}

// curl -sS  -H "Content-Type: application/json" --data @data1.json http://localhost:8080/
// curl -sS  -H "Content-Type: application/json" http://localhost:8080/stats

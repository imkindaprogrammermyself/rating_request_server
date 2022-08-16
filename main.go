package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go.uber.org/ratelimit"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"notyourfather/pr_server/data"
	"notyourfather/pr_server/utils"
	"os"
	"reflect"
	"strconv"
	"sync"
)

var REALMS = map[string]string{
	"ASIA": "https://api.worldofwarships.asia/wows/ships/stats/",
	"NA":   "https://api.worldofwarships.com/wows/ships/stats/",
	"EU":   "https://api.worldofwarships.eu/wows/ships/stats/",
	"RU":   "https://api.worldofwarships.ru/wows/ships/stats/",
}

var REQUIRED_ENV = [...]string{"APPLICATION_ID", "SERVER_TYPE", "SERVER_HOST", "SERVER_PORT", "RATE_LIMIT"}
var EXPECTED = dat.ExpectedValues{}

func main() {
	env_file := flag.String("env", ".env", ".env file")
	ev_file := flag.String("ev", "expected.json", "expected.json file")
	flag.Parse()

	err := godotenv.Load(*env_file)

	if err != nil {
		log.Fatalf("Error loading .env file. %s", err.Error())
	}

	for _, env_var := range REQUIRED_ENV {
		_, exists := os.LookupEnv(env_var)
		if !exists {
			log.Fatalf("Error! Missing [%s] environment variable.", env_var)
		}
	}

	rateLimitVal, rtcErr := strconv.Atoi(os.Getenv("RATE_LIMIT"))

	if rtcErr != nil {
		log.Fatalf("Invalid RATE_LIMIT value. %s\n", rtcErr.Error())
	}

	log.Printf("Rate limit is set to [%d]", rateLimitVal)

	rateLimit := ratelimit.New(rateLimitVal, ratelimit.WithoutSlack)
	content, load_err := os.ReadFile(*ev_file)

	if load_err != nil {
		log.Fatalf("Error! [%s] doesn't exists. %s", *ev_file, load_err.Error())
	}

	json_load_err := json.Unmarshal(content, &EXPECTED)

	if json_load_err != nil {
		log.Fatalln(reflect.TypeOf(EXPECTED.Data), json_load_err.Error())
	}

	serverType := os.Getenv("SERVER_TYPE")
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")

	log.Println("Starting PR server...")
	server, err := net.Listen(serverType, serverHost+":"+serverPort)
	if err != nil {
		log.Fatalf("Error listening: %s", err.Error())
	}
	defer server.Close()
	log.Printf("Listening on %s:%s\n", serverHost, serverPort)
	log.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			log.Fatalf("Error accepting: %s", err.Error())
		}
		log.Printf("Received connection from %s", connection.RemoteAddr().String())
		go processClient(connection, &rateLimit)
	}
}
func processClient(connection net.Conn, rateLimiter *ratelimit.Limiter) {
	defer connection.Close()

	bPayloadSize := make([]byte, 4)
	connection.Read(bPayloadSize)

	payloadSize := binary.LittleEndian.Uint32(bPayloadSize)
	bPayload := make([]byte, payloadSize)
	connection.Read(bPayload)

	request_data := dat.RequestPayload{}
	json.Unmarshal(bPayload, &request_data)
	result := fetchAll(request_data, rateLimiter)

	response, _ := json.Marshal(result)

	responseBuff := make([]byte, 4)
	binary.LittleEndian.PutUint32(responseBuff, uint32(len(response)))
	connection.Write(responseBuff)
	connection.Write([]byte(response))
	log.Printf("Data is sent to %s", connection.RemoteAddr().String())
}

func fetchAll(payload dat.RequestPayload, rateLimiter *ratelimit.Limiter) dat.PlayerRatings {
	var waitGroup sync.WaitGroup
	result := make(chan dat.PlayerRating)

	for _, playerInfo := range payload {
		waitGroup.Add(1)
		go func(pi dat.PlayerInfo) {
			result <- fetch(pi, rateLimiter)
			waitGroup.Done()
		}(playerInfo)
	}

	go func() {
		waitGroup.Wait()
		close(result)
	}()

	response := dat.PlayerRatings{}

	for r := range result {
		response = append(response, r)
	}

	return response
}

func fetch(playerInfo dat.PlayerInfo, rateLimit *ratelimit.Limiter) dat.PlayerRating {
	rt := *rateLimit
	rt.Take()

	data := url.Values{
		"application_id": {os.Getenv("APPLICATION_ID")},
		"extra":          {"pvp_solo"},
		"fields":         {"ship_id, pvp_solo.wins, pvp_solo.damage_dealt, pvp_solo.battles, pvp_solo.frags"},
		"in_garage":      {"0"},
		"language":       {"en"},
	}
	data.Set("account_id", fmt.Sprintf("%d", playerInfo.AccountID))

	resp, err := http.PostForm(REALMS[playerInfo.Realm], data)

	if err != nil {
		log.Fatal(err.Error())
		return dat.PlayerRating{ID: playerInfo.ID, Rating: -1}
	}

	result := dat.PlayerShipsStats{}
	decodeError := json.NewDecoder(resp.Body).Decode(&result)

	if decodeError != nil {
		log.Println(decodeError)
		return dat.PlayerRating{ID: playerInfo.ID, Rating: -1}
	}

	if result.Status == "error" {
		log.Println(result.Error)
		return dat.PlayerRating{ID: playerInfo.ID, Rating: -4}
	}

	if result.Meta.Hidden != nil {
		return dat.PlayerRating{ID: playerInfo.ID, Rating: -3}
	}

	if len(result.Data) == 0 {
		return dat.PlayerRating{ID: playerInfo.ID, Rating: -2}
	}

	a_dmgs := []uint64{}
	a_frgs := []int{}
	a_wins := []int{}

	e_dmgs := []float64{}
	e_frgs := []float64{}
	e_wins := []float64{}

	for _, pss := range result.Data[playerInfo.AccountID] {
		if val, ok := EXPECTED.Data[pss.ShipId]; ok {
			pvp_solo := pss.PvpSolo
			pvp_battles := pvp_solo.Battles

			if pvp_battles == 0 {
				continue
			}

			if v, ok := val.(map[string]interface{}); ok {
				a_dmgs = append(a_dmgs, pvp_solo.DamageDealt)
				a_frgs = append(a_frgs, pvp_solo.Frags)
				a_wins = append(a_wins, pvp_solo.Wins)

				ev_avg_dmg := v["average_damage_dealt"].(float64)
				ev_avg_frg := v["average_frags"].(float64)
				ev_win := v["win_rate"].(float64)

				e_dmgs = append(e_dmgs, ev_avg_dmg*float64(pvp_battles))
				e_frgs = append(e_frgs, ev_avg_frg*float64(pvp_battles))
				e_wins = append(e_wins, (float64(pvp_battles)*ev_win)/100)

			}
		}
	}

	r_dmg := float64(utils.Sum_uint64(a_dmgs)) / utils.Sum_floats64(e_dmgs)
	r_frg := float64(utils.Sum_int(a_frgs)) / utils.Sum_floats64(e_frgs)
	r_win := float64(utils.Sum_int(a_wins)) / utils.Sum_floats64(e_wins)

	n_dmg := math.Max(0, (r_dmg-0.4)/(1-0.4))
	n_frg := math.Max(0, (r_frg-0.1)/(1-0.1))
	n_win := math.Max(0, (r_win-0.7)/(1-0.7))

	pr := 700*n_dmg + 300*n_frg + 150*n_win

	return dat.PlayerRating{ID: playerInfo.ID, Rating: pr}
}

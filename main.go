package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	fegUSD float64
	fegCAD float64
	fegVND float64
	vndCAD float64
	vndUSA float64
	eurVND float64
	eurCAD float64
	eurUSD float64
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		Crawl(u, depth-1, fetcher)
	}
	return
}



func main() {
	go updateFeg()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error authenticating discord")
	}

	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	//httpClient := &http.Client{
	//	Timeout: time.Second * 10,
	//}

	//cg := coingecko2.NewClient(httpClient)

	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if strings.HasPrefix(m.Content, "f/") {
		cont:=m.Content[2:len(m.Content)]
		cmd := strings.Split(cont, " ")
		if cmd[0] == "set" && cmd[1] == "balance" {
			balance, err := getBalance(m.Author.ID)
			if err != nil {
				log.Fatal("Can not read wallets")
			}
			_,_ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v", balance))
		} else if cmd[0] == "balance" {

		} else if cmd[0] == "feg" || cmd[0] == "price" {

		} else if cmd[0] == "add" && cmd[1] == "warning" {

		} else if cmd[0] == "remove" && cmd[1] == "warning" {

		} else {

		}
		_,_ = s.ChannelMessageSend(m.ChannelID, m.Author.ID)
	}
}

func updateFeg() {
	for true {
		log.Println("Requesting api")
		resp, err := http.Get("https://api.coingecko.com/api/v3/coins/feg-token?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var result map[string]interface{}

		err = json.Unmarshal([]byte(body), &result)
		if err != nil {
			panic(err)
		}

		for i, v := range result {
			//log.Println("I: ", i)
			//log.Println("V: ", v)
			if i == "market_data" {

				marketData := v.(map[string]interface{})
				fegVND, _ = strconv.ParseFloat(fmt.Sprintf("%v", marketData["current_price"].(map[string]interface{})["vnd"]), 64)
				//log.Println(fegVND)
			}

		}

		//log.Println("Requesting api")
		resp, err = http.Get("http://data.fixer.io/api/latest?access_key=" + os.Getenv("ACCESS_KEY") + "&format=1&symbols=VND,CAD,USD")
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		var result2 map[string]interface{}

		err = json.Unmarshal([]byte(body), &result2)
		//log.Println("RESULT: ", result2)
		if err != nil {
			panic(err)
		}

		for i, v := range result2 {
			//log.Println("I: ", i)
			//log.Println("V: ", v)
			if i == "rates" {

				rates := v.(map[string]interface{})
				eurVND, _ = strconv.ParseFloat(fmt.Sprintf("%v", rates["VND"]), 64)
				//log.Println(eurVND)
				eurCAD, _ = strconv.ParseFloat(fmt.Sprintf("%v", rates["CAD"]), 64)
				//log.Println(eurCAD)
				eurUSD, _ = strconv.ParseFloat(fmt.Sprintf("%v", rates["USD"]), 64)
				//log.Println(eurUSD)
			}

		}
	}
	time.Sleep(1 * time.Second)
}

func calculate(userID string) float64 {
	bal, err := getBalance(userID)
	if err != nil {
		log.Fatal("Can not read wallets")
	}

	return bal
}

func getBalance(userID string) (float64, error) {
	file, err := os.Open("./wallets")
	if err != nil {
		return 0, err
	}

	var lines []string
	i := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		i++
	}

	for _, v := range lines {
		if strings.HasPrefix(v, userID) {
			//bal, err := strconv.ParseFloat(strings.Split(v, " ")[1], 64)
			//if err != nil {
			//	return 0, err
			//}

			wallet := strings.Split(v, " ")[1]
			log.Println("WALLET: ", wallet)
			resp, err := http.Get("https://etherscan.io/address/" + wallet + "#tokentxns")
			if err != nil {
				return 0, err
			}
			log.Println("BODY: ", resp)

			return 1, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}
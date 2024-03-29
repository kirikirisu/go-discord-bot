package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

var (
	Token string
)

type Data struct {
	Weather []Weather `json:"weather"`
	Main    Main      `json:"main"`
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	ICON        string `json:"icon"`
}

type Main struct {
	Temp      float32 `json:"temp"`
	FeelsLike float32 `json:"feels_like"`
	TempMin   float32 `json:"temp_min"`
	TempMax   float32 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
}

type TodoList struct {
	Todos []Todo
}

type Todo struct {
	Id        int
	CreatedAt time.Time
	Content   string
	Active    bool
}

var Db *sql.DB

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	var err error
	Db, err = sql.Open("postgres", "user=discordbot dbname=discordbot password=discordbot sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Dicord session,", err)
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error open connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func getWeather(wch chan Data) {
	url := "http://api.openweathermap.org/data/2.5/weather?q=nagano&appid=2b32f39866053d126c78cf3dbe0dbe0d"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data Data

	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	wch <- data
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "weather" {
		wch := make(chan Data, 1)
		go getWeather(wch)
		ws := <-wch
		w := ws.Weather[0].Main

		s.ChannelMessageSend(m.ChannelID, w)
	}

	if strings.HasPrefix(m.Content, "!set") {
		words := strings.Fields(m.Content)

		if len(words) == 1 {
			s.ChannelMessageSend(m.ChannelID, "plese input todo")
		}

		if len(words) == 2 {
			todo := Todo{Content: words[1], CreatedAt: time.Now()}
			err := todo.Create()
			if err != nil {
				fmt.Println("error insert todo", err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "save!!")
		}
	}
}

func (todo *Todo) Create() (err error) {
	err = Db.QueryRow("insert into todos (content, created_at) values ($1, $2) returning id", todo.Content, todo.CreatedAt).Scan(&todo.Id)
	return
}

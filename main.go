package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/mail"
	"strconv"
	"time"

	yaml "gopkg.in/yaml.v3"

	_ "github.com/denisenkom/go-mssqldb"
	strip "github.com/grokify/html-strip-tags-go" // => strip
)

var (
	query      string
	m          BaseMsg
	mHead      BaseMsg2
	botMessage BotMessage
	startID    int
	offset     int
	logged     int
	timeSynh   time.Time
	userex     userExists
)

func main() {
	startID = 425740
	textfile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config := Config{}
	err3 := yaml.Unmarshal([]byte(textfile), &config)
	if err3 != nil {
		log.Fatalf("error: %v", err)
	}
	botToken := config.Token
	//https://api.telegram.org/bot<token>/METHOD_NAME
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi + botToken
	conString := fmt.Sprintf("user id=%s;password=%s;port=%d;database=%s", config.Db.User, config.Db.Password, config.Db.Port, config.Db.Database)
	db, err := sql.Open("mssql", conString)
	rand.Seed(time.Now().UnixNano())

	if err != nil {
		fmt.Println("Error in connect DB")
		log.Panic(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Error ping")
		fmt.Scanf(" ")
	}

	for {
		query = "SELECT top 1 id FROM [zaprosi].[dbo].[perepiska] order by id desc"
		fmt.Println(query)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}
		if rows != nil {
			for rows.Next() {
				//fmt.Println("nexter")

				if err := rows.Scan(&startID); err != nil {
					log.Panic(err)
					fmt.Scanf(" ")
				}
			}
		}
		startID2 := startID
		timer := time.NewTimer(time.Second * 1)
		<-timer.C

		updates, err := getUpdates(botUrl, offset)

		if err != nil {
			log.Println("err: ", err.Error())
		}
		for _, update := range updates {

			query = fmt.Sprintf("SELECT CASE WHEN EXISTS (SELECT TOP (1) 1 FROM [zaprosi].[dbo].[tgbot] WHERE [chatid] = %d) THEN '1'ELSE '0' END", update.Message.Chat.ChatId)
			fmt.Println(query)

			rows, err := db.Query(query)
			if err != nil {
				log.Fatal(err)
				fmt.Scanf("")
			}
			if rows != nil {
				for rows.Next() {
					//fmt.Println("nexter")

					if err := rows.Scan(&userex.userexisis); err != nil {
						log.Panic(err)
						fmt.Scanf(" ")
					}
				}
			}

			fmt.Println(userex.userexisis)

			if userex.userexisis == 0 {
				zapr := fmt.Sprintf("INSERT INTO [dbo].[tgbot] ([email],[chatid],[pin],[loginned]) VALUES (null,%d,null,%d)", update.Message.Chat.ChatId, 0)
				_, err = db.Exec(zapr)
				if err != nil {
					log.Fatal(err)
					fmt.Scanf("")
				}
			}

			//err = respond(botUrl, update)
			//query = fmt.Sprintf("SELECT [id],[email],[chatid] FROM [zaprosi].[dbo].[tgbot]  where chatid = %d", update.Message.Chat.ChatId)
			query = fmt.Sprintf("SELECT [loginned] FROM [zaprosi].[dbo].[tgbot] where chatid = %d", update.Message.Chat.ChatId)
			fmt.Println(query)
			rows, err = db.Query(query)
			fmt.Println(rows)

			if err != nil {
				log.Fatal(err)
				fmt.Scanf("")
			}
			if rows != nil {
				for rows.Next() {
					//fmt.Println("nexter")

					if err := rows.Scan(&logged); err != nil {
						log.Panic(err)
						fmt.Scanf(" ")
					}
					if logged != 1 {
						logged = 0
					}
				}
			} else {
				logged = 0
			}

			if logged == 1 {
				switch {
				case update.Message.Text == "heyo":
					err = sendMessage(botUrl, "darova", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
				default:
					err = sendMessage(botUrl, "Пока команды не доступны. Но скоро будут!", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				}

			} else {

				switch {
				case update.Message.Text == "/start":
					err = sendMessage(botUrl, `Привет! 
					Я - телеграм-бот для работы с порталом SPOC! 
					Пока что я умею оповещать тебя о сообщениях, которые направляют тебе твои коллеги. Скоро я смогу оповещать тебя о новых кейсах, поступивших в работу, через меня можно будет отвечать напрямую в переписку, я научусь настраиваться под твои нужды, и еще очень-очень многое.
					Для регистрации просто введи свой корпоративный email.
					Если у тебя возникла какая-то проблема со мной, или же хочется предлжить что-то для доработки - обращайся к soreshnikov@dtln.ru/ @ro_anae`, update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				case valid(update.Message.Text):
					pin := rand.Intn(8999) + 1000
					zapr := fmt.Sprintf("update [dbo].[tgbot] set [email] = '%s',pin = %d where chatid=%d", update.Message.Text, pin, update.Message.Chat.ChatId)
					_, err = db.Exec(zapr)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				case len(update.Message.Text) == 4:
					pin := update.Message.Text
					query = fmt.Sprintf("SELECT pin FROM [zaprosi].[dbo].[tgbot] where chatid=%d", update.Message.Chat.ChatId)
					fmt.Println(query)
					rows, err := db.Query(query)

					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

					for rows.Next() {
						fmt.Println("nexter")
						userpin := 0
						if err := rows.Scan(&userpin); err != nil {
							log.Panic(err)
							fmt.Scanf("")
						}
						userpinstring := strconv.Itoa(userpin)
						if pin == userpinstring {
							err = sendMessage(botUrl, "Регистрация прошла успешно. Добро пожаловать!", update.Message.Chat.ChatId)
							if err != nil {
								log.Fatal(err)
								fmt.Scanf("")
							}
							zapr := fmt.Sprintf("UPDATE [dbo].[tgbot] SET [loginned] = 1 WHERE chatid=%d", update.Message.Chat.ChatId)
							_, err = db.Exec(zapr)
							if err != nil {
								log.Fatal(err)
								fmt.Scanf("")
							}
						} else {
							err = sendMessage(botUrl, "Код введен неверно!\n Новый пин отправлен тебе на почту.", update.Message.Chat.ChatId)
							if err != nil {
								log.Fatal(err)
								fmt.Scanf("")
							}

							pin1 := rand.Intn(8999) + 1000
							zapr1 := fmt.Sprintf("UPDATE [dbo].[tgbot] SET [pin] = %d WHERE chatid=%d", pin1, update.Message.Chat.ChatId)
							fmt.Println(zapr1)

							_, err = db.Exec(zapr1)

							if err != nil {
								log.Fatal(err)
								fmt.Scanf("")
							}

						}

					}

				default:
					err = sendMessage(botUrl, "Вы не зарегистрированы. Введите свой корпоративный email.", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				}
			}

			offset = update.UpdateId + 1
		}
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}
		fmt.Println(updates)

		query = fmt.Sprintf("With asd AS(SELECT [komyt],[datewhen],[nid],[mainid],[nproject],[valid] FROM [zaprosi].[dbo].[komy_table]),prichit AS (SELECT [dateprochit],[whoClick],[idN],[nproj] FROM [zaprosi].[dbo].[prochit_table]),usery AS (SELECT [shnam],[name],[secnam],[about],[phone],[email],[pozic],[dolzn],[pic],[groupa],[uwelSotr] FROM [zaprosi].[dbo].[profile])SELECT TOP (1000) [perepiska].[id],[shnam],[name],[secnam],[komyt],[kto],[perepiska].[nproject],[message],ISNULL([naprav],'main') AS naprav,[mainid], (select chatid from [zaprosi].[dbo].[tgbot] where email = komyt ) as chat FROM [zaprosi].[dbo].[perepiska]INNER JOIN asd ON [nid] = [id]INNER JOIN [profile] ON [email] = [komyt] WHERE komyt in (SELECT[email] FROM [zaprosi].[dbo].[tgbot] where loginned = 1) and [perepiska].id>%d order by id desc", startID)
		fmt.Println(query)

		rows, err = db.Query(query)
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}

		for rows.Next() {

			if err := rows.Scan(&m.id, &m.login, &m.name, &m.secname, &m.whom, &m.who, &m.nproject, &m.message, &m.naprav, &m.mainid, &m.chat); err != nil {
				log.Panic(err)
				fmt.Scanf(" ")
			}

			fmt.Println(m.message)

			msg := fmt.Sprintf("Номер запроса: %d \n Отправитель: %s \n Сообщение: \n %s", m.nproject, m.who, m.message)
			msg = strip.StripTags(msg)
			err := sendMessage(botUrl, msg, m.chat)
			if err != nil {

				log.Panic(err)

			}
			startID = m.id
		}

		query = fmt.Sprintf(`SELECT id,
			[kto]
			,[nproject]
			,[message]
			FROM [zaprosi].[dbo].[perepiska]  where nproject in (SELECT [n_project]
			FROM [zaprosi].[dbo].[zapr] where napr='ork' and mpp is null ) and id>%d`, startID2)
		fmt.Println(query)

		rows, err = db.Query(query)
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}

		for rows.Next() {

			if err := rows.Scan(&mHead.id, &mHead.who, &mHead.nproject, &mHead.message); err != nil {
				log.Panic(err)
				fmt.Scanf(" ")
			}

			fmt.Println(mHead.message)

			msg := fmt.Sprintf("Неназначенный\nНомер запроса: %d \n Отправитель: %s \n Сообщение: \n %s", mHead.nproject, mHead.who, mHead.message)
			msg = strip.StripTags(msg)
			err := sendMessage(botUrl, msg, 261609763)
			if err != nil {

				log.Panic(err)

			}
			startID2 = mHead.id
		}

		query = `SELECT [obn] FROM [zaprosi].[dbo].[obnov] where id = 1`
		fmt.Println(query)

		rows, err = db.Query(query)
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}

		for rows.Next() {

			if err := rows.Scan(&timeSynh); err != nil {
				log.Panic(err)
				fmt.Scanf(" ")
			}

			fmt.Println(timeSynh)

			now := time.Now()

			count := 1
			timeBottom := now.Add(time.Duration(-count) * time.Minute)

			if timeSynh.Before(timeBottom) {
				err := sendMessage(botUrl, "Ошибка синхронизации", 261609763)
				if err != nil {

					log.Panic(err)

				}
				err = sendMessage(botUrl, "Ошибка синхронизации", 319080225)
				if err != nil {

					log.Panic(err)

				}
			}
		}

	}

}

func sendMessage(botUrl string, msg string, chatId int) error {
	fmt.Println("sender")
	botMessage.ChatId = chatId
	botMessage.Text = msg
	buf, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}

func getUpdates(botUrl string, offset int) ([]Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + strconv.Itoa(offset))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}

	return restResponse.Result, nil
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

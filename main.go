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
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	strip "github.com/grokify/html-strip-tags-go"
	yaml "gopkg.in/yaml.v3"
)

var (
	query        string
	m            BaseMsg
	mHead        BaseMsg2
	botMessage   BotMessage
	startID      int
	offset       int
	logged       int
	timeSynh     time.Time
	userex       userExists
	emailProfile int
	chatPatch    int
	synhSend     = false
	tsynh        time.Time
	cycle        time.Duration
	db           *sql.DB
)

func main() {
	fmt.Println(`Set timeot cycle (seconds):`)
	fmt.Scanf("%d", &cycle)

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
	k := 1
	for {
		k++
		now1 := time.Now()
		fmt.Println(now1)
		//ping tgbotapi
		_, err = getUpdates(botUrl, offset)
		if err != nil {
			errapi := true
			for errapi {
				timerApi := time.NewTimer(time.Second * 10)
				<-timerApi.C
				_, err = getUpdates(botUrl, offset)
				if err == nil {
					errapi = false
				}
				fmt.Println("Error ping tgbotapi")
				fmt.Scanf(" ")
			}
			fmt.Println("reconnected to tgbotapi")
		} else {
			fmt.Println("tgbotapi connection ok")
		}

		//ping db
		err = db.Ping()
		if err != nil {
			_ = sendMessage(botUrl, "Error ping db", 261609763)
			errdb := true
			for errdb {
				timerdb := time.NewTimer(time.Second * 10)
				<-timerdb.C
				err = db.Ping()
				if err == nil {
					errdb = false
				}
				fmt.Println("Error ping db")

			}
			fmt.Println("reconnected to db")
		} else {
			fmt.Println("db connection ok")
		}

		query = "SELECT top 1 id FROM [zaprosi].[dbo].[perepiska] order by id desc"

		rows, err := db.Query(query)
		if err != nil {
			log.Panic(err)
			fmt.Scanf(" ")

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

		timer := time.NewTimer(time.Second * cycle)

		<-timer.C

		updates, err := getUpdates(botUrl, offset)

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
				zapr := fmt.Sprintf(`INSERT INTO [dbo].[tgbot] ([email],[chatid],[pin],[loginned],active,[hour_begin]
					,[hour_end]
					,[minute_begin]
					,[minute_end]) VALUES (null,%d,null,%d,1,9,18,45,45)`, update.Message.Chat.ChatId, 0)
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

			if strings.HasPrefix(update.Message.Text, "Оповещение") && (update.Message.Chat.ChatId == 261609763 || update.Message.Chat.ChatId == 319080225) {
				msg := strings.ReplaceAll(update.Message.Text, "Оповещение", "")
				query = `SELECT
				[chatid]
			 FROM [zaprosi].[dbo].[tgbot]  where active =1`
				rows, err = db.Query(query)
				if err != nil {
					log.Fatal(err)
					fmt.Scanf("")
				}
				if rows != nil {
					for rows.Next() {
						//fmt.Println("nexter")

						if err := rows.Scan(&chatPatch); err != nil {
							log.Panic(err)
							fmt.Scanf("")
						}
						err = sendMessage(botUrl, msg, chatPatch)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}

					}
					err = sendMessage(botUrl, "Оповещение отправлено", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
				} else {
					logged = 0
				}

			}

			if logged == 1 {
				switch {
				case update.Message.Text == "heyo":
					err = sendMessage(botUrl, "darova", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
				case update.Message.Reply_to_message.Text != "":
					res1 := strings.Index(update.Message.Reply_to_message.Text, "Номер запроса:  ")
					fmt.Println("Result 1: ", res1)
					npr := ""
					k := res1 + 12 + 28
					for i := res1 + 28; i < k; i++ {

						npr = npr + string(update.Message.Reply_to_message.Text[i])
					}
					zapr := fmt.Sprintf(`with collective AS (

							SELECT
							
							STRING_AGG(profile.[shnam], '|w|') AS group_komu
							
					
							
							,n_project
							
							FROM mp_zapr
							
							inner join profile ON profile.shnam = mp_zapr.[shnam]
							
							GROUP BY  n_project
							
							),
							
							promej as (
							
							select top 1 [tgbot].[email],'%s' as vrem_proj ,shnam
							
							from [zaprosi].[dbo].[tgbot]
							
							INNER JOIN profile on profile.[email] = [tgbot].[email]
							
							where chatid=%d
							
							)
							
							INSERT INTO [dbo].[perepiska]
							
							([datesend]
							
							,[komu]
							
							,[kto]
							
							,[nproject]
							
							,[message]
							
							,[naprav])
							
							SELECT GETDATE()
							
							, REPLACE(REPLACE(group_komu,promej.email + '|w|',''), promej.shnam + '|w|', '')
							
							, promej.[email]
							
							,'%s'
							
							,N'%s'
							
							,'main'
							
							FROM collective
							
							LEFT JOIN promej ON vrem_proj = n_project
							
							WHERE n_project = '%s'`, npr, update.Message.Chat.ChatId, npr, update.Message.Text, npr)

					_, err = db.Exec(zapr)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				case (strings.HasPrefix(update.Message.Text, "Начало") || strings.HasPrefix(update.Message.Text, "начало")):
					re := regexp.MustCompile("[0-9]+")
					a := re.FindAllString(update.Message.Text, -1)
					t1, _ := strconv.Atoi(a[0])
					t2, _ := strconv.Atoi(a[1])
					if t1 >= 0 && t1 <= 23 && t2 <= 59 && t2 >= 0 {
						zapr := fmt.Sprintf("update [dbo].[tgbot] set [hour_begin] = %d,[minute_begin] = %d  where chatid=%d", t1, t2, update.Message.Chat.ChatId)
						_, err = db.Exec(zapr)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
						msg := fmt.Sprintf("Время начала работы установлено на %d:%d", t1, t2)
						err = sendMessage(botUrl, msg, update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
					} else {

						err = sendMessage(botUrl, "Неправильный формат даты. Дата конца работы не изменена.", update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
					}

				case (strings.HasPrefix(update.Message.Text, "Конец") || strings.HasPrefix(update.Message.Text, "конец")):
					re := regexp.MustCompile("[0-9]+")
					a := re.FindAllString(update.Message.Text, -1)
					t1, _ := strconv.Atoi(a[0])
					t2, _ := strconv.Atoi(a[1])
					if t1 >= 0 && t1 <= 23 && t2 <= 59 && t2 >= 0 {
						zapr := fmt.Sprintf("update [dbo].[tgbot] set [hour_end] = %d,[minute_end] = %d  where chatid=%d", t1, t2, update.Message.Chat.ChatId)
						_, err = db.Exec(zapr)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
						msg := fmt.Sprintf("Время конца работы установлено на %d:%d", t1, t2)
						err = sendMessage(botUrl, msg, update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
					} else {

						err = sendMessage(botUrl, "Неправильный формат даты. Дата конца работы не изменена.", update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
					}

				case (strings.HasPrefix(update.Message.Text, "Включить") || strings.HasPrefix(update.Message.Text, "включить")):
					zapr := fmt.Sprintf("update [dbo].[tgbot] set active = 1  where chatid=%d", update.Message.Chat.ChatId)
					_, err = db.Exec(zapr)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
					err = sendMessage(botUrl, "Оповещения включены. Приятного пользования!", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				case (strings.HasPrefix(update.Message.Text, "Выключить") || strings.HasPrefix(update.Message.Text, "выключить")):
					zapr := fmt.Sprintf("update [dbo].[tgbot] set active = 0  where chatid=%d", update.Message.Chat.ChatId)
					_, err = db.Exec(zapr)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
					err = sendMessage(botUrl, "Оповещения выключены. \nВозвращайтесь! :)", update.Message.Chat.ChatId)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}

				default:
					err = sendMessage(botUrl, `Доступные на данный момент команды:
					Выключить - отключить оповещения
					Включить - включить оповещения
					Начало - Установить время начала отправки оповещения. По умолчанию 9.45. Пример команды: "Начало 8.30"
					Конец - Установить время начала отправки оповещения. По умолчанию 18.45. Пример команды: "Конец  21.15"

					Если у тебя возникла какая-то проблема со мной, или же хочется предлжить что-то для доработки - обращайся к soreshnikov@dtln.ru/ https://t.me/soreshnikov
					`, update.Message.Chat.ChatId)
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

					Доступные на данный момент команды:
					Выключить - отключить оповещения
					Включить - включить оповещения
					Начало - Установить время начала отправки оповещения. По умолчанию 9.45. Пример команды: "Начало 8.30"
					Конец - Установить время начала отправки оповещения. По умолчанию 18.45. Пример команды: "Конец  21.15"

					Если у тебя возникла какая-то проблема со мной, или же хочется предлжить что-то для доработки - обращайся к soreshnikov@dtln.ru/ https://t.me/soreshnikov`, update.Message.Chat.ChatId)
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
					query = fmt.Sprintf(`SELECT CASE WHEN EXISTS (SELECT TOP 1 [id]
						from [zaprosi].[dbo].[profile] where  [zaprosi].[dbo].[profile].email = (select top 1 email
						FROM [zaprosi].[dbo].[tgbot] where chatid=%d)) THEN '1'ELSE '0' END`, update.Message.Chat.ChatId)
					fmt.Println(query)

					rows, err := db.Query(query)
					if err != nil {
						log.Fatal(err)
						fmt.Scanf("")
					}
					if rows != nil {
						for rows.Next() {
							//fmt.Println("nexter")

							if err := rows.Scan(&emailProfile); err != nil {
								log.Panic(err)
								fmt.Scanf(" ")
							}
						}
					}

					if emailProfile == 1 {
						err = sendMessage(botUrl, "Отлично, осталось только ввести пин, который я прислал тебе на почту.", update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}
						emailText := fmt.Sprintf(`<td id=topText><b>Получение PIN</b></td></tr>
						</table></td></tr><tr><td><h3>Здравствуйте!</h3><p>Для вашей учетной записи был сгенерирован PIN: <div id=samtext style=padding:15px;font-size:40px;text-align:center;letter-spacing:5px;font-weight:500;>%d</div></p><br><br><br><br><br><br><br></td></tr>
						</table>
				 </BODY></HTML>`, pin)
						//err := sendEmail(update.Message.Text, "", "Telegram PIN", emailText)
						zapr := fmt.Sprintf(`INSERT INTO [dbo].[sendemail]
							([komu]
							,[kopiya]
							,[tema]
							,[telopisma])
						VALUES
							('%s;',null,N'%s',N'%s')`, update.Message.Text, "Telegram PIN", emailText)
						fmt.Println(zapr)

						_, err = db.Exec(zapr)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf(" ")
						}

					} else {
						err = sendMessage(botUrl, "Такой email не зарегистрирован на споке. Попробуйте проверить корректность введения и отправить его заново.", update.Message.Chat.ChatId)
						if err != nil {
							log.Fatal(err)
							fmt.Scanf("")
						}

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
					err = sendMessage(botUrl, `Вы не зарегистрированы. Введите свой корпоративный email. 
					Если у тебя возникла какая-то проблема со мной - обращайся к soreshnikov@dtln.ru/ https://t.me/soreshnikov`, update.Message.Chat.ChatId)
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

		query = fmt.Sprintf(`With table_whom

		AS
		
		(SELECT
		
		[komyt],
		
		MAX([datewhen]) as [datewhen],
		
		[nid],
		
		[mainid],
		
		[nproject],
		
		[valid],
		
		MAX([Zakazchik]) AS [Zakazchik],
		
		MIN([NameProject]) as namePr
		
		FROM [zaprosi].[dbo].[komy_table]
		
		INNER JOIN [zaprosi].[dbo].[zapr] on [zapr].[main_id] = [komy_table].mainid
		
		GROUP BY [komyt],
		
		[nid],
		
		[mainid],
		
		[nproject],
		
		[valid]
		
		),
		
		usery AS
		
		(
		
		SELECT
		
		[shnam],
		
		[name] + ' ' + [secnam] AS nameUser ,
		
		email
		
		FROM
		
		[zaprosi].[dbo].[profile]
		
		),
		
		chat_id AS (
		
		SELECT chatid, email from [zaprosi].[dbo].[tgbot]
		
		where loginned = 1 and active = 1
		
		and ((DATEPART(HOUR, GETDATE())>=hour_begin and DATEPART(HOUR, GETDATE())<=hour_end)
		
		or (DATEPART(HOUR, GETDATE())=hour_begin and DATEPART(MINUTE, GETDATE())>=MINUTE_begin))
		
		or (DATEPART(HOUR, GETDATE())=hour_end and DATEPART(MINUTE, GETDATE())<=MINUTE_end)
		
		),
		
		zapros_npr AS (
		
		SELECT  main_id
		
		FROM zapr
		
		GROUP BY main_id
		
		)
		
		SELECT  [perepiska].[id],
		
		pr_file.[shnam]
		
		, usery.nameUser
		
		,[komyt]
		
		,(pr_file.[name] + ' ' + pr_file.[secnam]) as kto
		
		,' ' + [perepiska].[nproject] as nproject
		
		,[message]
		
		, ISNULL([naprav],'main') AS naprav
		
		,[mainid]
		
		,[Zakazchik] as zakaz 
		
		,chatid as chat
		
		,namePr AS [NameProject]
		
		FROM [zaprosi].[dbo].[perepiska]
		
		INNER JOIN table_whom ON [nid] = [id]
		
		INNER JOIN usery ON usery.[email] = [komyt]
		
		INNER JOIN chat_id ON komyt = chat_id.email
		
		INNER JOIN profile as pr_file ON pr_file.shnam = [kto]
		
		WHERE
		
		--and DATEPART(HOUR, GETDATE())<=hour_end and DATEPART(MINUTE, GETDATE())<=MINUTE_end)
		
		[perepiska].id> %d
		
		order by [perepiska].id `, startID)
		//fmt.Println(query)

		rows, err = db.Query(query)
		if err != nil {
			log.Fatal(err)
			fmt.Scanf("")
		}

		for rows.Next() {

			if err := rows.Scan(&m.id, &m.login, &m.fullNameWhom, &m.whom, &m.who, &m.nproject, &m.message, &m.naprav, &m.mainid, &m.zakazchik, &m.chat, &m.description); err != nil {
				log.Panic(err)
				fmt.Scanf(" ")
			}

			fmt.Println(m.message)

			msg := fmt.Sprintf("%s\n %s\nНомер запроса: %s \nОтправитель: %s \n\nСообщение: \n %s", m.zakazchik, m.description, m.nproject, m.who, m.message)
			msg = cleartags(msg)
			fmt.Println(msg)
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
			FROM [zaprosi].[dbo].[zapr] where napr='ork' and mpp is null ) and id>%d order by id`, startID2)
		fmt.Println(startID2)

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

			msg := fmt.Sprintf("Неназначенный\nНомер запроса: %s\n Отправитель: %s \n Сообщение: \n %s", mHead.nproject, mHead.who, mHead.message)
			msg = cleartags(msg)
			fmt.Println(msg)
			err := sendMessage(botUrl, msg, 261609763)
			if err != nil {

				log.Panic(err)

			}
			startID2 = mHead.id
		}

		query = `SELECT [obn] FROM [zaprosi].[dbo].[obnov] where id = 1`
		//fmt.Println(query)

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
		}

		if synhSend && tsynh != timeSynh {

			synhSend = false

		}
		now := time.Now()
		countm := 2
		counth := 3
		timeBottom := now.Add(time.Duration(-countm) * time.Minute)
		timeBottom = timeBottom.Add(time.Duration(counth) * time.Hour)

		if timeSynh.Before(timeBottom) {
			if !synhSend {
				err := sendMessage(botUrl, "Ошибка синхронизации", 261609763)
				if err != nil {

					log.Panic(err)

				}
				err = sendMessage(botUrl, "Ошибка синхронизации", 319080225)
				if err != nil {

					log.Panic(err)

				}
				synhSend = true
				tsynh = timeSynh
			}

		}
		fmt.Println("")
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

func sendEmail(komu string, kopiya string, tema string, telopisma string) error {
	textfile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return err
	}
	config := Config{}
	err = yaml.Unmarshal([]byte(textfile), &config)
	if err != nil {
		return err
	}
	conString := fmt.Sprintf("user id=%s;password=%s;port=%d;database=%s", config.Db.User, config.Db.Password, config.Db.Port, config.Db.Database)
	db1, err := sql.Open("mssql", conString)
	if err != nil {
		return err
	}
	zapr := fmt.Sprintf(`INSERT INTO [dbo].[sendemail]
		([komu]
		,[kopiya]
		,[tema]
		,[telopisma])
 	 VALUES
		(%s,%s,%s,%s) GO`, komu, kopiya, tema, telopisma)

	_, err = db1.Exec(zapr)
	if err != nil {
		return err
	}

	/*
		SELECT TOP (1000) [id]
		,[komu]
		,[kopiya]
		,[tema]
		,[telopisma]
		,[datesend]
		,[sendWhen]
		,[category]
		,[lastMess]
		,[SendingInf]
		,[numFromperep]
		,[fileInput]
		FROM [zaprosi].[dbo].[sendemail]
	*/
	return nil
}

func cleartags(text string) (text1 string) {

	text = strings.Replace(text, "</tr>", "</tr>\n", -1)
	text = strings.Replace(text, "</td>", "</td> ", -1)
	text = strings.Replace(text, "&nbsp;", " ", -1)
	text = strings.Replace(text, "&gt;", " ", -1)
	text = strings.Replace(text, "&#34;", " ", -1)
	text1 = strip.StripTags(text)
	return

}

func Difference(a, b []int) (diff []int) {
	m := make(map[int]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

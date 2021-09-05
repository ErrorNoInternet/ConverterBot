package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/cpu"
)

var (
	startTime   int64
	cpuUsage    float64
	prefix      string = ")"
	embedColor  int    = 5592405
	currentUser *discordgo.User
	guildList   []*discordgo.GuildCreate
)

var suggestionManagers = []string{
	"531392146767347712",
}

var commands = map[string]string{
	"help":        "Display a list of available commands",
	"ping":        "Display the bot's current API latency",
	"status":      "Display the bot's current statistics",
	"invite":      "Display a link to invite the bot",
	"conversions": "Display a list of available conversions",
	"convert":     "Convert an amount to another unit",
	"currency":    "Convert an amount to another currency",
	"suggest":     "Send a suggestion to the bot creators",
}

var abbreviations = map[string]string{
	"m":    "meter",
	"cm":   "centimeter",
	"km":   "kilometer",
	"mi":   "mile",
	"bit":  "bit",
	"byte": "byte",
	"kb":   "kilobyte",
	"mb":   "megabyte",
	"gb":   "gigabyte",
	"tb":   "terabyte",
	"pb":   "petabyte",
	"ft":   "feet",
	"in":   "inch",
	"lb":   "pound",
	"kg":   "kilogram",
	"g":    "gram",
	"oz":   "ounce",
	"h/s":  "hash",
	"kh/s": "kilohash",
	"mh/s": "megahash",
	"ms":   "millisecond",
	"s":    "second",
	"min":  "minute",
	"hr":   "hour",
	"d":    "day",
	"w":    "week",
	"y":    "year",
	"au":   "astronomical unit",
}

var conversions = []ConversionData{
	ConversionData{Input: "m", Output: "cm", Type: "multiply", Number: 100.0},
	ConversionData{Input: "cm", Output: "mm", Type: "multiply", Number: 10.0},
	ConversionData{Input: "mi", Output: "km", Type: "multiply", Number: 1.609344},
	ConversionData{Input: "ft", Output: "cm", Type: "multiply", Number: 30.48},
	ConversionData{Input: "m", Output: "ft", Type: "multiply", Number: 3.2808399},
	ConversionData{Input: "km", Output: "ft", Type: "multiply", Number: 3280.8399},
	ConversionData{Input: "au", Output: "km", Type: "multiply", Number: 149597871},
	ConversionData{Input: "kg", Output: "lb", Type: "multiply", Number: 2.20462262},
	ConversionData{Input: "kg", Output: "oz", Type: "multiply", Number: 35.2739619},
	ConversionData{Input: "kg", Output: "g", Type: "multiply", Number: 1000},
	ConversionData{Input: "oz", Output: "g", Type: "multiply", Number: 28.3495231},
	ConversionData{Input: "ft", Output: "in", Type: "multiply", Number: 12},
	ConversionData{Input: "in", Output: "cm", Type: "multiply", Number: 2.54},
	ConversionData{Input: "in", Output: "m", Type: "multiply", Number: 0.0254},

	ConversionData{Input: "ms", Output: "s", Type: "divide", Number: 1000},
	ConversionData{Input: "ms", Output: "min", Type: "divide", Number: 60000},
	ConversionData{Input: "ms", Output: "hr", Type: "divide", Number: 3600000},
	ConversionData{Input: "ms", Output: "d", Type: "divide", Number: 86400000},
	ConversionData{Input: "ms", Output: "w", Type: "divide", Number: 604800000},
	ConversionData{Input: "ms", Output: "y", Type: "divide", Number: 31556952000},
	ConversionData{Input: "s", Output: "min", Type: "divide", Number: 60},
	ConversionData{Input: "s", Output: "hr", Type: "divide", Number: 3600},
	ConversionData{Input: "s", Output: "d", Type: "divide", Number: 86400},
	ConversionData{Input: "s", Output: "w", Type: "divide", Number: 604800},
	ConversionData{Input: "s", Output: "y", Type: "divide", Number: 31556926},
	ConversionData{Input: "min", Output: "hr", Type: "divide", Number: 60},
	ConversionData{Input: "min", Output: "d", Type: "divide", Number: 1440},
	ConversionData{Input: "min", Output: "w", Type: "divide", Number: 10080},
	ConversionData{Input: "min", Output: "y", Type: "divide", Number: 525948.766},
	ConversionData{Input: "hr", Output: "d", Type: "divide", Number: 24},
	ConversionData{Input: "hr", Output: "w", Type: "divide", Number: 168},
	ConversionData{Input: "hr", Output: "y", Type: "divide", Number: 8765.81277},
	ConversionData{Input: "d", Output: "w", Type: "divide", Number: 7},
	ConversionData{Input: "d", Output: "y", Type: "divide", Number: 365},
	ConversionData{Input: "w", Output: "y", Type: "divide", Number: 52.177457},

	ConversionData{Input: "h/s", Output: "kh/s", Type: "divide", Number: 1000},
	ConversionData{Input: "kh/s", Output: "mh/s", Type: "divide", Number: 1000},
	ConversionData{Input: "mh/s", Output: "h/s", Type: "multiply", Number: 1000000},
	ConversionData{Input: "bit", Output: "byte", Type: "divide", Number: 8},
	ConversionData{Input: "bit", Output: "kb", Type: "divide", Number: 8000},
	ConversionData{Input: "bit", Output: "mb", Type: "divide", Number: 8000000},
	ConversionData{Input: "bit", Output: "gb", Type: "divide", Number: 8000000000},
	ConversionData{Input: "bit", Output: "tb", Type: "divide", Number: 8000000000000},
	ConversionData{Input: "bit", Output: "pb", Type: "divide", Number: 8000000000000000},
	ConversionData{Input: "byte", Output: "kb", Type: "divide", Number: 1000},
	ConversionData{Input: "byte", Output: "mb", Type: "divide", Number: 1000000},
	ConversionData{Input: "byte", Output: "gb", Type: "divide", Number: 1000000000},
	ConversionData{Input: "byte", Output: "tb", Type: "divide", Number: 1000000000000},
	ConversionData{Input: "byte", Output: "pb", Type: "divide", Number: 1000000000000000},
	ConversionData{Input: "kb", Output: "mb", Type: "divide", Number: 1000},
	ConversionData{Input: "kb", Output: "gb", Type: "divide", Number: 1000000},
	ConversionData{Input: "kb", Output: "tb", Type: "divide", Number: 1000000000},
	ConversionData{Input: "kb", Output: "pb", Type: "divide", Number: 1000000000000},
	ConversionData{Input: "mb", Output: "gb", Type: "divide", Number: 1000},
	ConversionData{Input: "mb", Output: "tb", Type: "divide", Number: 1000000},
	ConversionData{Input: "mb", Output: "pb", Type: "divide", Number: 1000000000},
	ConversionData{Input: "gb", Output: "tb", Type: "divide", Number: 1000},
	ConversionData{Input: "gb", Output: "pb", Type: "divide", Number: 1000000},
	ConversionData{Input: "tb", Output: "pb", Type: "divide", Number: 1000},
}

type ConversionData struct {
	Input  string
	Output string
	Type   string
	Number float64
}

func main() {
	fmt.Println("Initializing bot...")
	startTime = time.Now().Unix()

	go updateCPU()
	for _, conversion := range conversions {
		conversionType := conversion.Type
		reversedType := ""
		if conversionType == "multiply" {
			reversedType = "divide"
		} else if conversionType == "divide" {
			reversedType = "multiply"
		}
		if reversedType != "" {
			conversions = append(conversions, ConversionData{conversion.Output, conversion.Input, reversedType, conversion.Number})
		}
	}
	rand.Seed(time.Now().UnixNano())
	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Println("Unable to load TOKEN variable")
		return
	}
	client, errorObject := discordgo.New("Bot " + token)
	if errorObject != nil {
		fmt.Println("Unable to login: " + errorObject.Error())
		return
	}
	currentUser, errorObject = client.User("@me")
	if errorObject != nil {
		fmt.Println("Unable to fetch current user: " + errorObject.Error())
		return
	}
	client.AddHandler(readyEvent)
	client.AddHandler(guildJoinEvent)
	client.AddHandler(messageCreateEvent)
	client.Identify.Intents = discordgo.IntentsAll
	errorObject = client.Open()
	if errorObject != nil {
		fmt.Println("Unable to login: " + errorObject.Error())
		return
	}
	<-make(chan struct{})
}

func updateCPU() {
	for {
		percent, _ := cpu.Percent(time.Second, true)
		cpuUsage = percent[0]
		time.Sleep(10)
	}
}

func convert(input, output string, amount float64) float64 {
	for _, conversion := range conversions {
		if input == strings.ToLower(conversion.Input) && output == strings.ToLower(conversion.Output) {
			if conversion.Type == "multiply" {
				return amount * conversion.Number
			}
			if conversion.Type == "divide" {
				return amount / conversion.Number
			}
		}
	}
	return 0.0
}

func readyEvent(session *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Successfully logged in as %v#%v\n", currentUser.Username, currentUser.Discriminator)
}

func guildJoinEvent(session *discordgo.Session, guild *discordgo.GuildCreate) {
	exists := false
	for _, joinedGuild := range guildList {
		if joinedGuild.ID == guild.ID {
			exists = true
		}
	}
	if !exists {
		guildList = append(guildList, guild)
	}
}

func humanizeNumber(number float64) string {
	stringNumber := fmt.Sprintf("%f", number)
	parts := strings.Split(stringNumber, ".")
	wholeNumber, _ := strconv.ParseInt(parts[0], 10, 0)
	return humanize.Comma(wholeNumber) + "." + parts[1]
}

func messageCreateEvent(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}

	if strings.Contains(message.Content, "<@") && strings.Contains(message.Content, currentUser.ID+">") {
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("My prefix here is `%v`", prefix))
	}

	if strings.HasPrefix(message.Content, prefix+"invite") {
		inviteLink := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%v&permissions=8&scope=bot", currentUser.ID)
		embed := &discordgo.MessageEmbed{
			Title:       "Invite Link",
			Description: fmt.Sprintf("You can invite me to your server using [this link](%v)", inviteLink),
			Color:       embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}

	if strings.HasPrefix(message.Content, prefix+"vote") {
		embed := &discordgo.MessageEmbed{
			Title:       "Vote Link",
			Description: fmt.Sprintf("You can vote for ConverterBot using [this link](%v)", "https://top.gg/bot/877069460186492978"),
			Color:       embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}

	if strings.HasPrefix(message.Content, prefix+"ping") {
		embed := &discordgo.MessageEmbed{
			Title:       "Pong :ping_pong:",
			Description: fmt.Sprintf("Latency: **%v ms**", session.HeartbeatLatency().Milliseconds()),
			Color:       embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}

	if strings.HasPrefix(message.Content, prefix+"status") {
		var memoryData runtime.MemStats
		runtime.ReadMemStats(&memoryData)
		ramUsage := memoryData.Alloc / 1024 / 1024
		timeDuration := session.HeartbeatLatency()
		botLatency := timeDuration.Milliseconds()
		currentTime := time.Now().Unix()
		secondsTime := currentTime - startTime
		minutesTime := secondsTime / 60
		hoursTime := minutesTime / 60
		secondsTime = secondsTime % 60
		minutesTime = minutesTime % 60
		secondsOutput := strconv.Itoa(int(secondsTime)) + "s"
		minutesOutput := strconv.Itoa(int(minutesTime)) + "m"
		hoursOutput := strconv.Itoa(int(hoursTime)) + "hr"
		totalMembers := 0
		for _, server := range guildList {
			totalMembers += server.MemberCount
		}
		threadCount := runtime.NumGoroutine()
		embed := &discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Bot Latency",
					Value:  fmt.Sprintf("```%v ms```", botLatency),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "CPU Usage",
					Value:  fmt.Sprintf("```%.1f%%```", cpuUsage),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "RAM Usage",
					Value:  fmt.Sprintf("```%v MB```", ramUsage),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Goroutines",
					Value:  fmt.Sprintf("```%v```", threadCount),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Guild Count",
					Value:  fmt.Sprintf("```%v```", len(guildList)),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Member Count",
					Value:  fmt.Sprintf("```%v```", totalMembers),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Commands",
					Value:  fmt.Sprintf("```%v```", len(commands)),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "DiscordGo",
					Value:  fmt.Sprintf("```%v```", discordgo.VERSION),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Bot Uptime",
					Value:  fmt.Sprintf("```%v %v %v```", hoursOutput, minutesOutput, secondsOutput),
					Inline: true,
				},
			},
			Color: embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}

	if strings.HasPrefix(message.Content, prefix+"conversions") {
		description := ""
		for _, conversion := range abbreviations {
			description += conversion + ", "
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Available Conversions",
			Description: strings.TrimSuffix(description, ", "),
			Color:       embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}

	if strings.HasPrefix(message.Content, prefix+"suggest") {
		arguments := strings.Split(message.Content, " ")
		if len(arguments) > 1 {
			suggestion := ""
			for index, argument := range arguments {
				if index != 0 {
					suggestion += argument + " "
				}
			}
			sentUsers := []string{}
			for _, userID := range suggestionManagers {
				for _, guild := range guildList {
					for _, member := range guild.Members {
						if member.User.ID == userID {
							sent := false
							for _, sentUser := range sentUsers {
								if sentUser == member.User.ID {
									sent = true
								}
							}
							if sent {
								continue
							}
							sentUsers = append(sentUsers, member.User.ID)
							channel, _ := session.UserChannelCreate(member.User.ID)
							session.ChannelMessageSend(
								channel.ID,
								fmt.Sprintf(
									"**%v#%v (**`%v`**) has sent a suggestion:**\n%v",
									message.Author.Username,
									message.Author.Discriminator,
									message.Author.ID,
									suggestion,
								),
							)
						}
					}
				}
			}
			session.ChannelMessageSend(message.ChannelID, "Your suggestion has been successfully sent")
		} else {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("The syntax is `%vsuggest <suggestion>`", prefix))
		}
	}

	if strings.HasPrefix(message.Content, prefix+"convert") {
		arguments := strings.Split(message.Content, " ")
		if len(arguments) == 4 {
			rawNumber, errorObject := strconv.ParseFloat(arguments[1], 64)
			if errorObject != nil {
				session.ChannelMessageSend(message.ChannelID, "Please enter a valid amount")
				return
			}
			amount := float64(rawNumber)
			input := strings.ToLower(arguments[2])
			output := strings.ToLower(arguments[3])
			supported := false
			for _, conversion := range conversions {
				if strings.ToLower(conversion.Input) == input {
					if strings.ToLower(conversion.Output) == output {
						supported = true
					}
				}
			}
			if !supported {
				if len(input) > 2 {
					if strings.HasSuffix(input, "s") {
						input = strings.TrimSuffix(input, "s")
					}
				}
				if len(output) > 2 {
					if strings.HasSuffix(output, "s") {
						output = strings.TrimSuffix(output, "s")
					}
				}
				for abbreviation, name := range abbreviations {
					if strings.ToLower(name) == input {
						input = abbreviation
					}
					if strings.ToLower(name) == output {
						output = abbreviation
					}
				}
			}
			for _, conversion := range conversions {
				if strings.ToLower(conversion.Input) == input {
					if strings.ToLower(conversion.Output) == output {
						supported = true
					}
				}
			}
			if supported {
				inputAbbreviation := "unknown"
				outputAbbreviation := "unknown"
				abbreviation, ok := abbreviations[input]
				if ok {
					inputAbbreviation = abbreviation
				}
				abbreviation, ok = abbreviations[output]
				if ok {
					outputAbbreviation = abbreviation
				}
				description := fmt.Sprintf(
					"**%v %v** = **%v %v**\n\n**Unit abbreviations:**\n`%v` = `%v`, `%v` = `%v`",
					humanizeNumber(amount), input, humanizeNumber(convert(input, output, amount)), output, input, inputAbbreviation, output, outputAbbreviation,
				)
				embed := &discordgo.MessageEmbed{
					Title:       "Conversion",
					Description: description,
					Color:       embedColor,
				}
				session.ChannelMessageSendEmbed(message.ChannelID, embed)
			} else {
				session.ChannelMessageSend(message.ChannelID, "That input/output pair is not supported")
			}
		} else {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("The syntax is `%vconvert <amount> <input> <output>`", prefix))
		}
	}

	if strings.HasPrefix(message.Content, prefix+"currency") {
		arguments := strings.Split(message.Content, " ")
		if len(arguments) == 4 {
			rawNumber, errorObject := strconv.ParseFloat(arguments[1], 64)
			if errorObject != nil {
				session.ChannelMessageSend(message.ChannelID, "Please enter a valid amount")
				return
			}
			amount := float64(rawNumber)
			input := strings.ToLower(arguments[2])
			output := strings.ToLower(arguments[3])
			rawResponse, errorObject := http.Get(fmt.Sprintf("https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/%v/%v.json", input, output))
			if errorObject != nil {
				session.ChannelMessageSend(message.ChannelID, "Unable to convert currency")
				return
			}
			response, errorObject := ioutil.ReadAll(rawResponse.Body)
			if errorObject != nil {
				session.ChannelMessageSend(message.ChannelID, "Unable to convert currency")
				return
			}
			if strings.Contains(string(response), "size exceeded") {
				session.ChannelMessageSend(message.ChannelID, "That currency was not found")
				return
			}
			rawNumber, errorObject = strconv.ParseFloat(strings.Split(strings.Split(string(response), ": ")[2], "\n")[0], 64)
			if errorObject != nil {
				session.ChannelMessageSend(message.ChannelID, "Unable to convert currency")
				return
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Currency Convert",
				Description: fmt.Sprintf("**%v %v** = **%v %v**", humanizeNumber(amount), strings.ToUpper(input), humanizeNumber(amount*rawNumber), strings.ToUpper(output)),
				Color:       embedColor,
			}
			session.ChannelMessageSendEmbed(message.ChannelID, embed)
		} else {
			session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("The syntax is `%vcurrency <amount> <currency> <currency>`", prefix))
		}
	}

	if strings.HasPrefix(message.Content, prefix+"help") || message.Content == prefix+"help" {
		commandDescription := ""
		for command, description := range commands {
			commandDescription += fmt.Sprintf("`%v%v` - %v\n", prefix, command, description)
		}
		embed := &discordgo.MessageEmbed{
			Title:       "ConverterBot Commands",
			Description: commandDescription,
			Color:       embedColor,
		}
		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}
}

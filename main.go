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
	prefix      string = "/"
	embedColor  int    = 5592405
	session     *discordgo.Session
	currentUser *discordgo.User
	guildList   []*discordgo.GuildCreate
)

var suggestionManagers = []string{
	"531392146767347712",
}

var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Display the bot's current API latency",
	},
	{
		Name:        "status",
		Description: "Display the bot's current statistics",
	},
	{
		Name:        "conversions",
		Description: "Display a list of available conversions",
	},
	{
		Name:        "suggest",
		Description: "Send a suggestion to ConverterBot's creators",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "suggestion",
				Description: "The suggestion you want to send",
				Required:    true,
			},
		},
	},
	{
		Name:        "invite",
		Description: "Display a link to invite the bot",
	},
	{
		Name:        "vote",
		Description: "Display a link to upvote ConverterBot",
	},
	{
		Name:        "convert",
		Description: "Convert different amounts to different units",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "amount",
				Description: "The amount you have (for the input unit)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "input-unit",
				Description: "The unit that the amount will be converted from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "output-unit",
				Description: "The unit that the amount will be converted to",
				Required:    true,
			},
		},
	},
	{
		Name:        "currency",
		Description: "Convert different amounts to different currencies",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "amount",
				Description: "The amount you have (for the input currency)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "input-currency",
				Description: "The currency that the amount will be converted from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "output-currency",
				Description: "The currency that the amount will be converted to",
				Required:    true,
			},
		},
	},
}

var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"ping": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		embed := &discordgo.MessageEmbed{
			Title:       "Pong :ping_pong:",
			Description: fmt.Sprintf("Latency: **%v ms**", session.HeartbeatLatency().Milliseconds()),
			Color:       embedColor,
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"status": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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
					Value:  fmt.Sprintf("```%v```", len(slashCommands)),
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
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"currency": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		errorText := "Unable to convert currencies"
		amount, errorObject := strconv.ParseFloat(strings.Replace(interaction.ApplicationCommandData().Options[0].StringValue(), " ", "", -1), 10)
		if errorObject != nil {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Please enter a valid amount", Flags: 64},
			})
			return
		}
		input := strings.Replace(strings.ToLower(interaction.ApplicationCommandData().Options[1].StringValue()), " ", "", -1)
		output := strings.Replace(strings.ToLower(interaction.ApplicationCommandData().Options[2].StringValue()), " ", "", -1)
		rawResponse, errorObject := http.Get(fmt.Sprintf("https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/%v/%v.json", input, output))
		if errorObject != nil {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: errorText, Flags: 64},
			})
			return
		}
		response, errorObject := ioutil.ReadAll(rawResponse.Body)
		if errorObject != nil {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: errorText, Flags: 64},
			})
			return
		}
		if strings.Contains(string(response), "size exceeded") {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "That currency was not found", Flags: 64},
			})
			return
		}
		rawNumber, errorObject := strconv.ParseFloat(strings.Split(strings.Split(string(response), ": ")[2], "\n")[0], 10)
		if errorObject != nil {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: errorText, Flags: 64},
			})
			return
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Currency Conversion",
			Description: fmt.Sprintf("**%v %v** = **%v %v**", humanizeNumber(amount), strings.ToUpper(input), humanizeNumber(amount*rawNumber), strings.ToUpper(output)),
			Color:       embedColor,
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"invite": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		inviteLink := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%v&permissions=8&scope=applications.commands%%20bot", currentUser.ID)
		embed := &discordgo.MessageEmbed{
			Title:       "Invite Link",
			Description: fmt.Sprintf("You can invite me to your server using [this link](%v)", inviteLink),
			Color:       embedColor,
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"vote": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		embed := &discordgo.MessageEmbed{
			Title:       "Vote Link",
			Description: fmt.Sprintf("You can vote for ConverterBot using [this link](%v)", "https://top.gg/bot/877069460186492978"),
			Color:       embedColor,
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"conversions": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		description := ""
		for _, conversion := range abbreviations {
			description += conversion + ", "
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Available Conversions",
			Description: strings.TrimSuffix(description, ", "),
			Color:       embedColor,
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
		})
	},
	"suggest": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		suggestion := interaction.ApplicationCommandData().Options[0].Value
		sentUsers := []string{}
		for _, userID := range suggestionManagers {
			for _, guild := range guildList {
				member, errorObject := session.State.Member(guild.ID, userID)
				if errorObject == nil {
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
					session.ChannelMessageSend(channel.ID, fmt.Sprintf(
						"**%v#%v (**`%v`**) has sent a new suggestion:**\n%v",
						interaction.Member.User.Username, interaction.Member.User.Discriminator, interaction.Member.User.ID, suggestion,
					))
				}
			}
		}
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "Your suggestion has been successfully sent"},
		})
	},
	"convert": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		rawNumber, errorObject := strconv.ParseFloat(strings.Replace(interaction.ApplicationCommandData().Options[0].StringValue(), " ", "", -1), 64)
		if errorObject != nil {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Please enter a valid amount", Flags: 64},
			})
			return
		}
		amount := float64(rawNumber)
		input := strings.Replace(strings.ToLower(interaction.ApplicationCommandData().Options[1].StringValue()), " ", "", -1)
		output := strings.Replace(strings.ToLower(interaction.ApplicationCommandData().Options[2].StringValue()), " ", "", -1)
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
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
			})
		} else {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "That input/output pair is not supported"},
			})
		}
	},
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
	session, errorObject := discordgo.New("Bot " + token)
	if errorObject != nil {
		fmt.Println("Unable to login: " + errorObject.Error())
		return
	}
	currentUser, errorObject = session.User("@me")
	if errorObject != nil {
		fmt.Println("Unable to fetch current user: " + errorObject.Error())
		return
	}
	session.AddHandler(readyEvent)
	session.AddHandler(guildJoinEvent)
	session.AddHandler(messageCreateEvent)
	session.Identify.Intents = discordgo.IntentsAll
	errorObject = session.Open()
	if errorObject != nil {
		fmt.Println("Unable to login: " + errorObject.Error())
		return
	}
	for _, value := range slashCommands {
		_, errorObject := session.ApplicationCommandCreate(session.State.User.ID, "", value)
		if errorObject != nil {
			fmt.Println("Error with slash command: " + errorObject.Error())
		}
	}
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, exists := commandHandlers[interaction.ApplicationCommandData().Name]; exists {
			handler(session, interaction)
		}
	})
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
		session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("My prefix here is `%v` (slash commands)", prefix))
	}
}

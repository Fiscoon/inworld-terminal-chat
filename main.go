package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
)

const (
	STATUS_URL                       = "http://localhost:3000/status"
	STATUS_CHECK_TIME                = 5 * time.Second
	CHARACTER_MESSAGE_POLL_FREQUENCY = 1000 * time.Millisecond

	DEFAULT_UID          = "-1"
	DEFAULT_SCENE_ID     = "workspaces/gensoukyou/scenes/test"
	DEFAULT_CHARACTER_ID = "-1"
	DEFAULT_PLAYER_NAME  = "Player"
	DEFAULT_SERVER_ID    = "1234"

	TYPING_PROMPT                     = "> "
	TYPING_DELAY                      = 50 * time.Millisecond
	SEND_MESSAGE_DELAY                = 2 * time.Second
	CHARACTER_TEXT_WIDTH              = 30
	SPEECH_SPRITE_SIZE                = 2
	SPEECH_SPRITE_CHANGE_TIME         = 200 * time.Millisecond
	CHARACTER_MESSAGE_SEPARATION_LINE = "-----"
	CHARACTER_MESSAGE_LINES_COUNT     = 10
)

type SpeechSprites [][]string

type CreateSessionParams struct {
	Uid         string `json:"uid"`
	SceneId     string `json:"sceneId"`
	CharacterId string `json:"characterId"`
	PlayerName  string `json:"playerName"`
	ServerId    string `json:"serverId"`
}

type CharacterMessage struct {
	Type      string `json:"type"`
	SessionId string `json:"sessionId"`
	Uid       string `json:"uid"`
	ServerId  string `json:"serverId"`
	Final     bool   `json:"final"`
	Text      string `json:"text"`
}

func main() {
	initializeLogFile()
	closePreviousChatSessions()
	go checkInworldStatus()
	sessionId := openChatSession()

	chatLoop(sessionId)
}

func initializeLogFile() {
	logFile, err := os.Create("log.txt")
	if err != nil {
		log.Println(err)
	}
	_ = logFile
	log.SetOutput(logFile)
}
func logMessage(logMessage ...interface{}) {
	if logMessage == nil {
		return
	}
	log.Println(logMessage...)
}

func chatLoop(sessionId string) {
	messageUrl := "http://localhost:3000/session/" + sessionId + "/message"
	messagesChan := make(chan CharacterMessage)
	lineChan := make(chan string)

	screen, _ := tcell.NewScreen()
	screen.Init()

	speechSprites := loadSpeechSprites()

	go pollMessages(messagesChan)
	go produceLines(messagesChan, lineChan)
	go handleUserInput(screen, messageUrl, lineChan)
	go renderResponses(screen, speechSprites, lineChan)

	select {}
}

func checkInworldStatus() {
	for {
		time.Sleep(STATUS_CHECK_TIME)
		_, err := http.Get(STATUS_URL)
		if err != nil {
			log.Fatalln("Couldn't communicate with the server. Check connectivity.")
		}
	}
}

func openChatSession() string {
	url := "http://localhost:3000/session/open"

	toSend := CreateSessionParams{
		Uid:         DEFAULT_UID,
		SceneId:     DEFAULT_SCENE_ID,
		CharacterId: DEFAULT_CHARACTER_ID,
		PlayerName:  DEFAULT_PLAYER_NAME,
		ServerId:    DEFAULT_SERVER_ID,
	}

	payload, err := json.Marshal(toSend)
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var jsonData map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Fatalln(err)
	}

	var sessionID string
	err = json.Unmarshal(jsonData["sessionId"], &sessionID)
	if err != nil {
		log.Fatalln(err)
	}

	logMessage(sessionID)

	return sessionID
}

func closePreviousChatSessions() {
	url := "http://localhost:3000/session/closeall/-1"

	req, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	defer req.Body.Close()
}

// pollMessagges polls raw messages from the events endpoint, unmarshals them
// and sends them to the "messagesChannel" CharacterMessage channel
func pollMessages(messagesChannel chan CharacterMessage) {
	for {
		time.Sleep(CHARACTER_MESSAGE_POLL_FREQUENCY)

		var messages []CharacterMessage

		response, _ := http.Get("http://localhost:3000/events")
		responseByte, _ := io.ReadAll(response.Body)
		err := json.Unmarshal(responseByte, &messages)
		if err != nil {
			log.Fatalln(err)
		}

		for _, message := range messages {
			messagesChannel <- message
		}
	}
}

// produceLines grabs the raw text sentences returned by the chatbot in "messagesChannel" and
// segments it in lines that have a maximum width. Those lines are then
// sent to the "lineChan" string channel to be rendered in screen.
func produceLines(messagesChannel chan CharacterMessage, lineChan chan string) {
	for message := range messagesChannel {
		var line string
		for _, word := range strings.Fields(message.Text) {
			if len(line+" "+word) > CHARACTER_TEXT_WIDTH {
				lineChan <- line
				line = word
			} else {
				if len(line) == 0 {
					line = word
					continue
				}
				line += " " + word
			}
		}
		lineChan <- line
	}
}

// sendMessage sends a message to the character using HTTP POST request.
// The message is sent as a JSON string with a "message" field containing the
// input text provided.
func sendMessage(messageUrl, inputText string) {
	messageJson := fmt.Sprintf(`{"message": "%s"}`, inputText)
	logMessage("User message", messageJson)
	message := json.RawMessage(messageJson)

	_, err := http.Post(messageUrl, "application/json", bytes.NewBuffer(message))
	if err != nil {
		log.Println(err)
	}
}

func loadSpeechSprites() SpeechSprites {
	speechSprites := make(SpeechSprites, SPEECH_SPRITE_SIZE)
	for i, _ := range speechSprites {
		fileName := "sprite" + strconv.Itoa(i) + ".txt"
		filePath := path.Join("sprites", fileName)

		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalln("Error opening file:", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			speechSprites[i] = append(speechSprites[i], line)
		}

		if err := scanner.Err(); err != nil {
			log.Fatalln(err)
		}
	}
	return speechSprites
}

func rowToRuneSlice(row string) []rune {
	var returnRune []rune
	for _, character := range row {
		returnRune = append(returnRune, rune(character))
	}
	return returnRune
}

func (speechSprites SpeechSprites) loopSpeechSprites(screen tcell.Screen, speechTime time.Duration) {
	for spriteTime := time.Duration(0); spriteTime < speechTime; spriteTime += SPEECH_SPRITE_SIZE * SPEECH_SPRITE_CHANGE_TIME {
		for _, sprite := range speechSprites {
			for num, line := range sprite {
				row := rowToRuneSlice(line)
				screen.SetContent(0, num, ' ', row, 0)
			}
			screen.Show()

			time.Sleep(SPEECH_SPRITE_CHANGE_TIME)
		}
	}
}

func renderResponses(screen tcell.Screen, speechSprites SpeechSprites, lineChan chan string) {
	var textBuffer []string
	var textAreaStyle tcell.Style
	textAreaStyle = textAreaStyle.Bold(true)

	sW, sH := screen.Size()
	viewPort := views.NewViewPort(screen, sW/2+sW/5, sH/2-sH/3, 30, 10)

	textArea := views.NewTextArea()
	textArea.SetStyle(textAreaStyle)
	textArea.Init()
	textArea.SetView(viewPort)

	for line := range lineChan {
		if line == "" {
			continue
		}

		speechTime := time.Duration(len(line)) * TYPING_DELAY
		go speechSprites.loopSpeechSprites(screen, speechTime)

		if len(textBuffer) >= CHARACTER_MESSAGE_LINES_COUNT {
			textBuffer = textBuffer[1:]
		}
		textBuffer = append(textBuffer, "")

		for _, char := range line {
			textBuffer[len(textBuffer)-1] += string(char)
			textArea.SetLines(textBuffer)
			textArea.Draw()
			screen.Show()
			time.Sleep(TYPING_DELAY)
		}
	}
}

func handleUserInput(screen tcell.Screen, messageUrl string, lineChan chan string) {
	var textAreaStyle tcell.Style
	textAreaStyle = textAreaStyle.Bold(true)
	textAreaStyle = textAreaStyle.Italic(true)
	sW, sH := screen.Size()
	viewPort := views.NewViewPort(screen, sW/24, sH-sH/8, sW-sW/12, 1)

	textArea := views.NewTextArea()
	textArea.SetStyle(textAreaStyle)
	textArea.EnableCursor(true)
	textArea.Init()
	textArea.SetView(viewPort)

	var textBuffer string
	for {
		event := screen.PollEvent()
		switch event := event.(type) {
		case *tcell.EventKey:
			key := event.Key()

			switch key {
			case tcell.KeyEsc:
				screen.Fini()
				panic("SHUTDOWN")
			case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
				if len(textBuffer) != 0 {
					textBuffer = textBuffer[:len(textBuffer)-1]
				}
			case tcell.KeyEnter:
				lineChan <- CHARACTER_MESSAGE_SEPARATION_LINE
				sendMessage(messageUrl, textBuffer)
				textBuffer = ""
			default:
				textBuffer += string(event.Rune())
			}
			textArea.SetContent(TYPING_PROMPT + textBuffer)
			textArea.Draw()
			screen.Show()
		}
	}
}

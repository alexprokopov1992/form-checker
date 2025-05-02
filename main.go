package main

import (
	"bufio"
	"fmt"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	formURL      string
	botToken     string
	chatID       string
	threadID     string
	pollInterval time.Duration
)

func main() {
	if err := loadSettings("settings.txt"); err != nil {
		fmt.Println("Помилка завантаження налаштувань:", err)
		return
	}

	wasOpen := false

	client := &http.Client{}

	defer sendTelegramMessage("Бот чекер завершив роботу!\n" + formURL)

	sendTelegramMessage("Роботу бота розпочато!\n" + formURL)

	sendTelegramMessage("Бот чекер черги запущено")
	for {
		isOpen := checkForm(client)

		if isOpen {
			playWavFile()
			if !wasOpen {
				sendTelegramMessage("⚠️⚠️⚠️ Google форма відкрилась! ⚠️⚠️⚠️\n" + formURL)
				wasOpen = true
			} else {
				sendTelegramMessage("⚠️⚠️⚠️ Google форма доступна! ⚠️⚠️⚠️\n" + formURL)
			}
		}

		time.Sleep(pollInterval)
	}
}

func checkForm(client *http.Client) bool {
	req, _ := http.NewRequest("GET", formURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Помилка запиту: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Помилка читання тіла відповіді: %v\n", err)
		return false
	}

	body := string(bodyBytes)

	if strings.Contains(body, "This form is no longer accepting responses") ||
		strings.Contains(body, "closedform") {
		log.Println("Форма закрита.")
		return false
	}

	log.Println("Форма відкрита!")
	return true
}

func sendTelegramMessage(message string) {
	if botToken == "" {
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	resp, err := http.PostForm(apiURL, map[string][]string{
		"chat_id":           {chatID},
		"message_thread_id": {threadID},
		"text":              {message},
	})
	if err != nil {
		fmt.Printf("Помилка відправки повідомлення в Telegram: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Помилка від Telegram API: %s\n", resp.Status)
	}
}

func loadSettings(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	settings := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		settings[key] = value
	}

	formURL = settings["form_url"]
	botToken = settings["bot_token"]
	chatID = settings["chat_id"]
	threadID = settings["thread_id"]
	pollInterval, err = time.ParseDuration(settings["poll_interval"])
	if err != nil {
		return fmt.Errorf("неправильний формат poll_interval: %v", err)
	}

	return nil
}

func playWavFile() {
	f, _ := os.Open("alert.wav")
	streamer, format, _ := wav.Decode(f)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
}

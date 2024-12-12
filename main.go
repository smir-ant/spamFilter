package main

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
)

func main() {
	// Загружаем переменные из .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[ERROR]: Ошибка загрузки файла .env: %v", err)
	}

	// Читаем токен из переменных окружения
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("[ERROR]: Токен не найден в переменной окружения TOKEN")
	}

	// Настраиваем бота
	log.Println("[INFO]: Настройка бота...")
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("[ERROR]: Ошибка создания бота: %v", err)
		return
	}
	log.Println("[INFO]: Бот успешно создан.")

	// Определяем стоп-слова и разрешенные эмодзи
	stopWordsEng := []string{"1:1", "my channel", "text me", "slot", "clone", "ebay", "buy here", "crypto", "gift", "stock", "binance", "visa", "casino", "bans", "btc", "bitcoin", "eth", "market", "ethereum", "profit", "onlyfans", "farting", "private", "horny", "sex", "porno", "girlfriend", "pussy", "anal", "telegram", "toncoin", "notcoin", "not", "ton", "bot"}
	stopWordsRus := []string{"слот", "доходность", "крипт", "желающи"/*х/е*/, "развиваться в"/*(области, сфере)*/, "набираем", "дневная доходность", "ограничено", "количество мест", "играю тут", "только тут", "проплачен", "цыпочки", "интим", "скрытые фото", "скрытым фото", "казино", "заработок", "нарко", "приват"/*+ный*/, "подруг", "порно", "секс", "разде"/*ть/нешь*/, "пошлая", "гей", "аналь"/*ный/ная*/, "трах", "шлюх", "закрыть долги", "оплатить лечение", "нуждае"/*тся*/, "нуждаю"/*щимся*/, "выигрыш"/*ь/ы*/}

	// ПОВТОРЯЮЩИЕСЯ ЭМОДЗИ НЕ УБИРАТЬ - ЭТО РАЗНЫЕ СИМВОЛЫ. Если тебе кажется, что дубликаты, то нет, это обе версии эмодзи (с вариационным селектором и без)
	// что такое варианционный селектор, чекай в readme
	allowedEmojis := []string{
		"👍", "☝️", "☝", "❄️", "❄", "🤝", "✏️", "✏", "❤️", "❤", "💪", "🙏",
		"👀", "💼", "💸", "🔥", "⚡️", "⚡", "✅", "🎁", "💳", "🔼", "🆘", "📲",
		"📌", "🤷", "‼️", "‼", "🆒", "⬆️", "⬆", "⛔️", "👉", "👉🏻", "🥹", "🥺",
		"😎", "☺️", "☺", "😊", "😋", "🤔", "💥", "😳", "🤗", "😄", "😆", "😂", "🤣",
	}
	
	// Упрощенное регулярное выражение для поиска эмодзи
	emojiRegex := regexp.MustCompile(`[\p{So}\p{Sk}]`)

	// Обработчик текстовых сообщений
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		log.Printf("===============\nОбработка сообщения от %s: %s", c.Sender().Username, c.Message().Text)
		text := c.Message().Text
		user := c.Sender()

		// Проверка на наличие английских и русских стоп-слов
		// log.Println("Приступаю к проверке на наличие запрещенных слов...")
		if containsStopWords(text, stopWordsEng) || containsStopWords(replaceSimilarChars(text), stopWordsRus) {
			// log.Printf("x Обнаружено запрещенное слово в сообщении от %s", user.Username)
			return banUserAndDelete(bot, c, user)
		}

		// Проверка на запрещенные эмодзи
		// log.Println("Приступаю к проверке на запрещенные эмодзи...")
		if containsDisallowedEmojis(text, allowedEmojis, emojiRegex) {
			// log.Printf("x Запрещенный эмодзи от %s", user.Username)
			return banUserAndDelete(bot, c, user)
		}

		log.Println("[PASS]: Сообщение прошло все проверки.")
		return nil
	})

	// Обработчик видеокружков
	bot.Handle(telebot.OnVideoNote, func(c telebot.Context) error {
		log.Printf("[BAN]: Обнаружен видеокружок от %s, бан пользователя...", c.Sender().Username)
		return banUserAndDelete(bot, c, c.Sender())
	})

	// Удаление сообщений о вступлении пользователя
	bot.Handle(telebot.OnUserJoined, func(c telebot.Context) error {
		log.Printf("[JOIN]: Пользователь %s присоединился, удаление сообщения о вступлении...", c.Sender().Username)
		err := bot.Delete(c.Message())
		if err != nil {
			log.Printf("[ERROR]: Ошибка удаления сообщения о вступлении: %v", err)
		}
		return err
	})

	// Удаление сообщений о выходе пользователя (включая тех, кто был забанен)
	bot.Handle(telebot.OnUserLeft, func(c telebot.Context) error {
		log.Printf("[LEAVE]: Пользователь %s покинул чат, удаление сообщения о выходе...", c.Sender().Username)
		err := bot.Delete(c.Message())
		if err != nil {
			log.Printf("[ERROR]: Ошибка удаления сообщения о выходе: %v", err)
		}
		return err
	})

	log.Println("Бот запущен")
	bot.Start()
}

// Бан пользователя без удаления последнего сообщения
func banUserAndDelete(bot *telebot.Bot, c telebot.Context, user *telebot.User) error {
	log.Printf("[BAN]: Бан пользователя %s и удаление...", user.Username)
	chat := c.Chat()
	chatMember := &telebot.ChatMember{User: user}

	// Удаляем сообщение, содержащее запрещённый контент
	// log.Printf("[BAN]: Удаление сообщения пользователя %s...", user.Username)
	if err := bot.Delete(c.Message()); err != nil {
		log.Printf("[ERROR]: Ошибка удаления сообщения пользователя %s: %v", user.Username, err)
		return err
	}

	// Баним пользователя
	err := bot.Ban(chat, chatMember)
	if err != nil {
		log.Printf("[ERROR]: Ошибка бана пользователя %s: %v", user.Username, err)
	}
	return err
}

// Замена похожих символов
func replaceSimilarChars(text string) string {
	// log.Println("Заменяем похожие символы в тексте...")
	replacements := map[string]string{
		"a": "а", "o": "о", "e": "е", "p": "р", "c": "с", "x": "х", "y": "у", "m": "м",
		"w": "ш", "t": "т", "k": "к", "b": "в", "n": "п", "0": "о", "3": "з", "6": "б", "u": "и",
		// Добавьте любые другие необходимые замены.
	}

	normalizedText := text
	for original, replacement := range replacements {
		normalizedText = strings.ReplaceAll(normalizedText, original, replacement)
	}
	return normalizedText
}


// Проверка на наличие стоп-слов
func containsStopWords(text string, stopWords []string) bool {
	for _, word := range stopWords {
		if strings.Contains(strings.ToLower(text), word) {
			log.Printf("[TRIGER]: Найдено стоп-слово [%s]", word)
			return true
		}
	}
	return false
}

// Проверка на запрещенные эмодзи
func containsDisallowedEmojis(text string, allowedEmojis []string, emojiRegex *regexp.Regexp) bool {
	// log.Println("Запущена проверка эмодзи в сообщении...")
	
	// Создаем множество разрешенных эмодзи
	allowedSet := make(map[string]struct{})
	for _, emoji := range allowedEmojis {
		allowedSet[emoji] = struct{}{}
	}

	// Ищем все эмодзи в тексте
	emojis := emojiRegex.FindAllString(text, -1)
	log.Printf("[INFO]: Эмодзи в тексте: %v", emojis)

	for _, emoji := range emojis {
		// Удаляем вариационные селекторы (если они есть)
		normalizedEmoji := strings.ReplaceAll(emoji, "\uFE0F", "")
		log.Printf("[INFO]: Проверяем эмодзи: %s (нормализованный: %s)", emoji, normalizedEmoji)

		// Проверяем оригинальный и нормализованный вариант эмодзи
		if _, found := allowedSet[emoji]; !found {
			if _, found := allowedSet[normalizedEmoji]; !found {
				log.Printf("[TRIGER]: Найден запрещенный эмодзи: %s", emoji)
				return true
			}
		}
	}
	return false
}

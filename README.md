# Сборка

```zsh
go mod tidy && \
GOOS=windows GOARCH=amd64 go build -o spam_filter_bot.exe && \
mkdir -p spamFilter && \
cp -f .env spam_filter_bot.exe spamFilter/
```

тут 4 команды:
1. Оптимизация зависимостей + создание go.sum
2. Сборка под windows
3. Создание папки `spamFilter`
4. Дублирование результата сборки и `.env` в эту папку

Готово. Не забудь, что `.exe` и `.env` должны идти рядом - поэтому я и вытаскиваю их в папку.

Ну и не забудь указать свой токен в `.env`, убрав example :)
---

### Зачем в наборе разрешенных эмодзи некоторые их них дублируются?

<!-- это кусок кода из main.go (44-47 строки) для github markdown -->

https://github.com/smir-ant/spamFilter/blob/ebe787b1fe16767ca300f39a2f93e9e56a0d26bc/main.go#L44C1-L47C70

Вариационный селектор — это специальный "невидимый" символ, который указывает компьютеру, как отображать определённые символы, такие как эмодзи.

Зачем он нужен?
Многие символы, включая эмодзи, могут иметь несколько вариаций отображения. Например:

🖤 (чёрное сердце) может быть показано как простой текстовый символ "♥" или как цветное изображение (эмодзи).

Вариационный селектор (Unicode U+FE0F) сообщает, что символ нужно отображать как полноценный эмодзи, а не как обычный текст. Если вариационный селектор отсутствует, система может показать символ в текстовом виде.


---

### Таймаут?

https://github.com/smir-ant/spamFilter/blob/ebe787b1fe16767ca300f39a2f93e9e56a0d26bc/main.go#L30C1-L30C58

Это максимальное время, в течение которого бот ждёт новые обновления от серверов Telegram в одном запросе. Telegram устанавливает ограничения на количество запросов от одного бота. Например:
- Для большинства методов: не более 100 запросов в секунду для одного бота.
- Хотя запросы на получение обновлений (Long Polling) обычно менее строго регулируются, Telegram может ограничить или замедлить ваш бот при слишком частых запросах.

---

made by <a href="https://github.com/smir-ant">smir-ant</a>
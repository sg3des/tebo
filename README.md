# tebo - smart telegram bot

## Features

## Usage

```shell
go get github.com/sg3des/tebo
```


### Initialize

Initialize new instance of bot and new connection to telegram server, at this time try to load bot information `/getMe` method, read/initialize history.

```go
func main() {
	bot, err := tebo.NewBot(token, ".history")
	if err != nil {
		// unavailable connection to telegram server or failed to initialize a history
	}

	bot.Handle("/hello", helloHandler)
}

func helloHandler(msg tebo.Message) string {
	return fmt.Sprintf("Hello %s", msg.From.Username)
}
```


### Handle commands

Handle commands is like with http handlers:

```go
bot.Handle("/hello", func(msg tebo.Message) string {
	//
	// your logic
	//
	return "response"
})
```

Complex commands routing, with more than one slash, arguments, allowed prefined arguments, or regular expressions:

```go
bot.Handle("/hello/world", ...)
bot.Handle("/alarm 08:00", ...)
bot.Handle("/some-command start|stop", ...)
bot.Handle("/first/\\w+ [0-9]", ...)
```


### Middleware

As well as the famous web frameworks, `tebo` allowed used middleware functions:

```go
func CheckTrustedUsers(next tebo.HandleFunc, msg tebo.Message) (tebo.HandleFunc, bool) {
	if msg.From.Username != "trustedUsername" {
		return nil, false
	}

	return next, true
}
```

For all handlers:

```go
bot.Pre(CheckTrustedUsers)
```

For handler:

```go
bot.Handle("/command", handler, CheckTrustedUsers)
```



### Send messages

`tebo` allow to send messages to known users(to exists chat), without any commands from user. It can be convenient for sending notifications, etc.

```go
err := bot.SendMessage(chatid, "some text")
// or send images
err := bot.SendPhoto(chatid, imgReader, "image caption")
```


### Resolve chat id

Use `chatid` in UI may be ugly, `chatname` is more prettier. Bot will find the chat name in the saved history and return its ID.

```go
chatid, ok := bot.LookupChatID(chatname)
```


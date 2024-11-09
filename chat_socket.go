package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// ارتقاء دهنده وب‌سوکت
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// نقشه‌ای برای نگهداری ارتباط کلاینت‌ها
var clients = make(map[*websocket.Conn]bool)

// کانال برای پیام‌ها
var broadcast = make(chan Message)

// ساختار پیام
type Message struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// هندلر وب‌سوکت
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// ارتقاء به وب‌سوکت
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// اضافه کردن کلاینت به لیست
	clients[ws] = true

	for {
		var msg Message
		// خواندن پیام از کلاینت
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error to read message: %v", err)
			delete(clients, ws)
			break
		}
		// ارسال پیام به کانال
		broadcast <- msg
	}
}

// هندلر برای ارسال پیام‌ها به کلاینت‌ها
func handleMessages() {
	for {
		// دریافت پیام از کانال
		msg := <-broadcast

		// ارسال پیام به تمام کلاینت‌ها
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error to send data to client: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// مسیر برای هندلر وب‌سوکت
	http.HandleFunc("/ws", handleConnections)

	// گوروتین برای مدیریت پیام‌ها
	go handleMessages()

	// اجرای سرور روی پورت 6771
	log.Println("rune in port  6771")
	err := http.ListenAndServe(":6771", nil)
	if err != nil {
		log.Fatal("error in run server: ", err)
	}
}

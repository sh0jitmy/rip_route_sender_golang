package main

import (
    "log"
    "net/http"
    "time"
)

func main() {
    // ルーティングテーブル初期化
    routingTable = make(map[string]RouteInfo)

    // 定期的なRIP更新のためのタイマー（30秒ごと）
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for {
            <-ticker.C
            log.Println("Sending periodic RIPv2 updates")
            sendPeriodicUpdates() // RIP更新を送信
        }
    }()

    // RIPv2パケットの受信を開始
    go listenForRipUpdates()

    // REST APIのルーティング設定
    http.HandleFunc("/send_route", handleRoute)
    http.HandleFunc("/delete_route", handleDeleteRoute) // ルート削除用のエンドポイント

    log.Println("REST API server is running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

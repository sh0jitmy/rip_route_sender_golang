package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
)

// 経路情報の構造体
type RouteInfo struct {
    Network string `json:"network"`
    Mask    string `json:"mask"`
    Metric  int    `json:"metric"`
}

// ルーティングテーブルのマップ
var routingTable map[string]RouteInfo

// APIリクエストを処理するハンドラー
func handleRoute(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var route RouteInfo
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    err = json.Unmarshal(body, &route)
    if err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // ルーティングテーブルに追加
    routingTable[route.Network] = route

    // RIPv2パケットを生成して送信
    err = sendRipPacket(route)
    if err != nil {
        http.Error(w, "Failed to send RIP packet", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("RIP packet sent successfully"))
}

// ルート情報の削除を処理するハンドラー
func handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var route RouteInfo
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    err = json.Unmarshal(body, &route)
    if err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // ルート情報の削除とポイゾニング
    _, exists := routingTable[route.Network]
    if exists {
        log.Printf("Poisoning route for network %s\n", route.Network)
        // ルートポイゾニング: メトリックを16にしてRIPパケットを送信
        route.Metric = 16
        err = sendRipPacket(route)
        if err != nil {
            http.Error(w, "Failed to send RIP poison packet", http.StatusInternalServerError)
            return
        }
        // テーブルから削除
        delete(routingTable, route.Network)
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Route poisoned and deleted successfully"))
}

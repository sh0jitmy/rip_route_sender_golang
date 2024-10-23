package main

import (
    "bytes"
    "encoding/binary"
    "encoding/json"
    "log"
    "net"
    "net/http"
)

// スプリットホライズンを考慮するために、受信元インターフェースの管理
var learnedFromInterface map[string]struct{} = make(map[string]struct{})

// RIPv2のパケットを受信して処理する
func listenForRipUpdates() {
    addr, err := net.ResolveUDPAddr("udp", ":520")
    if err != nil {
        log.Fatal("Failed to resolve UDP address:", err)
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        log.Fatal("Failed to listen on UDP port 520:", err)
    }
    defer conn.Close()

    log.Println("Listening for RIPv2 updates on UDP port 520")

    for {
        buf := make([]byte, 512)
        n, srcAddr, err := conn.ReadFromUDP(buf)
        if err != nil {
            log.Println("Failed to read from UDP:", err)
            continue
        }

        // パケットを解析し、経路情報をPOST送信
        log.Printf("Received RIPv2 update from %s\n", srcAddr.String())
        handleRipUpdate(buf[:n], srcAddr)
    }
}

// RIPv2パケットを解析し、経路情報を指定エンドポイントにPOST送信
func handleRipUpdate(packet []byte, srcAddr *net.UDPAddr) {
    if len(packet) < 4 {
        log.Println("Invalid RIPv2 packet")
        return
    }

    ripHeader := RIPHeader{
        Command:  packet[0],
        Version:  packet[1],
        Zero:     binary.BigEndian.Uint16(packet[2:4]),
    }

    // RIPv2 Response（Command = 2）だけを処理
    if ripHeader.Command != 2 {
        return
    }

    offset := 4
    for offset+20 <= len(packet) {
        var entry RIPEntry
        entry.AddressFamily = binary.BigEndian.Uint16(packet[offset : offset+2])
        entry.RouteTag = binary.BigEndian.Uint16(packet[offset+2 : offset+4])
        copy(entry.IPAddress[:], packet[offset+4:offset+8])
        copy(entry.SubnetMask[:], packet[offset+8:offset+12])
        copy(entry.NextHop[:], packet[offset+12:offset+16])
        entry.Metric = binary.BigEndian.Uint32(packet[offset+16 : offset+20])

        // 受信したインターフェースから学習しているかどうかを確認（スプリットホライズン）
        if _, learned := learnedFromInterface[srcAddr.IP.String()]; learned {
            log.Printf("Skipping route learned from the same interface (%s)\n", srcAddr.IP.String())
        } else {
            // 経路情報をJSONとしてエンドポイントにPOST送信
            route := RouteInfo{
                Network: net.IP(entry.IPAddress[:]).String(),
                Mask:    net.IP(entry.SubnetMask[:]).String(),
                Metric:  int(entry.Metric),
            }
            postRouteToAPI(route)
        }

        offset += 20
    }

    // 受信元インターフェースを記録（スプリットホライズン対策）
    learnedFromInterface[srcAddr.IP.String()] = struct{}{}
}

// 経路情報を指定されたエンドポイントにPOSTする
func postRouteToAPI(route RouteInfo) {
    jsonData, err := json.Marshal(route)
    if err != nil {
        log.Println("Failed to marshal route data:", err)
        return
    }

    endpoint := "http://example.com/route_update" // 送信先のエンドポイントURL
    resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Println("Failed to post route to API:", err)
        return
    }
    defer resp.Body.Close()

    log.Printf("Posted route %s to API, received status: %s\n", route.Network, resp.Status)
}

package main

import (
    "encoding/binary"
    "log"
    "net"
)

// RIPパケットのヘッダー
type RIPHeader struct {
    Command  uint8
    Version  uint8
    Zero     uint16
}

// RIP経路エントリ
type RIPEntry struct {
    AddressFamily uint16
    RouteTag      uint16
    IPAddress     [4]byte
    SubnetMask    [4]byte
    NextHop       [4]byte
    Metric        uint32
}

// RIPv2パケットを作成して送信する
func sendRipPacket(route RouteInfo) error {
    // UDP接続を設定
    addr, err := net.ResolveUDPAddr("udp", "224.0.0.9:520")
    if err != nil {
        return err
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return err
    }
    defer conn.Close()

    // RIPヘッダーを作成
    ripHeader := RIPHeader{
        Command: 2, // 2 = Response
        Version: 2, // RIPv2
        Zero:    0,
    }

    // RIPエントリを作成
    var entry RIPEntry
    entry.AddressFamily = 2 // IP (AF_INET)
    entry.RouteTag = 0      // Route Tagは0
    copy(entry.IPAddress[:], net.ParseIP(route.Network).To4())
    copy(entry.SubnetMask[:], net.ParseIP(route.Mask).To4())
    copy(entry.NextHop[:], net.ParseIP("0.0.0.0").To4())
    entry.Metric = uint32(route.Metric)

    // バッファにパケットをシリアライズ
    buf := make([]byte, 4+20) // RIPヘッダー(4バイト) + RIPエントリ(20バイト)
    buf[0] = ripHeader.Command
    buf[1] = ripHeader.Version
    binary.BigEndian.PutUint16(buf[2:], ripHeader.Zero)

    binary.BigEndian.PutUint16(buf[4:], entry.AddressFamily)
    binary.BigEndian.PutUint16(buf[6:], entry.RouteTag)
    copy(buf[8:], entry.IPAddress[:])
    copy(buf[12:], entry.SubnetMask[:])
    copy(buf[16:], entry.NextHop[:])
    binary.BigEndian.PutUint32(buf[20:], entry.Metric)

    // UDPソケットでパケットを送信
    _, err = conn.Write(buf)
    if err != nil {
        return err
    }

    log.Println("RIP packet sent:", buf)
    return nil
}

// 定期的に全ルート情報を送信する関数
func sendPeriodicUpdates() {
    for _, route := range routingTable {
        err := sendRipPacket(route)
        if err != nil {
            log.Println("Failed to send RIP packet:", err)
        }
    }
}

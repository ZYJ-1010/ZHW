package com.zhw.funds.revenue;

import com.sun.net.httpserver.HttpServer;
import com.zhw.funds.common.HttpJson;

import java.util.List;
import java.util.Map;

public class RevenueController {
    public static void register(HttpServer server) {
        server.createContext("/api/funds/revenue/templates", exchange -> HttpJson.ok(exchange, Map.of(
                "items", List.of(Map.of(
                        "id", 1,
                        "name", "default-free",
                        "gameType", "free",
                        "platformBps", 1000,
                        "creatorBps", 3000,
                        "memberBps", 6000,
                        "status", "active"
                ))
        )));
        server.createContext("/api/funds/revenue/simulate", exchange -> HttpJson.ok(exchange, Map.of(
                "gameId", 0,
                "amountCent", 0,
                "items", List.of(
                        Map.of("role", "platform", "amountCent", 0),
                        Map.of("role", "creator", "amountCent", 0),
                        Map.of("role", "member_pool", "amountCent", 0)
                ),
                "persisted", false
        )));
        server.createContext("/api/funds/revenue/generate", exchange -> HttpJson.ok(exchange, Map.of(
                "recordNo", "REV-PLACEHOLDER",
                "status", "pending_settlement",
                "persisted", false
        )));
        server.createContext("/api/funds/revenue/records", exchange -> HttpJson.ok(exchange, Map.of(
                "items", List.of()
        )));
        server.createContext("/api/funds/profit-sharing/receivers", exchange -> HttpJson.ok(exchange, Map.of(
                "mode", "placeholder",
                "status", "saved"
        )));
        server.createContext("/api/funds/profit-sharing/orders", exchange -> HttpJson.ok(exchange, Map.of(
                "outOrderNo", "PS-PLACEHOLDER",
                "status", "simulated"
        )));
        server.createContext("/api/funds/profit-sharing/return-orders", exchange -> HttpJson.ok(exchange, Map.of(
                "outReturnNo", "PSR-PLACEHOLDER",
                "status", "simulated"
        )));
    }
}

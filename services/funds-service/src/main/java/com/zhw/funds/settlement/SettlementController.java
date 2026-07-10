package com.zhw.funds.settlement;

import com.sun.net.httpserver.HttpServer;
import com.zhw.funds.common.HttpJson;

import java.util.List;
import java.util.Map;

public class SettlementController {
    public static void register(HttpServer server) {
        server.createContext("/api/funds/settlements/offline", exchange -> HttpJson.ok(exchange, Map.of(
                "settlementNo", "SETTLE-PLACEHOLDER",
                "method", "offline",
                "status", "registered"
        )));
        server.createContext("/api/funds/settlements", exchange -> HttpJson.ok(exchange, Map.of(
                "items", List.of()
        )));
        server.createContext("/api/funds/bills/download", exchange -> HttpJson.ok(exchange, Map.of(
                "taskNo", "BILL-PLACEHOLDER",
                "status", "created"
        )));
        server.createContext("/api/internal/funds/reconcile-runner", exchange -> HttpJson.ok(exchange, Map.of(
                "status", "ok",
                "diffCount", 0
        )));
    }
}

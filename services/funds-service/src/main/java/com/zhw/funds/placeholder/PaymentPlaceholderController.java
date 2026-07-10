package com.zhw.funds.placeholder;

import com.sun.net.httpserver.HttpServer;
import com.zhw.funds.common.HttpJson;

import java.util.Map;

public class PaymentPlaceholderController {
    public static void register(HttpServer server) {
        server.createContext("/api/funds/payment-precreate-placeholder", exchange -> HttpJson.ok(exchange, Map.of(
                "orderNo", "FREE-PLACEHOLDER",
                "payStatus", "free_no_pay",
                "amountCent", 0,
                "needWechatPay", false
        )));
        server.createContext("/api/funds/payment-callback-placeholder", exchange -> HttpJson.ok(exchange, Map.of(
                "received", true,
                "verified", true,
                "mode", "placeholder"
        )));
        server.createContext("/api/internal/pay/callback-placeholder", exchange -> HttpJson.ok(exchange, Map.of(
                "received", true,
                "verified", true,
                "mode", "internal-placeholder"
        )));
    }
}

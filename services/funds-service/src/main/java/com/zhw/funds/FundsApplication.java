package com.zhw.funds;

import com.zhw.funds.common.ApiResponse;
import com.zhw.funds.common.HttpJson;
import com.zhw.funds.common.Json;
import com.zhw.funds.placeholder.PaymentPlaceholderController;
import com.zhw.funds.revenue.RevenueController;
import com.zhw.funds.settlement.SettlementController;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.Map;

public class FundsApplication {
    public static void main(String[] args) throws IOException {
        int port = Integer.parseInt(System.getenv().getOrDefault("FUNDS_SERVICE_PORT", "8081"));
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);

        server.createContext("/health", exchange -> HttpJson.ok(exchange, Map.of(
                    "status", "ok",
                    "service", "funds-service"
        )));
        PaymentPlaceholderController.register(server);
        RevenueController.register(server);
        SettlementController.register(server);

        server.start();
        System.out.println("funds-service listening on :" + port);
    }
}

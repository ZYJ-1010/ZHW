package com.zhw.funds.common;

import com.sun.net.httpserver.HttpExchange;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;

public class HttpJson {
    public static void ok(HttpExchange exchange, Object data) throws IOException {
        write(exchange, 200, ApiResponse.ok(data));
    }

    public static void write(HttpExchange exchange, int status, Object payload) throws IOException {
        byte[] body = Json.stringify(payload).getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
        exchange.sendResponseHeaders(status, body.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(body);
        }
    }
}

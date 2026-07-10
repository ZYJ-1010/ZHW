package com.zhw.funds.common;

import java.util.Map;
import java.util.stream.Collectors;

public class Json {
    public static String stringify(Object value) {
        if (value instanceof ApiResponse<?> response) {
            return "{\"code\":" + response.code
                    + ",\"message\":\"" + escape(response.message) + "\""
                    + ",\"data\":" + stringify(response.data)
                    + "}";
        }
        if (value instanceof Map<?, ?> map) {
            return "{" + map.entrySet().stream()
                    .map(entry -> "\"" + escape(String.valueOf(entry.getKey())) + "\":" + stringify(entry.getValue()))
                    .collect(Collectors.joining(",")) + "}";
        }
        if (value instanceof Number || value instanceof Boolean) {
            return String.valueOf(value);
        }
        if (value == null) {
            return "null";
        }
        return "\"" + escape(String.valueOf(value)) + "\"";
    }

    private static String escape(String value) {
        return value.replace("\\", "\\\\").replace("\"", "\\\"");
    }
}


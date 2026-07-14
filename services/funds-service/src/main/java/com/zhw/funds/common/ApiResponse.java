package com.zhw.funds.common;

public class ApiResponse<T> {
    public final int code;
    public final String message;
    public final T data;

    private ApiResponse(int code, String message, T data) {
        this.code = code;
        this.message = message;
        this.data = data;
    }

    public static <T> ApiResponse<T> ok(T data) {
        return new ApiResponse<>(0, "ok", data);
    }
}


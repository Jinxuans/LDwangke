import axios from "axios";

const BASE = import.meta.env.VITE_API_BASE || "/api/v1/mall";

let _tid = "";
export function setTid(tid: string) {
    _tid = tid;
}

const http = axios.create({ timeout: 15000 });

http.interceptors.request.use((config) => {
    const token = localStorage.getItem("c_token");
    if (token) {
        config.headers["X-C-Token"] = token;
    }
    return config;
});

http.interceptors.response.use(
    (res) => {
        const { code, data, message } = res.data;
        if (code === 0) return data;
        return Promise.reject(new Error(message || "请求失败"));
    },
    (err) => Promise.reject(err),
);

function url(path: string) {
    return `${BASE}/${_tid}${path}`;
}

export function getShopInfo() {
    return http.get(url("/info"));
}

export function getProducts() {
    return http.get(url("/products"));
}

export function getProductDetail(cid: number) {
    return http.get(url(`/product/${cid}`));
}

export function queryOrder(keyword: string) {
    return http.get(url("/search"), { params: { keyword } });
}

export function queryCourse(data: { cid: number; userinfo: string }) {
    return http.post(url("/query"), data);
}

export function placeOrder(data: {
    cid: number;
    account: string;
    remark?: string;
}) {
    return http.post(url("/order"), data);
}

export function getMyOrders(params?: { page?: number; limit?: number }) {
    return http.get(url("/orders"), { params });
}

export function getOrderDetail(oid: number) {
    return http.get(url(`/order/${oid}`));
}

export function getPayChannels() {
    return http.get(url("/pay/channels"));
}

export function checkPay(outTradeNo: string) {
    return http.get(url("/pay/check"), { params: { out_trade_no: outTradeNo } });
}

export function confirmPay(outTradeNo: string) {
    return http.post(url("/pay/confirm"), { out_trade_no: outTradeNo });
}

export function createPay(data: {
    cid: number;
    account: string;
    password: string;
    pay_type: string;
    remark?: string;
    course_id?: string;
    course_name?: string;
    course_kcjs?: string;
}) {
    return http.post(url("/pay"), data);
}

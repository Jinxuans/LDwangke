import axios from "axios";

const BASE = import.meta.env.VITE_API_BASE || "/api/v1/mall";
const C_TOKEN_KEY = "c_token";
const C_USER_KEY = "c_user";
const GUEST_ORDER_KEY = "mall_guest_orders";
const PROMOTER_CODE_KEY = "mall_promoter_code";

let _tid = "";
export function setTid(tid: string) {
    _tid = tid;
}

export interface CUserProfile {
    id: number;
    account: string;
    nickname: string;
    tid: string;
    invite_code?: string;
}

export interface MallProfile {
    id: number;
    tid: number;
    account: string;
    nickname: string;
    phone: string;
    invite_code: string;
    referrer_id: number;
    referrer_account?: string;
    referrer_nickname?: string;
    commission_money: string;
    commission_cdmoney: string;
    commission_total: string;
    promotion_orders: number;
    promotion_enabled: boolean;
    register_enabled: boolean;
    commission_rate: number;
    addtime: string;
    mall_config: {
        register_enabled: boolean;
        promotion_enabled: boolean;
        commission_rate: number;
        show_categories: boolean;
        popup_notice_html?: string;
        customer_service?: {
            enabled?: boolean;
            type?: string;
            value?: string;
            label?: string;
        };
    };
}

export async function triggerCustomerService(config?: {
    enabled?: boolean;
    type?: string;
    value?: string;
    label?: string;
}) {
    if (!config?.enabled || !config?.value) return { ok: false, message: "客服信息未配置" };
    const type = String(config.type || "").toLowerCase();
    const value = String(config.value || "").trim();
    if (!value) return { ok: false, message: "客服信息未配置" };

    if (type === "link") {
        window.open(value, "_blank");
        return { ok: true, message: "正在打开客服链接" };
    }
    return {
        ok: true,
        mode: "popup" as const,
        title:
            type === "phone"
                ? "电话客服"
                : type === "qq"
                  ? "QQ客服"
                  : "微信客服",
        label: config.label || "联系客服",
        type,
        value,
    };
}

export interface PromotionOrderItem {
    id: number;
    out_trade_no: string;
    product_name: string;
    course_name: string;
    buyer_account: string;
    money: string;
    commission_amount: string;
    commission_rate: string;
    status: number;
    status_text: string;
    addtime: string;
    paytime: string;
}

export interface MallWithdrawItem {
    id: number;
    tid: number;
    c_uid: number;
    amount: number;
    method: string;
    account_name: string;
    account_no: string;
    bank_name: string;
    note: string;
    status: number;
    audit_remark: string;
    audit_uid: number;
    addtime: string;
    audit_time: string;
}

export interface GuestMallOrderRef {
    tid: string;
    outTradeNo: string;
    accessToken: string;
    createdAt: number;
}

export interface MallPayOrderItem {
    id: number;
    out_trade_no: string;
    trade_no: string;
    cid: number;
    product_name: string;
    course_name: string;
    school?: string;
    account: string;
    remark: string;
    pay_type: string;
    pay_url?: string;
    status: number;
    status_text: string;
    money: string;
    order_id: number;
    addtime: string;
    paytime: string;
}

export function getCUserToken() {
    return localStorage.getItem(C_TOKEN_KEY) || "";
}

export function getCUserProfile() {
    const raw = localStorage.getItem(C_USER_KEY);
    if (!raw) return null;
    try {
        return JSON.parse(raw) as CUserProfile;
    } catch {
        return null;
    }
}

export function isCUserLoggedIn(tid?: string) {
    const token = getCUserToken();
    const profile = getCUserProfile();
    if (!token || !profile) return false;
    return tid ? profile.tid === tid : true;
}

export function saveCUserSession(data: {
    token: string;
    id: number;
    account: string;
    nickname: string;
    tid: string;
    invite_code?: string;
}) {
    localStorage.setItem(C_TOKEN_KEY, data.token);
    localStorage.setItem(
        C_USER_KEY,
        JSON.stringify({
            id: data.id,
            account: data.account,
            nickname: data.nickname,
            tid: data.tid,
            invite_code: data.invite_code || "",
        }),
    );
}

export function clearCUserSession() {
    localStorage.removeItem(C_TOKEN_KEY);
    localStorage.removeItem(C_USER_KEY);
}

export function getGuestMallOrders(tid?: string) {
    const raw = localStorage.getItem(GUEST_ORDER_KEY);
    if (!raw) return [] as GuestMallOrderRef[];
    try {
        const list = JSON.parse(raw) as GuestMallOrderRef[];
        if (!Array.isArray(list)) return [];
        return tid ? list.filter((item) => item.tid === tid) : list;
    } catch {
        return [];
    }
}

function saveGuestMallOrders(list: GuestMallOrderRef[]) {
    localStorage.setItem(GUEST_ORDER_KEY, JSON.stringify(list.slice(0, 20)));
}

export function saveGuestMallOrder(ref: GuestMallOrderRef) {
    if (!ref.accessToken) return;
    const list = getGuestMallOrders().filter(
        (item) => !(item.tid === ref.tid && item.outTradeNo === ref.outTradeNo),
    );
    list.unshift(ref);
    saveGuestMallOrders(list);
}

export function removeGuestMallOrder(tid: string, outTradeNo: string) {
    const list = getGuestMallOrders().filter(
        (item) => !(item.tid === tid && item.outTradeNo === outTradeNo),
    );
    saveGuestMallOrders(list);
}

export function removeGuestMallOrders(tid: string, outTradeNos: string[]) {
    const keys = new Set(outTradeNos.filter(Boolean));
    if (!keys.size) return;
    const list = getGuestMallOrders().filter(
        (item) => !(item.tid === tid && keys.has(item.outTradeNo)),
    );
    saveGuestMallOrders(list);
}

function promoterStorageKey(tid?: string) {
    return tid ? `${PROMOTER_CODE_KEY}:${tid}` : PROMOTER_CODE_KEY;
}

export function getMallPromoterCode(tid?: string) {
    return localStorage.getItem(promoterStorageKey(tid)) || "";
}

export function saveMallPromoterCode(code: string, tid?: string) {
    const normalized = String(code || "").trim().toUpperCase();
    if (!normalized) return;
    localStorage.setItem(promoterStorageKey(tid), normalized);
}

export function clearMallPromoterCode(tid?: string) {
    localStorage.removeItem(promoterStorageKey(tid));
}

const http = axios.create({ timeout: 15000 });

function extractRequestId(source?: any) {
    const headers = source?.headers ?? source?.response?.headers ?? {};
    const requestId =
        headers["x-request-id"] ??
        headers["X-Request-ID"] ??
        source?.data?.request_id ??
        source?.response?.data?.request_id;
    if (Array.isArray(requestId)) {
        return requestId[0] || "";
    }
    return typeof requestId === "string" ? requestId : "";
}

function createRequestError(message: string, source?: any) {
    const requestId = extractRequestId(source);
    const finalMessage = requestId ? `${message} [RID: ${requestId}]` : message;
    const error = new Error(finalMessage) as Error & { requestId?: string };
    if (requestId) {
        error.requestId = requestId;
        console.error("Mall API request failed", {
            requestId,
            status: source?.status ?? source?.response?.status,
            url: source?.config?.url ?? source?.response?.config?.url,
        });
    }
    return error;
}

http.interceptors.request.use((config) => {
    const token = getCUserToken();
    if (token) {
        config.headers["X-C-Token"] = token;
    }
    return config;
});

http.interceptors.response.use(
    (res) => {
        const { code, data, message } = res.data;
        if (code === 0) return data;
        if (code === 401 || code === 1401) {
            clearCUserSession();
        }
        return Promise.reject(createRequestError(message || "请求失败", res));
    },
    (err) => Promise.reject(createRequestError(err?.message || "请求失败", err)),
);

function url(path: string) {
    return _tid ? `${BASE}/${_tid}${path}` : `${BASE}${path}`;
}

export function getShopInfo() {
    const key = `shop_info:${_tid || "default"}`;
    const cached = sessionStorage.getItem(key);
    if (cached) {
        try { return Promise.resolve(JSON.parse(cached)); } catch {}
    }
    return http.get(url("/info")).then((data) => {
        sessionStorage.setItem(key, JSON.stringify(data));
        return data;
    });
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

export function loginCUser(data: {
    account: string;
    password: string;
    guest_orders?: Array<{ out_trade_no: string; access_token: string }>;
}) {
    return http.post(url("/login"), data);
}

export function registerCUser(data: {
    account: string;
    password: string;
    nickname?: string;
    phone?: string;
    promoter_code?: string;
}) {
    return http.post(url("/register"), data);
}

export function placeOrder(data: {
    cid: number;
    school?: string;
    account: string;
}) {
    return http.post(url("/order"), data);
}

export function getMyOrders(params?: { page?: number; limit?: number }) {
    return http.get(url("/orders"), { params });
}

export function getMyProfile() {
    return http.get<any, MallProfile>(url("/me"));
}

export function getPromotionOrders(params?: { page?: number; limit?: number }) {
    return http.get<any, { list?: PromotionOrderItem[]; total?: number } | PromotionOrderItem[]>(
        url("/promotion/orders"),
        { params },
    );
}

export function createCommissionWithdraw(data: {
    amount: number;
    method?: string;
    account_name: string;
    account_no: string;
    bank_name?: string;
    note?: string;
}) {
    return http.post(url("/withdraw/request"), data);
}

export function getCommissionWithdrawRequests(params?: { page?: number; limit?: number; status?: number }) {
    return http.get<any, { list?: MallWithdrawItem[]; pagination?: { total?: number } } | MallWithdrawItem[]>(
        url("/withdraw/requests"),
        { params },
    );
}

export function getOrderDetail(oid: number) {
    return http.get(url(`/order/${oid}`));
}

export function getPayChannels() {
    return http.get(url("/pay/channels"));
}

export function checkPay(outTradeNo: string, accessToken?: string) {
    return http.get(url("/pay/check"), {
        params: { out_trade_no: outTradeNo, access_token: accessToken },
    });
}

export function confirmPay(outTradeNo: string, accessToken?: string) {
    return http.post(url("/pay/confirm"), {
        out_trade_no: outTradeNo,
        access_token: accessToken,
    });
}

export function getGuestOrder(outTradeNo: string, accessToken: string) {
    return http.get(url("/guest/order"), {
        params: { out_trade_no: outTradeNo, access_token: accessToken },
    });
}

export function createPay(data: {
    cid: number;
    school?: string;
    account: string;
    password: string;
    pay_type: string;
    promoter_code?: string;
    courses?: Array<{
        id: string;
        name: string;
        kcjs?: string;
    }>;
    course_id?: string;
    course_name?: string;
    course_kcjs?: string;
}) {
    return http.post<
        any,
        {
            out_trade_no: string;
            pay_url: string;
            money: string;
            access_token?: string;
        }
    >(url("/pay"), data);
}

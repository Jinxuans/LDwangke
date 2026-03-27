<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Button, Empty, NavBar, PullRefresh, showToast } from "vant";
import type { MallPayOrderItem } from "../api";
import {
    clearCUserSession,
    getCUserProfile,
    getGuestMallOrders,
    getGuestOrder,
    getMyOrders,
    isCUserLoggedIn,
    removeGuestMallOrder,
} from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";

const orders = ref<MallPayOrderItem[]>([]);
const loading = ref(true);
const refreshing = ref(false);
const loggedIn = ref(false);
const expandedMap = ref<Record<string, boolean>>({});
const guestOrderRefs = ref<
    Array<{
        tid: string;
        outTradeNo: string;
        accessToken: string;
        createdAt: number;
    }>
>([]);

const statusMap: Record<string, string> = {
    待支付: "status-default",
    已支付: "status-pending",
    下单中: "status-processing",
    已下单: "status-done",
};

const currentUser = computed(() => {
    if (!loggedIn.value) return null;
    return getCUserProfile();
});

const hasGuestOrders = computed(() => guestOrderRefs.value.length > 0);

function syncGuestOrderRefs() {
    guestOrderRefs.value = getGuestMallOrders(tid);
}

function displayOrderId(order: MallPayOrderItem) {
    return order.out_trade_no || `#${order.id}`;
}

function statusClass(order: MallPayOrderItem) {
    return statusMap[order.status_text] || "status-default";
}

function orderKey(order: MallPayOrderItem) {
    return order.out_trade_no || String(order.id);
}

function toggleDetail(order: MallPayOrderItem) {
    const key = orderKey(order);
    expandedMap.value = {
        ...expandedMap.value,
        [key]: !expandedMap.value[key],
    };
}

function detailVisible(order: MallPayOrderItem) {
    return !!expandedMap.value[orderKey(order)];
}

function goQuery(order: MallPayOrderItem) {
    router.push(`${basePath}/query?keyword=${encodeURIComponent(order.account || "")}`);
}

function canRepay(order: MallPayOrderItem) {
    return order.status === 0 && !!order.pay_url;
}

function goPay(order: MallPayOrderItem) {
    if (!order.pay_url) {
        showToast("该订单暂不可重新支付");
        return;
    }
    localStorage.setItem("pending_out_trade_no", order.out_trade_no || "");
    localStorage.setItem("pending_pay_url", order.pay_url);
    window.location.href = order.pay_url;
}

async function loadGuestOrders() {
    syncGuestOrderRefs();
    const refs = guestOrderRefs.value;
    if (!refs.length) {
        orders.value = [];
        return;
    }

    const list: MallPayOrderItem[] = [];
    for (const ref of refs) {
        try {
            const item = (await getGuestOrder(
                ref.outTradeNo,
                ref.accessToken,
            )) as unknown as MallPayOrderItem;
            list.push(item);
        } catch {
            removeGuestMallOrder(ref.tid, ref.outTradeNo);
            syncGuestOrderRefs();
        }
    }
    orders.value = list.sort((a, b) =>
        String(b.addtime || "").localeCompare(String(a.addtime || "")),
    );
}

async function loadOrders() {
    loading.value = true;
    try {
        if (!loggedIn.value) {
            await loadGuestOrders();
            return;
        }
        const res = (await getMyOrders()) as { list?: MallPayOrderItem[] } | MallPayOrderItem[];
        orders.value = Array.isArray(res) ? res : res?.list || [];
    } catch (e: any) {
        if ((e?.message || "").includes("登录") || (e?.message || "").includes("token")) {
            clearCUserSession();
            loggedIn.value = false;
            orders.value = [];
            showToast("登录已失效，请重新登录");
            return;
        }
        showToast(e?.message || "加载失败");
    } finally {
        loading.value = false;
        refreshing.value = false;
    }
}

async function onRefresh() {
    refreshing.value = true;
    await loadOrders();
}

function goLogin() {
    router.push(`${basePath}/login?redirect=${encodeURIComponent(`${basePath}/orders`)}`);
}

function logout() {
    clearCUserSession();
    loggedIn.value = false;
    orders.value = [];
    showToast("已退出登录");
}

onMounted(() => {
    loggedIn.value = isCUserLoggedIn(tid);
    syncGuestOrderRefs();
    loadOrders();
});
</script>

<template>
    <div class="orders-page">
        <NavBar title="订单" />

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>

        <div v-else-if="!loggedIn && !hasGuestOrders" class="login-state animate-fade-in-up">
            <div class="login-state-badge">支付订单</div>
            <h2 class="login-state-title">登录后查看你的支付订单</h2>
            <p class="login-state-desc">
                这里展示的是你在商城创建并支付的订单记录。课程处理进度请到“查进度”里按下单账号查询。
            </p>
            <div class="login-state-actions">
                <Button type="primary" round block @click="goLogin">会员登录</Button>
                <Button plain round block @click="router.push(`${basePath}/query`)">去查进度</Button>
            </div>
        </div>

        <PullRefresh v-else v-model="refreshing" @refresh="onRefresh">
            <div v-if="currentUser" class="member-bar">
                <div>
                    <div class="member-name">{{ currentUser.nickname || currentUser.account }}</div>
                    <div class="member-account">会员账号：{{ currentUser.account }}</div>
                </div>
                <Button size="small" plain round @click="logout">退出</Button>
            </div>
            <div v-else class="guest-bar">
                <div>
                    <div class="member-name">本机访客订单</div>
                    <div class="member-account">仅展示当前设备保存的支付订单记录</div>
                </div>
                <Button size="small" plain round @click="goLogin">会员登录</Button>
            </div>

            <Empty
                v-if="!orders.length"
                description="暂无支付订单"
                style="padding-top: 80px"
            />

            <div v-else class="order-list">
                <div
                    v-for="o in orders"
                    :key="o.out_trade_no || o.id"
                    class="order-card animate-fade-in-up"
                >
                    <div class="order-header">
                        <span class="order-id">支付单 {{ displayOrderId(o) }}</span>
                        <span class="status-badge" :class="statusClass(o)">
                            {{ o.status_text }}
                        </span>
                    </div>
                    <div class="order-body">
                        <div class="order-row">
                            <span class="order-label">商品</span>
                            <span class="order-value">{{ o.product_name || "商城商品" }}</span>
                        </div>
                        <div class="order-row" v-if="o.course_name">
                            <span class="order-label">课程</span>
                            <span class="order-value">{{ o.course_name }}</span>
                        </div>
                        <div class="order-row">
                            <span class="order-label">账号</span>
                            <span class="order-value">{{ o.account }}</span>
                        </div>
                        <div class="order-row" v-if="o.school">
                            <span class="order-label">学校</span>
                            <span class="order-value">{{ o.school }}</span>
                        </div>
                        <div class="order-row">
                            <span class="order-label">金额</span>
                            <span class="order-value price">¥{{ o.money }}</span>
                        </div>
                        <div class="order-row" v-if="o.order_id > 0">
                            <span class="order-label">业务单</span>
                            <span class="order-value">#{{ o.order_id }}</span>
                        </div>
                        <div class="order-row" v-if="o.paytime">
                            <span class="order-label">支付</span>
                            <span class="order-value muted">{{ o.paytime }}</span>
                        </div>
                        <div class="order-row" v-else-if="o.addtime">
                            <span class="order-label">创建</span>
                            <span class="order-value muted">{{ o.addtime }}</span>
                        </div>
                    </div>
                    <div class="order-actions">
                        <Button size="small" plain round @click="toggleDetail(o)">
                            {{ detailVisible(o) ? "收起详情" : "查看详情" }}
                        </Button>
                        <Button
                            v-if="canRepay(o)"
                            size="small"
                            type="primary"
                            round
                            class="pay-btn"
                            @click="goPay(o)"
                        >
                            去支付
                        </Button>
                        <Button
                            v-else
                            size="small"
                            round
                            class="query-btn"
                            :disabled="!o.account"
                            @click="goQuery(o)"
                        >
                            查进度
                        </Button>
                    </div>
                    <div v-if="detailVisible(o)" class="order-detail">
                        <div class="detail-title">下单详情</div>
                        <div class="detail-grid">
                            <div class="detail-item">
                                <span class="detail-label">支付状态</span>
                                <span class="detail-value">{{ o.status_text }}</span>
                            </div>
                            <div class="detail-item" v-if="o.course_name">
                                <span class="detail-label">下单课程</span>
                                <span class="detail-value">{{ o.course_name }}</span>
                            </div>
                            <div class="detail-item" v-if="o.school">
                                <span class="detail-label">学校</span>
                                <span class="detail-value">{{ o.school }}</span>
                            </div>
                            <div class="detail-item" v-if="o.remark">
                                <span class="detail-label">备注</span>
                                <span class="detail-value">{{ o.remark }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </PullRefresh>
    </div>
</template>

<style scoped>
.orders-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 60px;
}
.login-state {
    margin: 24px 14px 0;
    padding: 28px 20px;
    border-radius: 20px;
    background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
    box-shadow: var(--shadow-lg);
}
.login-state-badge {
    display: inline-flex;
    padding: 4px 10px;
    border-radius: var(--radius-full);
    background: var(--primary-bg);
    color: var(--primary-dark);
    font-size: 12px;
    font-weight: 700;
}
.login-state-title {
    margin-top: 14px;
    font-size: 22px;
    line-height: 1.3;
    color: var(--text-primary);
}
.login-state-desc {
    margin-top: 10px;
    font-size: 14px;
    color: var(--text-secondary);
}
.login-state-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 22px;
}
.member-bar,
.guest-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin: 12px 12px 0;
    padding: 14px;
    border-radius: var(--radius-lg);
}
.member-bar {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
}
.guest-bar {
    background: #fffaf0;
    border: 1px solid #f5e6b3;
}
.member-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
}
.member-account {
    margin-top: 4px;
    font-size: 12px;
    color: var(--text-muted);
}
.order-list {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
}
.order-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-color);
    overflow: hidden;
}
.order-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-light);
}
.order-id {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
}
.order-body {
    padding: 10px 14px;
    display: flex;
    flex-direction: column;
    gap: 6px;
}
.order-actions {
    display: flex;
    gap: 10px;
    padding: 0 14px 12px;
}
.order-actions :deep(.van-button) {
    flex: 1;
}
.pay-btn {
    background: linear-gradient(135deg, #1d4ed8 0%, #1e40af 100%) !important;
    border-color: #1e3a8a !important;
    color: #fff !important;
    box-shadow: 0 8px 18px rgba(29, 78, 216, 0.22);
}
.query-btn {
    background: #ffffff !important;
    border-color: #2563eb !important;
    color: #1e3a8a !important;
    font-weight: 600 !important;
}
.query-btn:disabled {
    color: #94a3b8 !important;
    border-color: #cbd5e1 !important;
    background: #f8fafc !important;
}
.order-detail {
    border-top: 1px solid var(--border-light);
    padding: 12px 14px 14px;
    background: #fafaf8;
}
.detail-title {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-secondary);
    margin-bottom: 10px;
}
.detail-grid {
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.detail-item {
    display: flex;
    gap: 10px;
    align-items: flex-start;
}
.detail-label {
    width: 72px;
    flex-shrink: 0;
    font-size: 12px;
    color: var(--text-muted);
}
.detail-value {
    flex: 1;
    word-break: break-all;
    font-size: 13px;
    color: var(--text-primary);
}
.order-row {
    display: flex;
    align-items: flex-start;
    gap: 8px;
}
.order-label {
    font-size: 12px;
    color: var(--text-muted);
    min-width: 32px;
    flex-shrink: 0;
    padding-top: 1px;
}
.order-value {
    font-size: 13px;
    color: var(--text-primary);
    flex: 1;
    word-break: break-all;
}
.order-value.price {
    color: #ef4444;
    font-weight: 600;
}
.order-value.muted {
    color: var(--text-muted);
}
</style>

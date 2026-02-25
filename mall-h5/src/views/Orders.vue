<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import { NavBar, Empty, PullRefresh, showToast } from "vant";
import { getMyOrders, getOrderDetail } from "../api";

const route = useRoute();
const orders = ref<any[]>([]);
const loading = ref(true);
const refreshing = ref(false);
let pollTimer: ReturnType<typeof setTimeout> | null = null;

const statusMap: Record<string, { label: string; cls: string }> = {
    待处理: { label: "待处理", cls: "status-pending" },
    进行中: { label: "进行中", cls: "status-processing" },
    已完成: { label: "已完成", cls: "status-done" },
    异常: { label: "异常", cls: "status-failed" },
    已取消: { label: "已取消", cls: "status-default" },
};

function getStatus(s: string) {
    return statusMap[s] || { label: s, cls: "status-default" };
}

function needsPoll(list: any[]) {
    return list.some((o) => o.status === "待处理" || o.status === "进行中");
}

async function pollPendingOrders() {
    const pending = orders.value.filter(
        (o) => o.status === "待处理" || o.status === "进行中",
    );
    for (const o of pending) {
        try {
            const fresh: any = await getOrderDetail(o.oid);
            const idx = orders.value.findIndex((x) => x.oid === o.oid);
            if (idx !== -1) orders.value[idx] = fresh;
        } catch {
            // ignore
        }
    }
    if (needsPoll(orders.value)) {
        pollTimer = setTimeout(pollPendingOrders, 5000);
    }
}

async function loadOrders() {
    try {
        const res: any = await getMyOrders();
        orders.value = Array.isArray(res) ? res : res?.list || [];
        if (needsPoll(orders.value)) {
            pollTimer = setTimeout(pollPendingOrders, 5000);
        }
    } catch {
        // ignore
    } finally {
        loading.value = false;
        refreshing.value = false;
    }
}

async function onRefresh() {
    refreshing.value = true;
    if (pollTimer) clearTimeout(pollTimer);
    await loadOrders();
}

onMounted(() => {
    if (route.query.paid === "1") {
        showToast({
            message: "支付成功，订单处理中",
            type: "success",
            duration: 3000,
        });
    }
    loadOrders();
});

onUnmounted(() => {
    if (pollTimer) clearTimeout(pollTimer);
});
</script>

<template>
    <div class="orders-page">
        <NavBar title="我的订单" />

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>

        <PullRefresh v-else v-model="refreshing" @refresh="onRefresh">
            <Empty
                v-if="!orders.length"
                description="暂无订单"
                style="padding-top: 80px"
            />
            <div v-else class="order-list">
                <div
                    v-for="o in orders"
                    :key="o.oid"
                    class="order-card animate-fade-in-up"
                >
                    <div class="order-header">
                        <span class="order-id">订单 #{{ o.oid }}</span>
                        <span
                            class="status-badge"
                            :class="getStatus(o.status).cls"
                        >
                            {{ getStatus(o.status).label }}
                        </span>
                    </div>
                    <div class="order-body">
                        <div class="order-row">
                            <span class="order-label">商品</span>
                            <span class="order-value">{{
                                o.class_name || o.kcname
                            }}</span>
                        </div>
                        <div class="order-row" v-if="o.retail_fees">
                            <span class="order-label">金额</span>
                            <span class="order-value price"
                                >¥{{ o.retail_fees }}</span
                            >
                        </div>
                        <div class="order-row" v-if="o.addtime">
                            <span class="order-label">时间</span>
                            <span class="order-value muted">{{
                                o.addtime
                            }}</span>
                        </div>
                    </div>
                    <div
                        class="order-progress"
                        v-if="
                            o.process !== undefined &&
                            o.process !== null &&
                            o.process !== ''
                        "
                    >
                        <div class="progress-bar">
                            <div
                                class="progress-fill"
                                :style="{
                                    width:
                                        Math.min(100, Number(o.process)) + '%',
                                }"
                            ></div>
                        </div>
                        <span class="progress-text">{{ o.process }}%</span>
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
.order-progress {
    padding: 8px 14px 12px;
    display: flex;
    align-items: center;
    gap: 8px;
}
.progress-bar {
    flex: 1;
    height: 4px;
    background: var(--border-color);
    border-radius: 2px;
    overflow: hidden;
}
.progress-fill {
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    transition: width 0.3s ease;
}
.progress-text {
    font-size: 12px;
    color: var(--text-muted);
    min-width: 32px;
    text-align: right;
}
</style>

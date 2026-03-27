<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Empty, NavBar, PullRefresh, showToast } from "vant";
import { getPromotionOrders, type PromotionOrderItem } from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";

const items = ref<PromotionOrderItem[]>([]);
const loading = ref(true);
const refreshing = ref(false);

async function loadOrders() {
    loading.value = true;
    try {
        const res = await getPromotionOrders();
        items.value = Array.isArray(res) ? res : res?.list || [];
    } catch (e: any) {
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

onMounted(loadOrders);
</script>

<template>
    <div class="promotion-page">
        <NavBar title="推广订单" left-arrow @click-left="router.push(`${basePath}/mine`)" />

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>

        <PullRefresh v-else v-model="refreshing" @refresh="onRefresh">
            <Empty v-if="!items.length" description="暂无推广订单" style="padding-top: 80px" />

            <div v-else class="promotion-list">
                <div v-for="item in items" :key="item.out_trade_no || item.id" class="promotion-card">
                    <div class="promotion-header">
                        <span class="promotion-title">{{ item.product_name || "商城商品" }}</span>
                        <span class="promotion-money">+¥{{ item.commission_amount }}</span>
                    </div>
                    <div class="promotion-row">
                        <span>购买账号</span>
                        <span>{{ item.buyer_account || "-" }}</span>
                    </div>
                    <div class="promotion-row" v-if="item.course_name">
                        <span>课程</span>
                        <span>{{ item.course_name }}</span>
                    </div>
                    <div class="promotion-row">
                        <span>订单金额</span>
                        <span>¥{{ item.money }}</span>
                    </div>
                    <div class="promotion-row">
                        <span>返利比例</span>
                        <span>{{ item.commission_rate }}%</span>
                    </div>
                    <div class="promotion-row">
                        <span>状态</span>
                        <span>{{ item.status_text }}</span>
                    </div>
                    <div class="promotion-time">{{ item.paytime || item.addtime }}</div>
                </div>
            </div>
        </PullRefresh>
    </div>
</template>

<style scoped>
.promotion-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 20px;
}
.promotion-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px;
}
.promotion-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 18px;
    padding: 14px;
    box-shadow: var(--shadow-md);
}
.promotion-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}
.promotion-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
}
.promotion-money {
    font-size: 18px;
    font-weight: 700;
    color: #16a34a;
}
.promotion-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-top: 10px;
    font-size: 13px;
    color: var(--text-secondary);
}
.promotion-time {
    margin-top: 12px;
    font-size: 12px;
    color: var(--text-muted);
}
</style>

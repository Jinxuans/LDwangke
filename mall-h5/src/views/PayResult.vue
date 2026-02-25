<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NavBar, Button, showToast } from "vant";
import { checkPay, confirmPay } from "../api";

const route = useRoute();
const router = useRouter();
const tid = route.params.tid as string;
const outTradeNo = (route.query.out_trade_no as string) || localStorage.getItem("pending_out_trade_no") || "";
const payUrl = localStorage.getItem("pending_pay_url") || "";

// 0=未支付 1=已支付待下单 2=已下单成功
const status = ref(0);
const checking = ref(false);
const confirming = ref(false);
const orderID = ref(0);
let pollTimer: ReturnType<typeof setTimeout> | null = null;
let pollCount = 0;

async function doCheck() {
    if (!outTradeNo) return;
    checking.value = true;
    try {
        const res: any = await checkPay(outTradeNo);
        status.value = res.status;
        orderID.value = res.order_id || 0;
        if (res.status === 2) {
            // 下单成功，清理缓存
            localStorage.removeItem("pending_out_trade_no");
            localStorage.removeItem("pending_pay_url");
        }
    } catch (e: any) {
        showToast(e?.message || "检测失败，请重试");
    } finally {
        checking.value = false;
    }
}

// 自动轮询：进入页面后每3秒检测一次，最多10次
function startAutoPoll() {
    if (pollCount >= 10 || status.value === 2) return;
    pollTimer = setTimeout(async () => {
        pollCount++;
        await doCheck();
        if (status.value !== 2) startAutoPoll();
    }, 3000);
}

onMounted(() => {
    if (outTradeNo) {
        doCheck();
        startAutoPoll();
    }
});

onUnmounted(() => {
    if (pollTimer) clearTimeout(pollTimer);
});

async function doConfirm() {
    if (!outTradeNo) return;
    confirming.value = true;
    try {
        const res: any = await confirmPay(outTradeNo);
        status.value = res.status;
        orderID.value = res.order_id || 0;
        if (res.status === 2) {
            localStorage.removeItem("pending_out_trade_no");
            localStorage.removeItem("pending_pay_url");
        }
    } catch (e: any) {
        showToast(e?.message || "提交失败，请重试");
    } finally {
        confirming.value = false;
    }
}

function goOrders() {
    router.push(`/${tid}/orders`);
}

function goPay() {
    if (payUrl) window.location.href = payUrl;
}
</script>

<template>
    <div class="pay-result-page">
        <NavBar title="支付结果" left-arrow @click-left="goOrders" />

        <div class="result-body animate-fade-in-up">
            <!-- 已下单成功 -->
            <template v-if="status === 2">
                <div class="icon-wrap success">
                    <van-icon name="checked" size="56" color="#10b981" />
                </div>
                <h2 class="result-title">下单成功</h2>
                <p class="result-desc">订单已提交，正在处理中</p>
                <div class="btn-group">
                    <Button type="primary" block round class="main-btn" @click="goOrders">
                        查看订单
                    </Button>
                </div>
            </template>

            <!-- 已支付，等待下单 -->
            <template v-else-if="status === 1">
                <div class="icon-wrap warning">
                    <van-icon name="clock-o" size="56" color="#f59e0b" />
                </div>
                <h2 class="result-title">支付成功</h2>
                <p class="result-desc">正在自动下单，或点击「我已支付」立即提交</p>
                <div class="btn-group">
                    <Button
                        type="primary"
                        block
                        round
                        :loading="confirming"
                        class="main-btn"
                        @click="doConfirm"
                    >
                        {{ confirming ? "提交中..." : "我已支付" }}
                    </Button>
                    <Button plain round block class="sub-btn" @click="goOrders">
                        查看订单列表
                    </Button>
                </div>
            </template>

            <!-- 未支付 -->
            <template v-else>
                <div class="icon-wrap pending">
                    <van-icon name="warning-o" size="56" color="#6366f1" />
                </div>
                <h2 class="result-title">等待支付</h2>
                <p class="result-desc">完成支付后点击「我已支付」提交订单</p>
                <div class="btn-group">
                    <Button
                        type="primary"
                        block
                        round
                        :loading="confirming"
                        class="main-btn"
                        @click="doConfirm"
                    >
                        {{ confirming ? "提交中..." : "我已支付" }}
                    </Button>
                    <Button
                        plain
                        round
                        block
                        :loading="checking"
                        class="sub-btn"
                        @click="doCheck"
                    >
                        {{ checking ? "检测中..." : "检测支付结果" }}
                    </Button>
                    <Button v-if="payUrl" plain round block class="sub-btn" @click="goPay">
                        重新支付
                    </Button>
                    <Button plain round block class="sub-btn" @click="goOrders">
                        查看订单列表
                    </Button>
                </div>
            </template>

            <p class="trade-no">订单号：{{ outTradeNo }}</p>
        </div>
    </div>
</template>

<style scoped>
.pay-result-page {
    min-height: 100vh;
    background: var(--bg-primary);
}
.result-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 48px 24px 32px;
    gap: 12px;
}
.icon-wrap {
    width: 88px;
    height: 88px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 8px;
}
.icon-wrap.success { background: var(--success-bg); }
.icon-wrap.warning { background: var(--warning-bg); }
.icon-wrap.pending { background: var(--primary-bg); }
.result-title {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
}
.result-desc {
    font-size: 14px;
    color: var(--text-muted);
    text-align: center;
}
.btn-group {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 16px;
}
.main-btn {
    height: 46px !important;
    font-size: 15px !important;
    font-weight: 600 !important;
}
.sub-btn {
    height: 42px !important;
    font-size: 14px !important;
    color: var(--text-secondary) !important;
    border-color: var(--border-color) !important;
}
.trade-no {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 16px;
    word-break: break-all;
    text-align: center;
}
</style>

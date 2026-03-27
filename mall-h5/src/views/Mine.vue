<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Button, Cell, CellGroup, NavBar, showToast } from "vant";
import {
    clearCUserSession,
    getCUserProfile,
    getGuestMallOrders,
    getMyProfile,
    getShopInfo,
    isCUserLoggedIn,
    triggerCustomerService,
    type MallProfile,
} from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";
const loggedIn = ref(isCUserLoggedIn(tid));
const loading = ref(false);
const shopConfig = ref<any>(null);
const profile = ref<MallProfile | null>(null);
const customerPopup = ref<null | { title: string; label: string; type: string; value: string }>(null);

const currentUser = computed(() => {
    if (profile.value) return profile.value;
    if (!loggedIn.value) return null;
    return getCUserProfile();
});

const guestOrderCount = computed(() => getGuestMallOrders(tid).length);
const promotionLink = computed(() => {
    const inviteCode = profile.value?.invite_code || currentUser.value?.invite_code || "";
    if (!inviteCode) return "";
    const path = tid ? `/mall/${tid}/` : "/mall/";
    return `${window.location.origin}${path}?sp=${encodeURIComponent(inviteCode)}`;
});

async function load() {
    loading.value = true;
    try {
        shopConfig.value = await getShopInfo();
        if (loggedIn.value) {
            profile.value = await getMyProfile();
        }
    } catch (e: any) {
        if (loggedIn.value) {
            clearCUserSession();
            loggedIn.value = false;
            profile.value = null;
        }
        if (e?.message) {
            showToast(e.message);
        }
    } finally {
        loading.value = false;
    }
}

function goLogin() {
    router.push(`${basePath}/login?redirect=${encodeURIComponent(`${basePath}/mine`)}`);
}

function goRegister() {
    router.push(`${basePath}/register?redirect=${encodeURIComponent(`${basePath}/mine`)}`);
}

function logout() {
    clearCUserSession();
    loggedIn.value = false;
    profile.value = null;
    showToast("已退出登录");
}

async function copyPromotionLink() {
    if (!promotionLink.value) {
        showToast("暂无推广链接");
        return;
    }
    await navigator.clipboard.writeText(promotionLink.value);
    showToast("推广链接已复制");
}

async function contactService() {
    const res = await triggerCustomerService(shopConfig.value?.mall_config?.customer_service);
    if ((res as any)?.mode === "popup") {
        customerPopup.value = res as any;
        return;
    }
    showToast((res as any)?.message || "操作失败");
}

async function copyCustomerValue() {
    if (!customerPopup.value?.value) return;
    await navigator.clipboard.writeText(customerPopup.value.value);
    showToast(customerPopup.value.type === "qq" ? "QQ号已复制" : customerPopup.value.type === "phone" ? "电话号码已复制" : "客服微信已复制");
}

function callCustomerPhone() {
    if (!customerPopup.value?.value) return;
    window.location.href = `tel:${customerPopup.value.value}`;
}

onMounted(load);
</script>

<template>
    <div class="mine-page">
        <NavBar title="我的" />

        <div class="hero-card" :class="{ compact: loading }">
            <template v-if="currentUser">
                <div class="hero-label">当前会员</div>
                <div class="hero-title">{{ (profile?.nickname || profile?.account) || currentUser.nickname || currentUser.account }}</div>
                <div class="hero-desc">登录账号：{{ profile?.account || currentUser.account }}</div>
                <template v-if="profile">
                <div class="wallet-row">
                    <div class="wallet-card">
                        <span class="wallet-label">可用佣金</span>
                        <span class="wallet-value">¥{{ profile.commission_money }}</span>
                    </div>
                    <div class="wallet-card">
                        <span class="wallet-label">冻结佣金</span>
                        <span class="wallet-value">¥{{ profile.commission_cdmoney }}</span>
                    </div>
                </div>
                <div class="wallet-row wallet-row-secondary">
                    <div class="wallet-card wallet-card-secondary">
                        <span class="wallet-label">累计返利</span>
                        <span class="wallet-value wallet-value-secondary">¥{{ profile.commission_total }}</span>
                    </div>
                    <div class="wallet-card wallet-card-secondary">
                        <span class="wallet-label">推广订单</span>
                        <span class="wallet-value wallet-value-secondary">{{ profile.promotion_orders }}</span>
                    </div>
                </div>
                <div class="hero-meta">
                    <span>推广码：{{ profile.invite_code || "未生成" }}</span>
                    <span v-if="profile.referrer_account">上级：{{ profile.referrer_nickname || profile.referrer_account }}</span>
                </div>
                </template>
                <div class="hero-actions">
                    <Button type="primary" round @click="router.push(`${basePath}/orders`)">看订单</Button>
                    <Button plain round @click="router.push(`${basePath}/promotion`)">推广订单</Button>
                </div>
                <div class="hero-actions hero-actions-secondary">
                    <Button type="success" round @click="router.push(`${basePath}/withdraw`)">佣金提现</Button>
                </div>
            </template>
            <template v-else>
                <div class="hero-label">未登录</div>
                <div class="hero-title">登录后可查看订单和推广收益</div>
                <div class="hero-desc">
                    当前设备已保存 {{ guestOrderCount }} 条访客支付订单，课程处理进度可直接用下单账号查询。
                </div>
                <div class="hero-actions">
                    <Button type="primary" round @click="goLogin">会员登录</Button>
                    <Button plain round @click="goRegister">
                        {{ shopConfig?.mall_config?.register_enabled ? "会员注册" : "申请开通账号" }}
                    </Button>
                </div>
            </template>
        </div>

        <div v-if="profile?.promotion_enabled" class="promo-panel">
            <div class="promo-head">
                <div>
                    <div class="promo-title">推广商城链接</div>
                    <div class="promo-desc">分享给其他人下单，成功成单后按 {{ profile.commission_rate }}% 返利。</div>
                </div>
                <Button size="small" type="primary" round @click="copyPromotionLink">复制链接</Button>
            </div>
            <div class="promo-link">{{ promotionLink }}</div>
        </div>

        <CellGroup inset class="menu-group">
            <Cell title="订单" label="查看支付订单记录" is-link @click="router.push(`${basePath}/orders`)" />
            <Cell title="查进度" label="按下单账号查询进行中的课程" is-link @click="router.push(`${basePath}/query`)" />
            <Cell
                v-if="shopConfig?.mall_config?.customer_service?.enabled"
                :title="shopConfig?.mall_config?.customer_service?.label || '联系客服'"
                label="支付、进度、售后问题可直接联系店铺客服"
                is-link
                @click="contactService"
            />
            <Cell
                v-if="profile?.promotion_enabled"
                title="推广订单"
                label="查看你的推广返利订单"
                is-link
                @click="router.push(`${basePath}/promotion`)"
            />
            <Cell
                v-if="profile?.promotion_enabled"
                title="佣金提现"
                label="提现你的推广佣金"
                is-link
                @click="router.push(`${basePath}/withdraw`)"
            />
            <Cell title="首页" label="返回商城商品列表" is-link @click="router.push(basePath || '/')" />
            <Cell v-if="loggedIn" title="退出登录" label="清除当前会员会话" is-link @click="logout" />
        </CellGroup>

        <div v-if="customerPopup" class="notice-popup-mask" @click.self="customerPopup = null">
            <div class="notice-popup animate-fade-in-up">
                <div class="notice-popup-header">
                    <span>{{ customerPopup.title }}</span>
                    <button class="notice-popup-close" @click="customerPopup = null">×</button>
                </div>
                <div class="notice-popup-body">
                    <div class="customer-value">{{ customerPopup.value }}</div>
                    <div class="customer-tip">若无法直接跳转，可手动复制后联系。</div>
                </div>
                <div class="customer-actions">
                    <Button plain round block @click="copyCustomerValue">复制{{ customerPopup.type === 'phone' ? '号码' : '账号' }}</Button>
                    <Button
                        v-if="customerPopup.type === 'phone'"
                        type="primary"
                        round
                        block
                        @click="callCustomerPhone"
                    >
                        立即拨号
                    </Button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.mine-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 60px;
}
.hero-card {
    margin: 14px 12px 0;
    padding: 16px 14px;
    border-radius: 16px;
    background: linear-gradient(135deg, #fff7ed 0%, #ffffff 55%, #eff6ff 100%);
    border: 1px solid #fde4c7;
    box-shadow: var(--shadow-lg);
}
.hero-label {
    font-size: 11px;
    color: #9a3412;
    font-weight: 700;
}
.hero-title {
    margin-top: 6px;
    font-size: 18px;
    line-height: 1.3;
    color: var(--text-primary);
    font-weight: 700;
}
.hero-desc {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--text-secondary);
}
.hero-actions {
    display: flex;
    gap: 10px;
    margin-top: 12px;
}
.hero-actions-secondary {
    margin-top: 8px;
}
.hero-actions :deep(.van-button) {
    flex: 1;
    height: 36px;
}
.wallet-row {
    display: flex;
    gap: 10px;
    margin-top: 12px;
}
.wallet-row-secondary {
    margin-top: 8px;
}
.wallet-card {
    flex: 1;
    padding: 10px 10px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.82);
    border: 1px solid rgba(251, 191, 36, 0.25);
}
.wallet-card-secondary {
    background: rgba(255, 255, 255, 0.68);
    border-color: rgba(148, 163, 184, 0.25);
}
.wallet-label {
    display: block;
    font-size: 11px;
    color: var(--text-muted);
}
.wallet-value {
    display: block;
    margin-top: 6px;
    font-size: 18px;
    font-weight: 700;
    color: #c2410c;
}
.wallet-value-secondary {
    color: #1e293b;
}
.hero-meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 10px;
    font-size: 11px;
    color: var(--text-secondary);
}
.promo-panel {
    margin: 14px 12px 0;
    padding: 12px;
    border-radius: 14px;
    background: #fff;
    border: 1px solid var(--border-color);
    box-shadow: var(--shadow-md);
}
.promo-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}
.promo-title {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-primary);
}
.promo-desc {
    margin-top: 4px;
    font-size: 11px;
    line-height: 1.5;
    color: var(--text-muted);
}
.promo-link {
    margin-top: 10px;
    padding: 10px;
    border-radius: 12px;
    background: #f8fafc;
    font-size: 11px;
    line-height: 1.5;
    color: var(--text-secondary);
    word-break: break-all;
}
.menu-group {
    margin-top: 12px;
}
.notice-popup-mask {
    position: fixed;
    inset: 0;
    z-index: 50;
    background: rgba(15, 23, 42, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
}
.notice-popup {
    width: min(100%, 420px);
    max-height: 80vh;
    overflow: hidden;
    border-radius: 20px;
    background: #fff;
    box-shadow: 0 20px 50px rgba(15, 23, 42, 0.24);
    display: flex;
    flex-direction: column;
}
.notice-popup-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 18px 12px;
    border-bottom: 1px solid #e2e8f0;
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
}
.notice-popup-close {
    border: none;
    background: transparent;
    font-size: 24px;
    line-height: 1;
    color: #64748b;
}
.notice-popup-body {
    padding: 16px 18px;
    overflow: auto;
    font-size: 14px;
    line-height: 1.8;
    color: #334155;
}
.customer-value {
    padding: 12px 14px;
    border-radius: 14px;
    background: #f8fafc;
    color: #0f172a;
    font-size: 15px;
    font-weight: 700;
    word-break: break-all;
}
.customer-tip {
    margin-top: 10px;
    font-size: 12px;
    color: #64748b;
}
.customer-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 0 18px 18px;
}
</style>

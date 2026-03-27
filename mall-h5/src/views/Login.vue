<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NavBar, Button, CellGroup, Field, Form, showToast } from "vant";
import {
    getGuestMallOrders,
    isCUserLoggedIn,
    loginCUser,
    removeGuestMallOrders,
    saveCUserSession,
} from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";
const account = ref("");
const password = ref("");
const submitting = ref(false);

const redirectPath = computed(() => {
    const raw = route.query.redirect;
    return typeof raw === "string" && raw ? raw : `${basePath}/orders`;
});

onMounted(() => {
    if (isCUserLoggedIn(tid)) {
        router.replace(redirectPath.value);
    }
});

function goBack() {
    if (window.history.length > 1) {
        router.back();
        return;
    }
    router.push(basePath || "/");
}

async function handleSubmit() {
    if (!account.value.trim()) {
        showToast("请填写会员账号");
        return;
    }
    if (!password.value.trim()) {
        showToast("请填写会员密码");
        return;
    }
    submitting.value = true;
    try {
        const guestOrders = getGuestMallOrders(tid).map((item) => ({
            out_trade_no: item.outTradeNo,
            access_token: item.accessToken,
        }));
        const res: any = await loginCUser({
            account: account.value.trim(),
            password: password.value.trim(),
            guest_orders: guestOrders,
        });
        saveCUserSession({
            token: res.token,
            id: res.id,
            account: res.account,
            nickname: res.nickname,
            tid,
            invite_code: res.invite_code,
        });
        const mergedOrders = Array.isArray(res?.merged_guest_orders)
            ? res.merged_guest_orders.filter(Boolean)
            : [];
        if (mergedOrders.length > 0) {
            removeGuestMallOrders(tid, mergedOrders);
        }
        showToast({
            type: "success",
            message:
                mergedOrders.length > 0
                    ? `登录成功，已合并 ${mergedOrders.length} 笔本机订单`
                    : "登录成功",
            duration: 1500,
        });
        setTimeout(() => {
            router.replace(redirectPath.value);
        }, 200);
    } catch (e: any) {
        showToast(e?.message || "登录失败");
    } finally {
        submitting.value = false;
    }
}
</script>

<template>
    <div class="login-page">
        <NavBar title="会员登录" left-arrow @click-left="goBack" />

        <div class="login-hero animate-fade-in-down">
            <div class="login-badge">MEMBER</div>
            <h1 class="login-title">登录后查看专属订单</h1>
            <p class="login-desc">使用商家为你开通的会员账号登录，订单记录将按会员隔离。</p>
        </div>

        <div class="login-card animate-fade-in-up">
            <Form @submit="handleSubmit">
                <CellGroup inset>
                    <Field
                        v-model="account"
                        label="账号"
                        placeholder="请输入会员账号"
                        clearable
                    />
                    <Field
                        v-model="password"
                        label="密码"
                        type="password"
                        placeholder="请输入会员密码"
                        clearable
                    />
                </CellGroup>
                <div class="login-actions">
                    <Button
                        type="primary"
                        block
                        round
                        native-type="submit"
                        :loading="submitting"
                    >
                        {{ submitting ? "登录中..." : "立即登录" }}
                    </Button>
                    <Button plain block round @click="router.push(`${basePath}/query`)">
                        先去查订单
                    </Button>
                    <Button plain block round @click="router.push(`${basePath}/register`)">
                        去注册会员
                    </Button>
                </div>
            </Form>
        </div>
    </div>
</template>

<style scoped>
.login-page {
    min-height: 100vh;
    background:
        radial-gradient(circle at top, rgba(99, 102, 241, 0.14), transparent 36%),
        linear-gradient(180deg, #f8fafc 0%, #eef2ff 100%);
}
.login-hero {
    padding: 28px 20px 18px;
}
.login-badge {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: var(--radius-full);
    background: rgba(99, 102, 241, 0.12);
    color: var(--primary-dark);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
}
.login-title {
    margin-top: 14px;
    font-size: 26px;
    line-height: 1.2;
    font-weight: 700;
    color: var(--text-primary);
}
.login-desc {
    margin-top: 10px;
    font-size: 14px;
    color: var(--text-secondary);
}
.login-card {
    margin: 0 14px;
    padding: 18px 0 20px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.88);
    box-shadow: 0 18px 40px rgba(15, 23, 42, 0.08);
    backdrop-filter: blur(12px);
}
.login-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 18px 16px 0;
}
</style>

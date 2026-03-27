<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NavBar, Button, CellGroup, Field, Form, showToast } from "vant";
import {
    getMallPromoterCode,
    isCUserLoggedIn,
    registerCUser,
    saveCUserSession,
} from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";
const account = ref("");
const password = ref("");
const nickname = ref("");
const phone = ref("");
const submitting = ref(false);

const redirectPath = computed(() => {
    const raw = route.query.redirect;
    return typeof raw === "string" && raw ? raw : `${basePath}/mine`;
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
        showToast("请填写登录密码");
        return;
    }
    submitting.value = true;
    try {
        const res: any = await registerCUser({
            account: account.value.trim(),
            password: password.value.trim(),
            nickname: nickname.value.trim(),
            phone: phone.value.trim(),
            promoter_code: getMallPromoterCode(tid),
        });
        saveCUserSession({
            token: res.token,
            id: res.id,
            account: res.account,
            nickname: res.nickname,
            tid,
            invite_code: res.invite_code,
        });
        showToast({ type: "success", message: "注册成功", duration: 1200 });
        setTimeout(() => router.replace(redirectPath.value), 200);
    } catch (e: any) {
        showToast(e?.message || "注册失败");
    } finally {
        submitting.value = false;
    }
}
</script>

<template>
    <div class="register-page">
        <NavBar title="会员注册" left-arrow @click-left="goBack" />

        <div class="register-hero animate-fade-in-down">
            <div class="register-badge">JOIN</div>
            <h1 class="register-title">注册后可同步订单和推广收益</h1>
            <p class="register-desc">创建商城会员账号后，可查看自己的支付订单、推广链接和返利记录。</p>
        </div>

        <div class="register-card animate-fade-in-up">
            <Form @submit="handleSubmit">
                <CellGroup inset>
                    <Field v-model="account" label="账号" placeholder="请输入会员账号" clearable />
                    <Field v-model="password" label="密码" type="password" placeholder="请输入登录密码" clearable />
                    <Field v-model="nickname" label="昵称" placeholder="选填，不填则默认账号" clearable />
                    <Field v-model="phone" label="手机号" placeholder="选填" clearable />
                </CellGroup>
                <div class="register-actions">
                    <Button type="primary" block round native-type="submit" :loading="submitting">
                        {{ submitting ? "注册中..." : "立即注册" }}
                    </Button>
                    <Button plain block round @click="router.push(`${basePath}/login`)">
                        已有账号，去登录
                    </Button>
                </div>
            </Form>
        </div>
    </div>
</template>

<style scoped>
.register-page {
    min-height: 100vh;
    background:
        radial-gradient(circle at top, rgba(249, 115, 22, 0.14), transparent 36%),
        linear-gradient(180deg, #fff7ed 0%, #f8fafc 100%);
}
.register-hero {
    padding: 28px 20px 18px;
}
.register-badge {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: var(--radius-full);
    background: rgba(249, 115, 22, 0.14);
    color: #c2410c;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
}
.register-title {
    margin-top: 14px;
    font-size: 26px;
    line-height: 1.2;
    font-weight: 700;
    color: var(--text-primary);
}
.register-desc {
    margin-top: 10px;
    font-size: 14px;
    color: var(--text-secondary);
}
.register-card {
    margin: 0 14px;
    padding: 18px 0 20px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.92);
    box-shadow: 0 18px 40px rgba(15, 23, 42, 0.08);
}
.register-actions {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 18px 16px 0;
}
</style>

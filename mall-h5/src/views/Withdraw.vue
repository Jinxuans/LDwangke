<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Button, CellGroup, Field, NavBar, Popup, PullRefresh, showToast } from "vant";
import { createCommissionWithdraw, getCommissionWithdrawRequests, getMyProfile, type MallProfile, type MallWithdrawItem } from "../api";

const route = useRoute();
const router = useRouter();
const tid = String(route.params.tid || "");
const basePath = tid ? `/${tid}` : "";

const loading = ref(true);
const refreshing = ref(false);
const visible = ref(false);
const submitting = ref(false);
const profile = ref<MallProfile | null>(null);
const list = ref<MallWithdrawItem[]>([]);

const form = reactive({
    amount: "",
    method: "manual",
    account_name: "",
    account_no: "",
    bank_name: "",
    note: "",
});

const available = computed(() => Number(profile.value?.commission_money || 0));
const frozen = computed(() => Number(profile.value?.commission_cdmoney || 0));

function statusText(status: number) {
    if (status === 1) return "已通过";
    if (status === -1) return "已驳回";
    return "待审核";
}

async function loadData() {
    loading.value = true;
    try {
        profile.value = await getMyProfile();
        const res = await getCommissionWithdrawRequests();
        list.value = Array.isArray(res) ? res : res?.list || [];
    } catch (e: any) {
        showToast(e?.message || "加载失败");
    } finally {
        loading.value = false;
        refreshing.value = false;
    }
}

async function onRefresh() {
    refreshing.value = true;
    await loadData();
}

async function submit() {
    const amount = Number(form.amount || 0);
    if (!amount || amount <= 0) {
        showToast("请输入有效提现金额");
        return;
    }
    if (!form.account_name.trim()) {
        showToast("请填写收款人");
        return;
    }
    if (!form.account_no.trim()) {
        showToast("请填写收款账号");
        return;
    }
    submitting.value = true;
    try {
        await createCommissionWithdraw({
            amount,
            method: form.method,
            account_name: form.account_name.trim(),
            account_no: form.account_no.trim(),
            bank_name: form.bank_name.trim(),
            note: form.note.trim(),
        });
        showToast({ type: "success", message: "提现申请已提交", duration: 1200 });
        visible.value = false;
        form.amount = "";
        form.account_name = "";
        form.account_no = "";
        form.bank_name = "";
        form.note = "";
        await loadData();
    } catch (e: any) {
        showToast(e?.message || "提交失败");
    } finally {
        submitting.value = false;
    }
}

async function fillAll() {
    form.amount = available.value.toFixed(2);
}

onMounted(loadData);
</script>

<template>
    <div class="withdraw-page">
        <NavBar title="佣金提现" left-arrow @click-left="router.push(`${basePath}/mine`)" />

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>

        <PullRefresh v-else v-model="refreshing" @refresh="onRefresh">
            <div class="wallet-board">
                <div class="wallet-card">
                    <span class="wallet-label">可提现佣金</span>
                    <span class="wallet-value">¥{{ available.toFixed(2) }}</span>
                </div>
                <div class="wallet-card">
                    <span class="wallet-label">冻结佣金</span>
                    <span class="wallet-value">¥{{ frozen.toFixed(2) }}</span>
                </div>
            </div>

            <div class="action-box">
                <Button type="primary" round block @click="visible = true">申请提现</Button>
            </div>

            <div class="section-title">提现记录</div>
            <div v-if="!list.length" class="empty-box">暂无提现记录</div>
            <div v-else class="record-list">
                <div v-for="item in list" :key="item.id" class="record-card">
                    <div class="record-head">
                        <span class="record-amount">¥{{ Number(item.amount).toFixed(2) }}</span>
                        <span class="record-status">{{ statusText(item.status) }}</span>
                    </div>
                    <div class="record-row">
                        <span>收款信息</span>
                        <span>{{ item.account_name }} / {{ item.account_no }}</span>
                    </div>
                    <div class="record-row">
                        <span>方式</span>
                        <span>{{ item.bank_name || item.method || "-" }}</span>
                    </div>
                    <div class="record-row" v-if="item.audit_remark">
                        <span>审核备注</span>
                        <span>{{ item.audit_remark }}</span>
                    </div>
                    <div class="record-time">{{ item.addtime }}</div>
                </div>
            </div>
        </PullRefresh>

        <Popup v-model:show="visible" round position="bottom" :style="{ padding: '18px 16px 28px' }">
            <div class="popup-title">申请提现</div>
            <CellGroup inset>
                <Field v-model="form.amount" label="金额" type="number" placeholder="请输入提现金额">
                    <template #button>
                        <Button size="small" plain round @click="fillAll">全部提现</Button>
                    </template>
                </Field>
                <Field v-model="form.account_name" label="收款人" placeholder="请输入收款人姓名" />
                <Field v-model="form.account_no" label="账号" placeholder="请输入收款账号" />
                <Field v-model="form.bank_name" label="开户行" placeholder="选填，如支付宝/银行卡" />
                <Field v-model="form.note" label="备注" type="textarea" rows="2" placeholder="选填" />
            </CellGroup>
            <div class="popup-actions">
                <Button block round @click="visible = false">取消</Button>
                <Button type="primary" block round :loading="submitting" @click="submit">提交申请</Button>
            </div>
        </Popup>
    </div>
</template>

<style scoped>
.withdraw-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 20px;
}
.wallet-board {
    display: flex;
    gap: 12px;
    padding: 14px 12px 0;
}
.wallet-card {
    flex: 1;
    padding: 16px 14px;
    border-radius: 18px;
    background: linear-gradient(145deg, #fff 0%, #fef3c7 100%);
    border: 1px solid #fde68a;
}
.wallet-label {
    display: block;
    font-size: 12px;
    color: var(--text-muted);
}
.wallet-value {
    display: block;
    margin-top: 10px;
    font-size: 24px;
    font-weight: 700;
    color: #b45309;
}
.action-box {
    padding: 14px 12px 0;
}
.section-title {
    padding: 18px 14px 8px;
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
}
.empty-box {
    margin: 0 12px;
    padding: 32px 14px;
    border-radius: 16px;
    background: #fff;
    text-align: center;
    color: var(--text-muted);
}
.record-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 0 12px 12px;
}
.record-card {
    padding: 14px;
    border-radius: 18px;
    background: #fff;
    border: 1px solid var(--border-color);
}
.record-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
}
.record-amount {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
}
.record-status {
    font-size: 12px;
    color: #b45309;
}
.record-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-top: 10px;
    font-size: 13px;
    color: var(--text-secondary);
}
.record-time {
    margin-top: 12px;
    font-size: 12px;
    color: var(--text-muted);
}
.popup-title {
    margin-bottom: 14px;
    text-align: center;
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
}
.popup-actions {
    display: flex;
    gap: 12px;
    margin-top: 16px;
}
.popup-actions :deep(.van-button) {
    flex: 1;
}
</style>

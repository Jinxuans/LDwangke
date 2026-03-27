<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { NavBar, Field, Button, CellGroup, Empty, showToast } from "vant";
import { getShopInfo, queryOrder, triggerCustomerService } from "../api";

const route = useRoute();
const keyword = ref("");
const loading = ref(false);
const results = ref<any[]>([]);
const searched = ref(false);
const shopConfig = ref<any>(null);
const customerPopup = ref<null | { title: string; label: string; type: string; value: string }>(null);

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

async function handleSearch() {
    if (!keyword.value.trim()) return;
    loading.value = true;
    searched.value = false;
    try {
        const res: any = await queryOrder(keyword.value.trim());
        results.value = Array.isArray(res) ? res : res?.list || [];
    } catch (e: any) {
        results.value = [];
        showToast(e?.message || "查询失败");
    } finally {
        loading.value = false;
        searched.value = true;
    }
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

onMounted(() => {
    const q = String(route.query.keyword || "").trim();
    void getShopInfo().then((res) => {
        shopConfig.value = res;
    }).catch(() => {});
    if (!q) return;
    keyword.value = q;
    handleSearch();
});
</script>

<template>
    <div class="query-page">
        <NavBar title="查进度" />

        <div class="search-section">
            <CellGroup inset>
                <Field
                    v-model="keyword"
                    placeholder="输入下单账号，查询正在处理的课程"
                    clearable
                    left-icon="search"
                    @keyup.enter="handleSearch"
                />
            </CellGroup>
            <div class="search-btn-wrap">
                <Button
                    type="primary"
                    block
                    round
                    :loading="loading"
                    class="search-btn"
                    @click="handleSearch"
                >
                    {{ loading ? "查询中..." : "查进度" }}
                </Button>
            </div>
        </div>

        <div class="divider"></div>

        <div v-if="shopConfig?.mall_config?.customer_service?.enabled" class="service-card">
            <div>
                <div class="service-title">查询不到或进度异常？</div>
                <div class="service-desc">可以直接联系店铺客服处理订单问题。</div>
            </div>
            <Button size="small" round type="primary" class="service-btn" @click="contactService">
                {{ shopConfig?.mall_config?.customer_service?.label || "联系客服" }}
            </Button>
        </div>

        <div class="divider"></div>

        <div v-if="!searched && !loading" class="hint-state">
            <van-icon name="search" size="40" color="#d1d5db" />
            <p>输入下单账号开始查询课程进度</p>
        </div>

        <div v-else-if="searched">
            <Empty
                v-if="!results.length"
                description="未找到正在处理的课程"
                style="padding-top: 60px"
            />
            <div v-else class="result-list">
                <div class="result-count">找到 {{ results.length }} 条记录</div>
                <div
                    v-for="o in results"
                    :key="o.oid"
                    class="order-card animate-fade-in-up"
                >
                    <div class="order-header">
                        <span class="order-id">业务单 #{{ o.oid }}</span>
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
                        <div class="order-row">
                            <span class="order-label">账号</span>
                            <span class="order-value">{{
                                o.account || o.user
                            }}</span>
                        </div>
                        <div class="order-row" v-if="o.kcname">
                            <span class="order-label">课程</span>
                            <span class="order-value">{{ o.kcname }}</span>
                        </div>
                        <div class="order-row" v-if="o.addtime">
                            <span class="order-label">时间</span>
                            <span class="order-value muted">{{
                                o.addtime
                            }}</span>
                        </div>
                        <div class="order-row" v-if="o.remarks">
                            <span class="order-label">日志</span>
                            <span class="order-value">{{ o.remarks }}</span>
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
        </div>

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
.query-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 60px;
}
.search-section {
    padding: 16px 0 0;
    background: var(--bg-secondary);
}
.search-btn-wrap {
    padding: 12px 16px 16px;
}
.search-btn {
    height: 44px !important;
    font-size: 15px !important;
    font-weight: 600 !important;
}
.divider {
    height: 8px;
    background: var(--bg-primary);
}
.service-card {
    margin: 12px 12px 0;
    padding: 14px;
    border-radius: 16px;
    border: 1px solid #bfdbfe;
    background: linear-gradient(135deg, #eff6ff 0%, #ffffff 100%);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}
.service-title {
    font-size: 14px;
    font-weight: 700;
    color: #1e3a8a;
}
.service-desc {
    margin-top: 4px;
    font-size: 12px;
    color: #64748b;
}
.service-btn {
    flex-shrink: 0;
}
.hint-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 80px 20px;
    color: var(--text-muted);
    font-size: 14px;
}
.result-list {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
}
.result-count {
    font-size: 12px;
    color: var(--text-muted);
    padding: 0 2px 4px;
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

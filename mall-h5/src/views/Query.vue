<script setup lang="ts">
import { ref } from "vue";
import { NavBar, Field, Button, CellGroup, Empty } from "vant";
import { queryOrder } from "../api";

const keyword = ref("");
const loading = ref(false);
const results = ref<any[]>([]);
const searched = ref(false);

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
    } catch (e) {
        results.value = [];
    } finally {
        loading.value = false;
        searched.value = true;
    }
}
</script>

<template>
    <div class="query-page">
        <NavBar title="查询订单" />

        <div class="search-section">
            <CellGroup inset>
                <Field
                    v-model="keyword"
                    placeholder="输入账号或订单号查询"
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
                    {{ loading ? "查询中..." : "查询" }}
                </Button>
            </div>
        </div>

        <div class="divider"></div>

        <div v-if="!searched && !loading" class="hint-state">
            <van-icon name="search" size="40" color="#d1d5db" />
            <p>输入账号或订单号开始查询</p>
        </div>

        <div v-else-if="searched">
            <Empty
                v-if="!results.length"
                description="未找到相关订单"
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
                        <div class="order-row">
                            <span class="order-label">账号</span>
                            <span class="order-value">{{
                                o.account || o.user
                            }}</span>
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
</style>

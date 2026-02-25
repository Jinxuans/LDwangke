<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { NavBar, NoticeBar, Empty } from "vant";
import { getShopInfo, getProducts } from "../api";

const route = useRoute();
const router = useRouter();
const tid = route.params.tid as string;
const shop = ref<any>(null);
const products = ref<any[]>([]);
const loading = ref(true);
const activeCategory = ref("全部");

// Extract unique categories from products
const categories = computed(() => {
    const cats = new Set<string>();
    products.value.forEach(p => {
        if (p.fenlei_name) cats.add(p.fenlei_name);
    });
    return ["全部", ...Array.from(cats)];
});

// Filter products by selected category
const filteredProducts = computed(() => {
    if (activeCategory.value === "全部") return products.value;
    return products.value.filter(p => p.fenlei_name === activeCategory.value);
});

// Generate a color based on name
function getAvatarColor(name: string): string {
    const colors = [
        "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
        "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
        "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
        "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
        "linear-gradient(135deg, #fa709a 0%, #fee140 100%)",
        "linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)",
        "linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)",
        "linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)",
    ];
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length]!;
}

onMounted(async () => {
    try {
        const [shopRes, prodRes] = await Promise.all([
            getShopInfo(),
            getProducts(),
        ]);
        shop.value = shopRes;
        products.value = Array.isArray(prodRes) ? prodRes : [];
    } catch (e) {
        console.error(e);
    } finally {
        loading.value = false;
    }
});
</script>

<template>
    <div class="home-page">
        <NavBar :title="shop?.shop_name || '精选商城'" class="home-nav">
            <template #right>
                <span
                    class="nav-btn"
                    @click="router.push(`/${tid}/query`)"
                >
                    <van-icon name="search" size="20" />
                </span>
            </template>
        </NavBar>

        <div class="shop-bar animate-fade-in-down">
            <van-image
                round
                width="40"
                height="40"
                :src="
                    shop?.shop_logo ||
                    'https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg'
                "
            />
            <div class="shop-info">
                <span class="shop-name">{{ shop?.shop_name || "精选商城" }}</span>
                <span class="shop-sub">好物优选 · 品质保证</span>
            </div>
            <van-button
                size="mini"
                plain
                type="primary"
                round
                @click="router.push(`/${tid}/query`)"
                >查订单</van-button
            >
        </div>

        <NoticeBar
            v-if="shop?.shop_desc"
            :text="shop.shop_desc"
            left-icon="volume-o"
            scrollable
        />

        <!-- Category Tabs -->
        <div v-if="products.length > 0" class="category-tabs">
            <div class="category-scroll">
                <span
                    v-for="cat in categories"
                    :key="cat"
                    class="category-tab"
                    :class="{ active: activeCategory === cat }"
                    @click="activeCategory = cat"
                >
                    {{ cat }}
                </span>
            </div>
        </div>

        <div class="section-header">
            <span class="section-title">{{ activeCategory === '全部' ? '全部商品' : activeCategory }}</span>
            <span class="section-count" v-if="filteredProducts.length"
                >共 {{ filteredProducts.length }} 件</span
            >
        </div>

        <div v-if="loading" class="loading-container">
            <div class="loading-spinner"></div>
            <p class="loading-text">加载中...</p>
        </div>
        <Empty
            v-else-if="!filteredProducts.length"
            :description="activeCategory === '全部' ? '暂无商品' : '该分类暂无商品'"
            style="padding-top: 80px"
        />
        <div v-else class="product-grid">
            <div
                v-for="(p, i) in filteredProducts"
                :key="p.cid"
                class="product-card animate-fade-in-up"
                :style="{ animationDelay: `${i * 0.05}s` }"
                @click="router.push(`/${tid}/product/${p.cid}`)"
            >
                <div class="card-img-wrap">
                    <div 
                        class="product-placeholder"
                        :style="{ background: getAvatarColor(p.name || '商品') }"
                    >
                        <span class="placeholder-char">{{ (p.name || '商')[0] }}</span>
                    </div>
                </div>
                <div class="card-body">
                    <p class="card-name">{{ p.name }}</p>
                    <div class="card-footer">
                        <span class="price-tag">¥{{ p.retail_price }}</span>
                        <van-icon name="arrow" size="12" color="#94a3b8" />
                    </div>
                </div>
            </div>
        </div>
        <div class="page-bottom safe-area-bottom"></div>
    </div>
</template>

<style scoped>
.home-page {
    min-height: 100vh;
    background: var(--bg-primary);
    padding-bottom: 60px;
}
.home-nav {
    background: var(--bg-secondary) !important;
}
.nav-btn {
    display: flex;
    align-items: center;
    padding: 6px;
    color: var(--text-secondary);
    cursor: pointer;
}
.nav-btn:active {
    color: var(--primary-color);
}
.shop-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
}
.shop-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
}
.shop-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.shop-sub {
    font-size: 12px;
    color: var(--text-muted);
}
.category-tabs {
    background: var(--bg-secondary);
    padding: 8px 0;
    border-bottom: 1px solid var(--border-light);
}
.category-scroll {
    display: flex;
    gap: 8px;
    padding: 0 12px;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
}
.category-scroll::-webkit-scrollbar {
    display: none;
}
.category-tab {
    flex-shrink: 0;
    padding: 6px 14px;
    font-size: 13px;
    color: var(--text-secondary);
    background: var(--bg-primary);
    border-radius: var(--radius-full);
    cursor: pointer;
    transition: all 0.2s ease;
    border: 1px solid var(--border-color);
}
.category-tab.active {
    background: var(--primary-color);
    color: #fff;
    border-color: var(--primary-color);
}
.section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 10px;
}
.section-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
}
.section-count {
    font-size: 12px;
    color: var(--text-muted);
}
.product-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    padding: 0 12px;
}
.product-card {
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    overflow: hidden;
    border: 1px solid var(--border-color);
    cursor: pointer;
    transition:
        box-shadow 0.2s ease,
        transform 0.15s ease;
}
.product-card:active {
    transform: scale(0.97);
    box-shadow: var(--shadow-md);
}
.card-img-wrap {
    overflow: hidden;
    background: var(--bg-primary);
}
.product-placeholder {
    width: 100%;
    height: 150px;
    display: flex;
    align-items: center;
    justify-content: center;
}
.placeholder-char {
    font-size: 48px;
    font-weight: 700;
    color: rgba(255, 255, 255, 0.9);
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.card-body {
    padding: 10px 10px 12px;
}
.card-name {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.4;
    margin-bottom: 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}
.card-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.page-bottom {
    height: 24px;
}
</style>

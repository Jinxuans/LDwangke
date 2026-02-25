import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
    history: createWebHistory("/mall"),
    routes: [
        {
            path: "/:tid",
            component: () => import("../views/Layout.vue"),
            children: [
                {
                    path: "",
                    name: "Home",
                    component: () => import("../views/Home.vue"),
                },
                {
                    path: "product/:cid",
                    name: "Product",
                    component: () => import("../views/Product.vue"),
                },

                {
                    path: "orders",
                    name: "Orders",
                    component: () => import("../views/Orders.vue"),
                },
                {
                    path: "query",
                    name: "Query",
                    component: () => import("../views/Query.vue"),
                },
                {
                    path: "pay-result",
                    name: "PayResult",
                    component: () => import("../views/PayResult.vue"),
                },
            ],
        },
        {
            path: "/:pathMatch(.*)*",
            component: () => import("../views/NotFound.vue"),
        },
    ],
});

export default router;

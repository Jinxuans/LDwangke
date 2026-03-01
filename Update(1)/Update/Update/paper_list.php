<?php
$title = '论文管理';
require_once('head.php'); ?>
<link href="css/element.css" rel="stylesheet">
<link href="css/tailwind.css" rel="stylesheet">
<div class="app-content-body" style="padding: 15px;" id="PaperList">
    <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="论文名称">
            <el-input v-model="searchForm.title" placeholder="请输入论文名称" size="small" style="width: 160px;"></el-input>
        </el-form-item>
        <el-form-item label="商品名称">
            <el-input v-model="searchForm.shopname" placeholder="请输入商品名称" size="small" style="width: 160px;"></el-input>
        </el-form-item>
        <el-form-item label="学生姓名">
            <el-input v-model="searchForm.studentName" placeholder="请输入学生姓名" size="small" style="width: 160px;"></el-input>
        </el-form-item>
        <el-form-item label="状态">
            <el-select v-model="searchForm.state" placeholder="请选择状态" size="small" style="width: 120px;">
                <el-option label="全部" value=""></el-option>
                <el-option label="待处理" value="0"></el-option>
                <el-option label="正在处理" value="1"></el-option>
                <el-option label="处理完成" value="2"></el-option>
                <el-option label="处理异常" value="3"></el-option>
            </el-select>
        </el-form-item>
        <el-form-item>
            <el-button type="primary" size="small" @click="search">搜索</el-button>
            <el-button size="small" @click="resetSearch">重置</el-button>
        </el-form-item>
    </el-form>
    <el-table
        :data="listData"
        style="width: 100%"
        v-loading="tableLoad">
        <el-table-column
            prop="shopname"
            label="商品名称"
            width="100">
        </el-table-column>
        <el-table-column
            prop="title"
            label="论文名称">
            <template slot-scope="scope">
                <el-tooltip effect="dark" :content="scope.row.title" placement="top">
                  <div class="text-ellipsis overflow-hidden">{{ scope.row.title }}</div>
                </el-tooltip>
              </template>
        </el-table-column>
        <el-table-column
            prop="studentName"
            label="姓名"
            width="80">
        </el-table-column>
        <el-table-column
            prop="major"
            label="专业"
            width="80">
        </el-table-column>
        <el-table-column
          prop="requires"
          label="论文要求"
          width="80">
          <template slot-scope="scope">
            <el-tooltip effect="dark" :content="scope.row.requires" placement="top">
              <div class="text-ellipsis overflow-hidden">{{ scope.row.requires }}</div>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
            label="是否降重"
            width="100">
            <template slot-scope="scope">
                <el-tag size="medium" type="success" v-if="scope.row.jiangchong == 1">降重</el-tag>
                <el-tag size="medium" type="warning" v-else>无需降重</el-tag>
            </template>
        </el-table-column>
        <el-table-column
            label="降低AIGC"
            width="100">
            <template slot-scope="scope">
                <el-tag size="medium" type="success" v-if="scope.row.aigc == 1">降低</el-tag>
                <el-tag size="medium" type="warning" v-else>无需降低</el-tag>
            </template>
        </el-table-column>
        <el-table-column
            prop="address"
            label="论文下载"
            width="100">
            <template slot-scope="scope">
                <el-link type="primary" :underline="false" @click="handlePaperDownload(scope.row)">论文下载</el-link>
            </template>
        </el-table-column>
        <el-table-column
            prop="price"
            label="价格"
            width="80">
        </el-table-column>
        <el-table-column
            label="状态"
            width="80">
            <template slot-scope="scope">
                <el-tag size="medium" type="warning" v-if="scope.row.state == 1">正在处理</el-tag>
                <el-tag size="medium" type="success" v-else-if="scope.row.state == 2">处理完成</el-tag>
                <el-tag size="medium" type="danger" v-else-if="scope.row.state == 3">处理异常</el-tag>
                <el-tag size="medium" type="info" v-else>待处理</el-tag>
            </template>
        </el-table-column>
        <el-table-column
            label="报告"
            width="180">
            <template slot-scope="scope">
                <el-tooltip class="item" effect="dark" :content="scope.row.reportContent" placement="top">
                    <div class="text-ellipsis overflow-hidden">{{scope.row.reportContent}}</div>
                </el-tooltip>
            </template>
        </el-table-column>
        <el-table-column
            prop="createTime"
            label="下单时间"
            width="180">
        </el-table-column>
        <el-table-column
            label="生成文档"
            width="180">
            <template slot-scope="scope">
                <div v-if="isNaN(scope.row.shopcode)">
                    <span style="color: #999;">仅论文类商品支持</span>
                </div>
                <div v-else>
                    <div style="margin-bottom: 5px;">
                        <el-link 
                            type="primary" 
                            :underline="false" 
                            @click="scope.row.rws ? handleDownloadTask(scope.row) : handleGenerateTask(scope.row)">
                            {{ scope.row.rws ? '下载任务书' : '生成任务书' }}
                        </el-link>
                    </div>
                    <div>
                        <el-link 
                            type="primary" 
                            :underline="false" 
                            @click="scope.row.ktbg ? handleDownloadProposal(scope.row) : handleGenerateProposal(scope.row)">
                            {{ scope.row.ktbg ? '下载开题报告' : '生成开题报告' }}
                        </el-link>
                    </div>
                </div>
            </template>
        </el-table-column>
    </el-table>
    <div style="margin-top:20px;text-align:right">
        <el-pagination
            background
            layout="prev, pager, next"
            :current-page="pageNum"
            :page-size="pageSize"
            :total="totalPageNum"
            @current-change="currentPageChange">
        </el-pagination>
    </div>
</div>
<script src="js/vue.min.js"></script>
<script src="js/vue-resource.min.js"></script>
<script src="js/element.js"></script>

<script type="text/javascript">
    var vm = new Vue({
        el: "#PaperList",
        data: {
            listData: [],
            pageNum: 1, //页码
            pageSize: 20, //页面数量
            totalPageNum: 0, //总页面数量
            tableLoad: false, //表格Load
            // 搜索表单数据
            searchForm: {
                title: '',
                shopname: '',
                studentName: '',
                state: ''
            },
            searchParams: {}
        },
        methods: {
            // 搜索订单
            search() {
                // 保存搜索参数
                this.searchParams = {
                    ...this.searchForm
                };
                // 重置到第一页
                this.pageNum = 1;
                // 获取数据
                this.getList();
            },

            // 重置搜索
            resetSearch() {
                // 重置表单
                this.searchForm = {
                    title: '',
                    shopname: '',
                    studentName: '',
                    state: ''
                };
                // 清空搜索参数
                this.searchParams = {};
                // 重置到第一页
                this.pageNum = 1;
                // 获取数据
                this.getList();
            },
            
            //当前页更改时
            currentPageChange(value) {
                this.pageNum = value;
                this.getList();
            },
            
            //获取列表
            getList() {
                this.tableLoad = true;
                // 构建请求参数
                let params = {
                    pageNum: this.pageNum,
                    pageSize: this.pageSize
                };
                // 添加搜索参数
                Object.keys(this.searchParams).forEach(key => {
                    if (this.searchParams[key] !== '') {
                        params[key] = this.searchParams[key];
                    }
                });

                // 构建URL参数
                let queryString = Object.entries(params)
                    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
                    .join('&');

                this.$http.get("aisdk/http.php?act=getList&" + queryString).then(function(res) {
                    this.tableLoad = false;
                    if (res.body.code == 200) {
                        this.listData = res.body.rows;
                        this.totalPageNum = res.body.total;
                    } else {
                        this.$message.error(res.body.msg);
                    }
                }, function(error) {
                    this.tableLoad = false;
                    this.$message.error('获取数据失败: ' + error.statusText);
                });
            },
            
            //论文下载
            handlePaperDownload(row) {
                if (!row.url) {
                    this.$message.error('下载地址不存在');
                    return;
                }

                this.$http.get("aisdk/http.php?act=paperDownload&orderId=" + row.url + "&fileName=-" + row.title).then(function(res) {
                    if (res.body.code == 200) {
                        const link = document.createElement('a')
                        const protocol = window.location.protocol;
                        const httpUrl = res.body.msg;
                        const httpsUrl = httpUrl.replace('http://', protocol + '//');
                        link.href = httpsUrl;
                        link.style.display = 'none';
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                    } else {
                        this.$message.error(res.body.msg || '下载失败');
                    }
                }, function(error) {
                    this.$message.error('下载请求失败: ' + error.statusText);
                });
            },
            
            // 生成任务书
            handleGenerateTask(row) {
                this.$confirm('确定要生成任务书吗?', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$http.post("aisdk/http.php?act=generateTask", {
                        id: row.id
                    }, {
                        emulateJSON: true
                    }).then(function(res) {
                        if (res.body.code == 200) {
                            this.$message.success('生成任务书成功');
                            this.getList(); // 刷新列表
                        } else {
                            this.$message.error(res.body.msg || '生成任务书失败');
                        }
                    }, function(error) {
                        this.$message.error('请求失败: ' + error.statusText);
                    });
                }).catch(() => {
                    this.$message.info('已取消操作');
                });
            },
            
            // 下载任务书
            handleDownloadTask(row) {
                this.$http.get("aisdk/http.php?act=paperDownload&orderId=" + row.rws + "&fileName=任务书-" + row.title).then(function(res) {
                    if (res.body.code == 200) {
                        const link = document.createElement('a');
                        const protocol = window.location.protocol;
                        const httpUrl = res.body.msg;
                        const httpsUrl = httpUrl.replace('http://', protocol + '//');
                        link.href = httpsUrl;
                        link.style.display = 'none';
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                    } else {
                        this.$message.error(res.body.msg || '下载任务书失败');
                    }
                }, function(error) {
                    this.$message.error('下载请求失败: ' + error.statusText);
                });
            },
            
            // 生成开题报告
            handleGenerateProposal(row) {
                this.$confirm('确定要生成开题报告吗?', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                }).then(() => {
                    this.$http.post("aisdk/http.php?act=generateProposal", {
                        id: row.id
                    }, {
                        emulateJSON: true
                    }).then(function(res) {
                        if (res.body.code == 200) {
                            this.$message.success('生成开题报告成功');
                            this.getList(); // 刷新列表
                        } else {
                            this.$message.error(res.body.msg || '生成开题报告失败');
                        }
                    }, function(error) {
                        this.$message.error('请求失败: ' + error.statusText);
                    });
                }).catch(() => {
                    this.$message.info('已取消操作');
                });
            },
            
            // 下载开题报告
            handleDownloadProposal(row) {
                this.$http.get("aisdk/http.php?act=paperDownload&orderId=" + row.ktbg + "&fileName=开题报告-" + row.title).then(function(res) {
                    if (res.body.code == 200) {
                        const link = document.createElement('a');
                        const protocol = window.location.protocol;
                        const httpUrl = res.body.msg;
                        const httpsUrl = httpUrl.replace('http://', protocol + '//');
                        link.href = httpsUrl;
                        link.style.display = 'none';
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                    } else {
                        this.$message.error(res.body.msg || '下载开题报告失败');
                    }
                }, function(error) {
                    this.$message.error('下载请求失败: ' + error.statusText);
                });
            }
        },
        mounted() {
            this.getList();
        }
    });
</script>
<style>
    .el-loading-spinner {
        left: 50%
    }
    .text-ellipsis {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
</style>
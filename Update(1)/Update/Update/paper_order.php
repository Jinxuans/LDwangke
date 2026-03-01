<?php
$title = '论文下单';
include('head.php');
?>
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">

<link href="css/element.css" rel="stylesheet">
<link href="css/tailwind.css?v=1111.211" rel="stylesheet">

<style>
/* 基础样式 */
body {
    min-width: auto;
    overflow-x: hidden;
}

/* 移动端样式 */
@media only screen and (max-width: 768px) {
    .app-content-body {
        padding: 10px !important;
    }
    .el-row--flex {
        display: flex;
        flex-direction: column;
    }
    .el-col {
        width: 100% !important;
        margin-bottom: 15px;
    }
    .box-card {
        margin-left: 0 !important;
        margin-right: 0 !important;
    }
    /* 对话框宽度调整 */
    .el-dialog {
        width: 90% !important;
        max-width: 100%;
    }
}

/* 桌面端样式 */
@media only screen and (min-width: 769px) {
    body {
        min-width: 1024px;
    }
    .el-col-md-10 {
        width: 41.66667%;
    }
    .el-col-md-14 {
        width: 58.33333%;
    }
    .box-card {
        min-height: 700px;
    }
}

/* 公共卡片样式 */
.box-card {
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    border-radius: 4px;
}

.el-card__body {
    height: 100%;
    overflow-y: auto;
}

.el-card__body .el-loading-parent--relative {
    height: 100%;
}

.el-card__body .el-loading-spinner {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    margin-top: 0;
}

.el-card__body .el-loading-spinner .circular {
    display: block;
    margin: 0 auto;
}

.outline-chapters-wrap {
    height: 100%;
}

.outline-chapters-wrap .el-empty {
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
}
 .upload-components input {
        display: none;
    }
    
    /* 新增：章节操作按钮样式 */
    .chapter-actions {
        display: flex;
        justify-content: flex-end;
        margin-top: 10px;
    }
    
    .section-actions {
        display: flex;
        justify-content: flex-end;
        margin-top: 5px;
    }
    
    .add-chapter-btn {
        margin-bottom: 15px;
    }
</style>

<div class="app-content-body" style="padding: 15px;" id="PaperOrder">
    <el-row :gutter="20" type='flex' style="height:max-content;">
        <!-- 左侧表单部分 - 移动端全宽，桌面端占10/24 -->
        <el-col :xs="24" :sm="24" :md="10" :lg="10" :xl="10">
            <el-card class="box-card" style="margin-left:10px;height:100%">
                <div>
                    <el-form ref="form" :model="form" :rules="rules" label-width="80px" size="medium">
                        <el-form-item label="商品类型" prop="shopcode">
                            <el-select v-model="form.shopcode" placeholder="请选择" @change="shopCodeChange">
                                <el-option
                                    v-for="item in goodsTypeOptions"
                                    :key="item.value"
                                    :label="item.label"
                                    :value="item.value">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="商品价格">
                            <el-link type="danger" :underline="false" style="font-size:22px">{{orderPrice}}</el-link>
                        </el-form-item>
                        <el-form-item label="论文标题" prop="title">
                            <el-row>
                                <el-col :span="16">
                                    <el-input v-model="form.title" placeholder="请输入论文标题"></el-input>
                                </el-col>
                                <el-col :span="6" :offset="2">
                                    <el-link type="primary" :underline="false" @click="openGenerateTitleDialog">生成标题</el-link>
                                </el-col>
                            </el-row>
                        </el-form-item>
                        <el-form-item label="姓名">
                            <el-input v-model="form.studentName" placeholder="请输入姓名（选填）"></el-input>
                        </el-form-item>
                        <el-form-item label="专业">
                            <el-input v-model="form.major" placeholder="请输入专业（选填）"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button type="primary" @click="handleGenerateOutline('form')">生成大纲</el-button>
                        </el-form-item>
                        <el-form-item label="论文要求">
                            <el-input type="textarea" v-model="form.requires" :rows="5" placeholder="请输入论文具体要求（选填）：
1. 可以填写论文的具体方向和主题
2. 可以填写需要包含的重点内容
3. 可以填写论文的现有框架
                                "></el-input>
                        </el-form-item>
                        <el-form-item label="附加服务">
                            <el-row style="border:1px solid #DCDFE6; padding:10px;margin-bottom:10px">
                                <el-col :span="14">
                                    任务书
                                </el-col>
                                <el-col :span="4">
                                    ¥{{rwsPrice}}
                                </el-col>
                                <el-col :span="4">
                                    <el-switch
                                        v-model="form.rws" :inactive-value="0"
                                        :active-value="1" @change="rwsChange">
                                    </el-switch>
                                </el-col>
                            </el-row>
                            <el-row style="border:1px solid #DCDFE6; padding:10px;margin-bottom:10px">
                                <el-col :span="14">
                                    开题报告
                                </el-col>
                                <el-col :span="4">
                                    ¥{{ktbgPrice}}
                                </el-col>
                                <el-col :span="4">
                                    <el-switch
                                        v-model="form.ktbg" :inactive-value="0"
                                        :active-value="1" @change="ktbgChange">
                                    </el-switch>
                                </el-col>
                            </el-row>
                            <el-row style="border:1px solid #DCDFE6; padding:10px;margin-bottom:10px">
                                <el-col :span="14">
                                    降低AIGC痕迹
                                </el-col>
                                <el-col :span="4">
                                    ¥{{jcPrice}}
                                </el-col>
                                <el-col :span="4">
                                    <el-switch
                                        v-model="form.jiangchong"
                                        :inactive-value="0"
                                        :active-value="1"
                                        @change="jcChange">
                                    </el-switch>
                                </el-col>
                            </el-row>
                            <el-row style="border:1px solid #DCDFE6; padding:10px">
                                <el-col :span="14">
                                    选择模板
                                </el-col>
                                <el-col :span="4">
                                    免费
                                </el-col>
                                <el-col :span="4">
                                    <el-button type="primary" size="small" @click="openPaperTemplateDialog">选择模板</el-button>
                                    <!-- 新增模板按钮 -->
                                    <el-button type="success" size="small" style="margin-left: 10px" @click="openAddTemplateDialog">新增模板</el-button>
                                </el-col>
                                <el-col :span="24" v-if="currentSelectedPaperTemplateTitle != ''">
                                    已选择模板：{{currentSelectedPaperTemplateTitle}}
                                </el-col>
                            </el-row>
                        </el-form-item>
                        <el-form-item>
                            <el-button type="primary" @click="onSubmit('form')" :loading="onSubmitLoad">提交订单</el-button>
                            <el-button @click="resetForm">重置</el-button>
                        </el-form-item>
                    </el-form>
                </div>
            </el-card>
        </el-col>
        
        <!-- 右侧大纲部分 - 移动端全宽，桌面端占14/24 -->
        <el-col :xs="24" :sm="24" :md="14" :lg="14" :xl="14">
            <el-card class="box-card" style="margin-left:10px;height:100%;">
                <div v-if="!isOutlineChaptersShow" class="outline-chapters-wrap">
                    <el-empty description="暂无大纲数据，请先生成大纲"></el-empty>
                </div>
                <div v-else v-loading="outlineChaptersLoad">
                    <!-- 添加章节按钮 -->
                    <div class="add-chapter-btn">
                        <el-button type="primary" icon="el-icon-plus" size="small" @click="addChapter">添加章节</el-button>
                    </div>
                    
                    <div class="border p-[15px] mb-10" v-for="(item, index) in outlineChapters" :key="index">
                        <el-row type="flex" align="middle">
                            <el-col :span="1" class="text-3xl font-semibold text-[#409EFF]">
                                {{index+1}}
                            </el-col>
                            <el-col :span="12">
                                <el-input size="mini" v-model="item.chapter_title" placeholder="请输入章节标题"></el-input>
                            </el-col>
                        </el-row>
                        <el-row class="mt-5">
                            <el-col :span="22" :offset="2">
                                <el-input type="textarea" :rows="2" v-model="item.chapter_desc" placeholder="请输入章节描述"></el-input>
                            </el-col>
                        </el-row>
                        
                        <!-- 小节列表 -->
                        <template v-for="(itemT, indexT) in item.sections" :key="indexT">
                            <el-row type="flex" align="middle" class="mt-5">
                                <el-col :span="1" :offset="3" class="text-[#909399] mr-2">
                                    {{index+1}}.{{indexT+1}}
                                </el-col>
                                <el-col :span="12">
                                    <el-input size="mini" v-model="itemT.section_title" placeholder="请输入小节标题"></el-input>
                                </el-col>
                                <el-col :span="4" :offset="1">
                                    <el-button type="danger" icon="el-icon-delete" circle size="mini" 
                                        @click="removeSection(index, indexT)" 
                                        :disabled="item.sections.length <= 1"></el-button>
                                </el-col>
                            </el-row>
                            <el-row class="mt-5">
                                <el-col :span="18" :offset="6">
                                    <el-input type="textarea" :rows="2" v-model="itemT.section_desc" placeholder="请输入小节写作要点"></el-input>
                                </el-col>
                            </el-row>
                        </template>
                        
                        <!-- 添加小节按钮 -->
                        <div class="section-actions">
                            <el-button type="success" icon="el-icon-plus" size="mini" @click="addSection(index)">添加小节</el-button>
                        </div>
                        
                        <!-- 章节操作按钮 -->
                        <div class="chapter-actions">
                            <el-button type="danger" icon="el-icon-delete" size="mini" @click="removeChapter(index)">删除章节</el-button>
                        </div>
                    </div>
                </div>
            </el-card>
        </el-col>
    </el-row>
    
    <!-- 生成标题对话框 -->
    <el-dialog
        title="生成论文标题"
        :visible.sync="generateTitleDialogVisible"
        width="50%">
        <el-input v-model="generateTitleMajor" placeholder="请输入专业方向，例如：计算机科学、经济学、教育学等"></el-input>
        <div v-if="titleOptions.length > 0">
            <el-radio-group v-model="selectedTitle" size="small" style="width:100%;margin-top:20px">
                <el-radio :label="item" border v-for="item in titleOptions" style="width:100%;margin:5px 0">{{item}}</el-radio>
            </el-radio-group>
        </div>
        <div slot="footer">
            <el-button @click="generateTitleDialogVisible=false">取消</el-button>
            <el-button type="primary" @click="handleGenerateTitle" :disabled="generateTitleMajor=='' || generateTitleLoad">{{generateTitleLoad ? '正在生成中...' : '生成标题'}}</el-button>
            <el-button type="success" :disabled="selectedTitle==''" @click="handleSelectedTitle">使用选中的标题</el-button>
        </div>
    </el-dialog>
    
    <!-- 选择模板对话框 -->
    <el-dialog
        title="选择论文模板"
        :visible.sync="paperTemplateDialogVisible"
        width="50%">
        <el-row :gutter="20" style="margin-bottom: 15px;">
            <el-col :span="12">
                <el-radio-group v-model="templateType" @change="templateTypeChange">
                    <el-radio-button :label="1">公共模板</el-radio-button>
                    <el-radio-button :label="0">我的模板</el-radio-button>
                </el-radio-group>
            </el-col>
            <el-col :span="12">
                <el-input v-model="paperTemplateTableSearch" placeholder="请输入模板名称"></el-input>
            </el-col>
            <el-col :span="2" :offset="20">
                <el-button type="primary" icon="el-icon-search" @click="getPaperTemplateList">搜索</el-button>
            </el-col>
        </el-row>
        <div>
            <el-table
                ref="singleTable"
                :data="paperTemplateData"
                highlight-current-row
                v-loading="paperTemplateTableLoad"
                @current-change="handleCurrentRowChange"
                style="width: 100%">
                <el-table-column
                    property="name"
                    label="模板名称">
                </el-table-column>
            </el-table>
        </div>
        <div style="margin-top:20px;text-align:right">
            <el-pagination
                :hide-on-single-page="paperTemplateTotalPageNum <= paperTemplatePageSize"
                background
                layout="prev, pager, next"
                :current-page="paperTemplatePageNum"
                :page-size="paperTemplatePageSize"
                :total="paperTemplateTotalPageNum"
                @current-change="paperTemplateCurrentPageChange">
            </el-pagination>
        </div>
        <div slot="footer">
            <el-button @click="paperTemplateDialogVisible=false">取消</el-button>
            <el-button type="primary" @click="handleSelectedPaperTemplate" :disabled="selectedPaperTemplateFormat == ''">使用此模板</el-button>
        </div>
    </el-dialog>

    <!-- 新增模板对话框 -->
    <el-dialog
        title="新增论文模板"
        :visible.sync="addTemplateDialogVisible"
        width="35%"
        append-to-body>
        <el-form ref="templateForm" :model="templateForm" :rules="templateRules" label-width="100px">
            <el-form-item label="模板名称" prop="name">
                <el-input 
                    v-model="templateForm.name" 
                    placeholder="请输入模板名称" 
                    clearable
                    required>
                </el-input>
            </el-form-item>
            <el-form-item label="上传模板" prop="file">
                <el-upload
                    class="upload-components"
                    action="aisdk/http.php?act=uploadCover"
                    accept=".docx"
                    :limit="1"
                    :on-change="handleTemplateUpload"
                    :on-success="handleUploadSuccess"
                    :on-error="handleUploadError"
                    :auto-upload="true"
                    :show-file-list="false">
                    <i class="el-icon-upload"></i>
                    <div class="el-upload__text">
                        <em>点击上传Word模板（.docx格式）</em>
                    </div>
                    <div class="el-upload__tip" slot="tip">支持.docx格式，最大50MB</div>
                </el-upload>
                <div v-if="templateForm.coverUrl" class="mt-2 text-sm text-success">
                    <i class="el-icon-check"></i> 已上传: {{templateForm.name}}
                </div>
            </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
            <el-button @click="addTemplateDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="fileSubmit" :loading="submitTemplateLoading">
                {{ submitTemplateLoading ? '提交中...' : '提交模板' }}
            </el-button>
        </div>
    </el-dialog>
</div>

<script src="js/vue.min.js"></script>
<script src="js/vue-resource.min.js"></script>
<script src="js/element.js"></script>

<script type="text/javascript">
var vm = new Vue({
    el: "#PaperOrder",
    data: {
        // 新增模板相关数据
        addTemplateDialogVisible: false,
        submitTemplateLoading: false,
        templateForm: {
            name: '',
            file: null,
            coverUrl: '', // 存储上传后的文件地址
        },
        templateRules: {
            name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
            file: [{ required: true, message: '请上传模板文件', trigger: 'change' }]
        },
        
        // 原有数据
        input: '',
        goodsTypeOptions: [{
            value: '6000',
            label: '论文6000字'
        }, {
            value: '8000',
            label: '论文8000字'
        }, {
            value: '10000',
            label: '论文10000字'
        }, {
            value: '12000',
            label: '论文12000字'
        }, {
            value: '15000',
            label: '论文15000字'
        }], //商品类型
        orderPrice: 0, //商品价格
        generateTitleDialogVisible: false, //生成标题框
        generateTitleMajor: '', //生成标题专业方向内容
        generateTitleLoad: false, //生成标题Load
        titleOptions: [], //标题列表
        selectedTitle: '', //选中标题
        paperTemplateDialogVisible: false, //模板选择框
        paperTemplateData: [], //论文模板数据
        paperTemplatePageNum: 1, //页码
        paperTemplatePageSize: 9, //页面数量
        paperTemplateTotalPageNum: 0, //总页面数量
        paperTemplateTableLoad: false, //表格Load
        paperTemplateTableSearch: '', //模板搜索
        selectedPaperTemplateTitle: '', //选中模板标题
        selectedPaperTemplateFormat: '', //选中模板Format
        currentSelectedPaperTemplateTitle: '', //当前使用模板标题
        outlineChapters: [], //大纲内容
        isOutlineChaptersShow: false,
        rwsPrice: <?= bcmul($conf['lunwen_api_rws_price'], $userrow['addprice'], 2) ?>,
        ktbgPrice: <?= bcmul($conf['lunwen_api_ktbg_price'], $userrow['addprice'], 2) ?>,
        jcPrice: <?= bcmul($conf['lunwen_api_jdaigchj_price'], $userrow['addprice'], 2) ?>,
        isRwsPrice: 0, //是否有价格
        isKtbgPrice: 0,
        isJcPrice: 0,
        form: {
            shopcode: '',
            title: '',
            studentName: '',
            major: '',
            requires: '',
            yijian: '',
            ktbg: 0,
            jiangchong: 0,
            rws: 0,
            tigang:'',
        },
        rules: {
            shopcode: [{
                required: true,
                message: '请选择商品类型',
                trigger: 'blur'
            }],
            title: [{
                required: true,
                message: '请输入论文标题',
                trigger: 'blur'
            }],
        },
        timer: null, //定时器
        outlineChaptersLoad: false, //大纲load
        onSubmitLoad: false, //提交Load
        templateType: 1, // 1: 公共模板, 0: 我的模板
    },
    methods: {
        // 新增：章节管理方法
        addChapter() {
            this.outlineChapters.push({
                chapter_title: '新章节',
                chapter_desc: '',
                sections: [{
                    section_title: '新小节',
                    section_desc: '',
                    sub_sections: [] // 添加空数组
                }]
            });
        },
        
        removeChapter(index) {
            this.$confirm('确定要删除这个章节吗？', '提示', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }).then(() => {
                this.outlineChapters.splice(index, 1);
                this.$message.success('章节已删除');
            }).catch(() => {});
        },
        
        addSection(chapterIndex) {
            this.outlineChapters[chapterIndex].sections.push({
                section_title: '新小节',
                section_desc: '',
                sub_sections: [] // 添加空数组
            });
        },
        
        removeSection(chapterIndex, sectionIndex) {
            if (this.outlineChapters[chapterIndex].sections.length <= 1) {
                this.$message.warning('每个章节至少需要一个小节');
                return;
            }
            
            this.$confirm('确定要删除这个小节吗？', '提示', {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
            }).then(() => {
                this.outlineChapters[chapterIndex].sections.splice(sectionIndex, 1);
                this.$message.success('小节已删除');
            }).catch(() => {});
        },
        
        // 新增：检查提纲内容是否有效
        validateOutline() {
            if (!this.isOutlineChaptersShow || this.outlineChapters.length === 0) {
                this.$message.warning('请先生成大纲或添加章节');
                return false;
            }
            
            for (let i = 0; i < this.outlineChapters.length; i++) {
                const chapter = this.outlineChapters[i];
                
                // 检查章节标题
                if (!chapter.chapter_title || !chapter.chapter_title.trim()) {
                    this.$message.error(`第 ${i+1} 章节标题不能为空`);
                    return false;
                }
                
                // 检查小节
                if (chapter.sections.length === 0) {
                    this.$message.error(`第 ${i+1} 章节至少需要一个小节`);
                    return false;
                }
                
                for (let j = 0; j < chapter.sections.length; j++) {
                    const section = chapter.sections[j];
                    
                    // 检查小节标题
                    if (!section.section_title || !section.section_title.trim()) {
                        this.$message.error(`第 ${i+1} 章节的第 ${j+1} 小节标题不能为空`);
                        return false;
                    }
                    
                    // 确保有sub_sections字段
                    if (!section.hasOwnProperty('sub_sections')) {
                        section.sub_sections = [];
                    }
                }
            }
            
            return true;
        },
        
        // 新增模板相关方法
        openAddTemplateDialog() {
            this.addTemplateDialogVisible = true;
            this.templateForm = { name: '', file: null, coverUrl: '' };
            this.$refs.templateForm && this.$refs.templateForm.resetFields();
        },
        
        handleTemplateUpload(file) {
            this.templateForm.file = file.raw;
            // 自动填充模板名称（如果为空）
            if (!this.templateForm.name) {
                this.templateForm.name = file.name.replace(/\.docx$/, '');
            }
        },
        
        // 处理上传成功
        handleUploadSuccess(response, file, fileList) {
            console.log('上传响应:', response); // 关键调试步骤
            if (response.code === 200) {
                // 从服务器响应中获取文件URL
                this.templateForm.coverUrl = response.data.url;
                
                this.$message.success('文件上传成功，自动获取文件名称作为模板名称(可自行编辑),提交模板即可保存');
            } else {
                this.templateForm.coverUrl = '';
                this.$message.error(`上传失败: ${response.msg || '服务器未返回有效数据'}`);
            }
        },
        
        // 处理上传失败
        handleUploadError(err, file, fileList) {
            this.templateForm.coverUrl = '';
            this.$message.error('上传失败，请重试');
            console.error('上传错误:', err);
        },
        
        // 提交模板信息（使用JSON参数）
        fileSubmit() {
            this.$refs.templateForm.validate((valid) => {
                if (valid) {
                    console.log('提交模板信息:', this.templateForm); // 调试用
                    
                    // 检查coverUrl是否存在
                    if (!this.templateForm.coverUrl) {
                        this.$message.error('请先上传文件');
                        return;
                    }
                    
                    this.submitTemplateLoading = true;
                    
                    // 构建JSON参数
                    const jsonData = {
                        name: this.templateForm.name,
                        coverUrl: this.templateForm.coverUrl,
                        imgString: "",
                        formatSettings: "{\"margins\":{\"top\":2.54,\"bottom\":2.54,\"left\":2.54,\"right\":2.54},\"titles\":{\"level1\":{\"fontFamily\":\"黑体\",\"fontSize\":\"18\",\"alignment\":\"center\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":12,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false},\"level2\":{\"fontFamily\":\"黑体\",\"fontSize\":\"16\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":12,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false},\"level3\":{\"fontFamily\":\"黑体\",\"fontSize\":\"14\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":12,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false}},\"body\":{\"fontFamily\":\"宋体\",\"fontSize\":\"12\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":2,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":0,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":0,\"spaceAfterUnit\":\"pt\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false},\"cover\":{\"templateName\":\"集美学院\",\"coverUrl\":\"e74c1e7f-65df-4b9f-8b2d-5721f6b1de2a\",\"coverFileName\":\"-新建+Microsoft+Word+文档.docx.docx\"},\"abstract\":{\"chinese\":{\"fontFamily\":\"宋体\",\"fontSize\":\"12\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":2,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":6,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false},\"english\":{\"fontFamily\":\"Times New Roman\",\"fontSize\":\"12\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":6,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false},\"chineseTitle\":{\"fontFamily\":\"黑体\",\"fontSize\":\"16\",\"alignment\":\"center\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false},\"englishTitle\":{\"fontFamily\":\"Times New Roman\",\"fontSize\":\"16\",\"alignment\":\"center\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false}},\"header\":{\"content\":\"\",\"fontFamily\":\"宋体\",\"fontSize\":\"10.5\",\"alignment\":\"center\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false,\"showLine\":true,\"lineStyle\":\"solid\",\"lineWidth\":1,\"marginTop\":1.5,\"marginTopUnit\":\"cm\"},\"footer\":{\"marginBottom\":1.75,\"marginBottomUnit\":\"cm\"},\"acknowledgement\":{\"title\":{\"fontFamily\":\"黑体\",\"fontSize\":\"16\",\"alignment\":\"center\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false},\"content\":{\"fontFamily\":\"宋体\",\"fontSize\":\"12\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":2,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":0,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":0,\"spaceAfterUnit\":\"pt\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false}},\"references\":{\"title\":{\"fontFamily\":\"黑体\",\"fontSize\":\"16\",\"alignment\":\"center\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":0,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":12,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":6,\"spaceAfterUnit\":\"pt\",\"isBold\":true,\"isItalic\":false,\"isUnderline\":false},\"content\":{\"fontFamily\":\"宋体\",\"fontSize\":\"12\",\"alignment\":\"left\",\"lineHeight\":\"1.5\",\"lineHeightType\":\"1.5\",\"lineHeightValue\":1.5,\"firstLineIndent\":2,\"firstLineIndentUnit\":\"char\",\"spaceBefore\":0,\"spaceBeforeUnit\":\"pt\",\"spaceAfter\":0,\"spaceAfterUnit\":\"pt\",\"isBold\":false,\"isItalic\":false,\"isUnderline\":false}},\"additionalOptions\":{\"promptText\":\"生成的章节标题遵循以下规则\\n第一章 xxxxx\\n1.1 xxxxxx\\n1.2 xxxxx\\n1.3 xxxxxx\\n1.3.1 xxxxx\"}}",
                        isPublic: "0"
                    };
                    
                    // 使用JSON格式发送请求
                    this.$http.post("aisdk/http.php?act=systemtemplate", jsonData, {
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        emulateJSON: false // 禁用表单编码，使用JSON
                    }).then((res) => {
                        this.submitTemplateLoading = false;
                        
                        if (res.body.code == 200) {
                            this.$message.success('模板创建成功，前往我的模板查看');
                            this.addTemplateDialogVisible = false;
                            this.getPaperTemplateList(); // 刷新模板列表
                            
                            // 自动选择刚上传的模板
                            this.currentSelectedPaperTemplateTitle = this.templateForm.name;
                            this.form.format = this.templateForm.coverUrl;
                        } else {
                            this.$message.error(res.body.msg || '提交模板失败，请重试');
                        }
                    }).catch((err) => {
                        this.submitTemplateLoading = false;
                        this.$message.error('网络错误，请稍后再试');
                        console.error('提交模板错误:', err);
                    });
                }
            });
        },
        
        // 提交生成大纲
        handleGenerateOutline(formName) {
            this.$refs[formName].validate((valid) => {
                if (valid) {
                    const outlineForm = {
                        customRequirements: this.form.requires,
                        title: this.form.title,
                        wordCount: this.form.shopcode
                    };
                    this.$http.post("aisdk/http.php?act=generateOutline", outlineForm, {
                        emulateJSON: true
                    }).then(function(res) {
                        if (res.body.code == 200) {
                            this.timer = setInterval(() => {
                                this.checkGenerateOutlineProgress(res.body.msg);
                            }, 3000);
                            this.isOutlineChaptersShow = true;
                            this.outlineChaptersLoad = true;
                        } else {
                            this.$message.error(res.body.msg);
                        }
                    });
                } else {
                    return false;
                }
            });
        },
        
        //检索大纲进度
        checkGenerateOutlineProgress(orderId) {
            this.$http.get("aisdk/http.php?act=outlineStatus&orderId=" + orderId).then(function(res) {
                if (res.body.msg != 'processing') {
                    this.outlineChaptersLoad = false;
                    
                    // 确保章节中的小节都有sub_sections字段
                    if (res.body.data && res.body.data.chapters) {
                        res.body.data.chapters.forEach(chapter => {
                            if (chapter.sections) {
                                chapter.sections.forEach(section => {
                                    if (!section.hasOwnProperty('sub_sections')) {
                                        section.sub_sections = [];
                                    }
                                });
                            }
                        });
                    }
                    
                    this.outlineChapters = res.body.data.chapters;
                    this.form.tigang = JSON.stringify(res.body.data); // 保持原始格式
                    clearInterval(this.timer);
                }
            });
        },
        
        shopCodeChange(value) {
            switch (value) {
                case '6000':
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice + <?= bcmul($conf['lunwen_api_6000_price'], $userrow['addprice'], 2) ?>;
                    break;
                case '8000':
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice + <?= bcmul($conf['lunwen_api_8000_price'], $userrow['addprice'], 2) ?>;
                    break;
                case '10000':
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice + <?= bcmul($conf['lunwen_api_10000_price'], $userrow['addprice'], 2) ?>;
                    break;
                case '12000':
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice + <?= bcmul($conf['lunwen_api_12000_price'], $userrow['addprice'], 2) ?>;
                    break;
                case '15000':
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice + <?= bcmul($conf['lunwen_api_15000_price'], $userrow['addprice'], 2) ?>;
                    break;
                default:
                    this.orderPrice = this.isRwsPrice + this.isKtbgPrice + this.isJcPrice;
            }
        },
        
        //打开生成标题框
        openGenerateTitleDialog() {
            this.generateTitleDialogVisible = true;
            this.generateTitleMajor = '';
            this.titleOptions = [];
            this.selectedTitle = '';
            this.generateTitleLoad = false;
        },
        
        //获取可选择标题列表
        handleGenerateTitle() {
            if (this.generateTitleMajor == '') {
                this.$message.warning('请输入专业方向');
                return;
            }
            this.generateTitleLoad = true;
            this.$http.post("aisdk/http.php?act=generateTitles", {
                direction: this.generateTitleMajor
            }, {
                emulateJSON: true
            }).then(function(res) {
                this.generateTitleLoad = false;
                
                if (res.body.code == 200) {
                    this.titleOptions = res.body.data.titles;
                } else {
                    this.$message.error(res.body.msg);
                }
            });
        },
        
        //选择标题
        handleSelectedTitle() {
            this.form.title = this.selectedTitle;
            this.generateTitleDialogVisible = false;
        },
        
        //打开模板选择框
        openPaperTemplateDialog() {
            this.paperTemplatePageNum = 1;
            this.paperTemplateTableSearch = "";
            this.templateType = 1; // 默认显示公共模板
            this.paperTemplateDialogVisible = true;
            this.getPaperTemplateList();
        },
        
        //获取论文模板列表
        getPaperTemplateList() {
            this.paperTemplateTableLoad = true;
            if (this.paperTemplateTableSearch != '') {
                this.paperTemplatePageNum = 1;
            }
            // 根据模板类型设置isPublic参数
            const isPublic = this.templateType === 1 ? 1 : 0;
            this.$http.get("aisdk/http.php?act=getTemplate&pageNum=" + this.paperTemplatePageNum + "&pageSize=" + this.paperTemplatePageSize + '&name=' + this.paperTemplateTableSearch + '&isPublic=' + isPublic).then(function(res) {
                this.paperTemplateTableLoad = false;
                if (res.body.code == 200) {
                    this.paperTemplateData = res.body.rows;
                    this.paperTemplateTotalPageNum = res.body.total;
                } else {
                    this.$message.error(res.body.msg);
                }
            });
        },
        
        // 模板类型切换
        templateTypeChange(value) {
            this.templateType = value;
            this.getPaperTemplateList();
        },
        
        //论文模板当前页更改时
        paperTemplateCurrentPageChange(value) {
            this.paperTemplatePageNum = value;
            this.getPaperTemplateList();
        },
        
        //获取当前选中行
        handleCurrentRowChange(row) {
            this.selectedPaperTemplateTitle = row.name;
            this.selectedPaperTemplateFormat = row.formatSettings;
        },
        
        //使用模板
        handleSelectedPaperTemplate() {
            this.currentSelectedPaperTemplateTitle = this.selectedPaperTemplateTitle;
            this.form.format = this.selectedPaperTemplateFormat;
            this.paperTemplateDialogVisible = false;
        },
        
        //重置表单
        resetForm() {
            this.currentSelectedPaperTemplateTitle = '', //当前使用模板标题
            this.form = {
                format: '',
                shopcode: '',
                title: '',
                studentName: '',
                major: '',
                requires: '',
                yijian: '',
                ktbg: 0,
                jiangchong: 0,
                rws: 0,
                tigang:''
            }
            this.outlineChapters = [];
            this.isOutlineChaptersShow = false;
        },
        
        // 修改后的提交方法：解决提纲数据格式问题
        onSubmit(formName) {
            this.$refs[formName].validate((valid) => {
                if (valid) {
                    // 验证提纲内容
                    if (!this.validateOutline()) {
                        return;
                    }
                    
                    this.form.title = this.filterTitleSymbols(this.form.title);
                    this.onSubmitLoad = true;
                    
                    // 确保所有小节都有sub_sections字段
                    this.outlineChapters.forEach(chapter => {
                        chapter.sections.forEach(section => {
                            if (!section.hasOwnProperty('sub_sections')) {
                                section.sub_sections = [];
                            }
                        });
                    });
                    
                    // 构建符合源台要求的提纲数据结构
                    const tigangData = {
                        chapters: this.outlineChapters
                    };
                    
                    // 双重转义：先创建对象，再转换为字符串
                    this.form.tigang = JSON.stringify(tigangData);
                    
                    this.$http.post("aisdk/http.php?act=paperOrder", this.form, {
                        emulateJSON: true
                    }).then((res) => {
                        this.onSubmitLoad = false;
                        console.log("接口响应数据:", res.body);
                        console.log("提交的提纲数据:", this.form.tigang);
                        
                        if (res.body.code === 200) {
                            this.$message.success(res.body.msg || res.body.data);
                            this.resetForm();
                        } else {
                            // 新增：适配code=-1和data字段的情况
                            const isBalanceError = 
                                (res.body.code === -1 || res.body.code === 500) && 
                                (res.body.data || res.body.msg).includes("余额不足");
                            
                            if (isBalanceError) {
                                this.$message.error("源台余额不足，请联系管理员");
                            } else {
                                // 显示完整错误信息（包含code和data/msg）
                                this.$message.error(`接口错误: ${res.body.code}\n错误详情: ${res.body.data || res.body.msg}`);
                            }
                        }
                    }).catch((err) => {
                        this.onSubmitLoad = false;
                        this.$message.error(`网络请求失败: ${err.message}`);
                        console.error("接口请求错误:", err);
                    });
                } else {
                    return false;
                }
            });
        },
        
        // 标题符号过滤函数
        filterTitleSymbols(title) {
            // 过滤所有非中文字符、英文字母、数字的特殊符号
            return title.replace(/[^\u4e00-\u9fa5a-zA-Z0-9]/g, '');
        },
        
        rwsChange(value) {
            if (value == 1) {
                this.orderPrice = this.orderPrice + this.rwsPrice;
                this.isRwsPrice = JSON.parse(JSON.stringify(this.rwsPrice));
            } else {
                this.orderPrice = this.orderPrice - this.rwsPrice;
                this.isRwsPrice = 0;
            }
        },
        
        ktbgChange(value) {
            if (value == 1) {
                this.orderPrice = this.orderPrice + this.ktbgPrice;
                this.isKtbgPrice = JSON.parse(JSON.stringify(this.ktbgPrice));
            } else {
                this.orderPrice = this.orderPrice - this.ktbgPrice;
                this.isKtbgPrice = 0;
            }
        },
        
        jcChange(value) {
            if (value == 1) {
                this.orderPrice = this.orderPrice + this.jcPrice;
                this.isJcPrice = JSON.parse(JSON.stringify(this.jcPrice));
            } else {
                this.orderPrice = this.orderPrice - this.jcPrice;
                this.isJcPrice = 0;
            }
        },
    },
    mounted() {
    }
});
</script>
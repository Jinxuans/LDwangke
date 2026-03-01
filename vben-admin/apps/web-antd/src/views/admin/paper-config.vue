<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Form, FormItem, Input, Button, Tabs, TabPane, message, Spin } from 'ant-design-vue';
import { paperConfigGetApi, paperConfigSaveApi, type PaperConfig } from '#/api/paper';

const loading = ref(false);
const saving = ref(false);
const form = ref<PaperConfig>({
  lunwen_api_username: '',
  lunwen_api_password: '',
  lunwen_api_6000_price: '30',
  lunwen_api_8000_price: '40',
  lunwen_api_10000_price: '50',
  lunwen_api_12000_price: '60',
  lunwen_api_15000_price: '75',
  lunwen_api_rws_price: '10',
  lunwen_api_ktbg_price: '10',
  lunwen_api_jdaigchj_price: '10',
  lunwen_api_xgdl_price: '3',
  lunwen_api_jcl_price: '3',
  lunwen_api_jdaigcl_price: '3',
});

async function loadConfig() {
  loading.value = true;
  try {
    const res = await paperConfigGetApi();
    if (res && typeof res === 'object') {
      Object.assign(form.value, res);
    }
  } catch (e: any) {
    message.error('加载配置失败: ' + (e.message || ''));
  }
  loading.value = false;
}

async function handleSave() {
  saving.value = true;
  try {
    await paperConfigSaveApi(form.value);
    message.success('保存成功');
  } catch (e: any) {
    message.error('保存失败: ' + (e.message || ''));
  }
  saving.value = false;
}

onMounted(() => { loadConfig(); });
</script>

<template>
  <Page title="智文论文配置" description="配置智文论文API账号和价格">
    <Spin :spinning="loading">
      <Card>
        <Tabs>
          <TabPane key="api" tab="接口配置">
            <Form :labelCol="{ span: 4 }" :wrapperCol="{ span: 16 }" style="max-width: 600px">
              <FormItem label="登录账号">
                <Input v-model:value="form.lunwen_api_username" placeholder="请输入登录账号" />
              </FormItem>
              <FormItem label="登录密码">
                <Input v-model:value="form.lunwen_api_password" placeholder="请输入登录密码" />
              </FormItem>
            </Form>
          </TabPane>
          <TabPane key="price" tab="价格配置（实际扣费=本页价格×用户费率）">
            <Form :labelCol="{ span: 6 }" :wrapperCol="{ span: 14 }" style="max-width: 600px">
              <FormItem label="论文6000字价格">
                <Input v-model:value="form.lunwen_api_6000_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="论文8000字价格">
                <Input v-model:value="form.lunwen_api_8000_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="论文10000字价格">
                <Input v-model:value="form.lunwen_api_10000_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="论文12000字价格">
                <Input v-model:value="form.lunwen_api_12000_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="论文15000字价格">
                <Input v-model:value="form.lunwen_api_15000_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="任务书价格">
                <Input v-model:value="form.lunwen_api_rws_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="开题报告价格">
                <Input v-model:value="form.lunwen_api_ktbg_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="降低AIGC痕迹价格">
                <Input v-model:value="form.lunwen_api_jdaigchj_price" placeholder="请输入价格" suffix="元" />
              </FormItem>
              <FormItem label="修改段落千字价格">
                <Input v-model:value="form.lunwen_api_xgdl_price" placeholder="请输入价格" suffix="元/千字" />
              </FormItem>
              <FormItem label="降重率千字价格">
                <Input v-model:value="form.lunwen_api_jcl_price" placeholder="请输入价格" suffix="元/千字" />
              </FormItem>
              <FormItem label="降低AIGC率千字价格">
                <Input v-model:value="form.lunwen_api_jdaigcl_price" placeholder="请输入价格" suffix="元/千字" />
              </FormItem>
            </Form>
          </TabPane>
        </Tabs>
        <div class="mt-4">
          <Button type="primary" @click="handleSave" :loading="saving">保存配置</Button>
        </div>
      </Card>
    </Spin>
  </Page>
</template>

import { requestClient } from '#/api/request';

export interface ClassItem {
  cid: number;
  name: string;
  price: string;
  content?: string;
  fenlei?: string;
  docking?: string;
  noun?: string;
  status: number;
}

export interface ClassCategory {
  id: number;
  name: string;
  recommend?: number;
  log?: number;
  ticket?: number;
  changepass?: number;
  allowpause?: number;
}

export interface ClassListPagedResult {
  list: ClassItem[];
  pagination: {
    has_more: boolean;
    limit: number;
    page: number;
    total: number;
  };
}

export interface CourseItem {
  id: string;
  name: string;
  kcjs?: string;
  studyStartTime?: string;
  studyEndTime?: string;
  examStartTime?: string;
  examEndTime?: string;
  learnStatusName?: string;
  complete?: string;
  idx?: number;
  select?: boolean;
}

export interface CourseQueryResult {
  data: CourseItem[];
  msg: string;
  userinfo: string;
  userName: string;
}

/** 获取课程列表 */
export async function getClassListApi(params?: {
  fenlei?: number;
  search?: string;
}) {
  return requestClient.get<ClassItem[]>('/class/list', { params });
}

/** 获取分页课程列表 */
export async function getClassListPagedApi(params?: {
  favorite?: number;
  fenlei?: number;
  limit?: number;
  page?: number;
  search?: string;
}) {
  return requestClient.get<ClassListPagedResult>('/class/list-paged', {
    params,
  });
}

/** 获取课程分类 */
export async function getClassCategoriesApi() {
  return requestClient.get<ClassCategory[]>('/class/categories');
}

/** 查课 */
export async function queryCourseApi(cid: number, userinfo: string) {
  return requestClient.post<CourseQueryResult>('/class/search', {
    cid,
    userinfo,
  });
}

/** 获取课程所属分类的开关配置 */
export async function getCategorySwitchesApi(cid: number) {
  return requestClient.get<{
    allowpause: number;
    changepass: number;
    log: number;
    ticket: number;
  }>('/class/category-switches', { params: { cid } });
}

/** 下单 */
export async function addOrderApi(data: {
  cid: number;
  data: Array<{ data: CourseItem; userinfo: string; userName: string }>;
  remarks?: string;
}) {
  return requestClient.post('/order/add', data);
}

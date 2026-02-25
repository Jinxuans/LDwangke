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

export interface CourseQueryResult {
  userinfo: string;
  userName: string;
  msg: string;
  data: CourseItem[];
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

/** 获取课程列表 */
export async function getClassListApi(params?: { fenlei?: number; search?: string }) {
  return requestClient.get<ClassItem[]>('/class/list', { params });
}

/** 获取课程分类 */
export async function getClassCategoriesApi() {
  return requestClient.get<ClassCategory[]>('/class/categories');
}

/** 查课 */
export async function queryCourseApi(cid: number, userinfo: string) {
  return requestClient.post<CourseQueryResult>('/class/search', { cid, userinfo });
}

/** 获取课程所属分类的开关配置 */
export async function getCategorySwitchesApi(cid: number) {
  return requestClient.get<{ log: number; ticket: number; changepass: number; allowpause: number }>('/class/category-switches', { params: { cid } });
}

/** 下单 */
export async function addOrderApi(data: {
  cid: number;
  data: Array<{ userinfo: string; userName: string; data: CourseItem }>;
  remarks?: string;
}) {
  return requestClient.post('/order/add', data);
}

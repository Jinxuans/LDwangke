<script lang="ts">
import type { Component } from 'vue';
// ─── Ant Design Icons ───
import {
  ThunderboltOutlined, FireOutlined, TrophyOutlined, AimOutlined,
  DashboardOutlined, HeartOutlined, RocketOutlined, FlagOutlined,
  CrownOutlined, StarOutlined, SafetyCertificateOutlined, SmileOutlined,
  BankOutlined, TeamOutlined, SolutionOutlined, ScheduleOutlined,
  IdcardOutlined, AccountBookOutlined, DollarOutlined, FundOutlined,
  AuditOutlined, ContactsOutlined, ReconciliationOutlined, InsuranceOutlined,
  ReadOutlined, BookOutlined, FileTextOutlined, HighlightOutlined,
  ExperimentOutlined, BulbOutlined, FormOutlined, FileDoneOutlined,
  FileSearchOutlined, EditOutlined, ProfileOutlined, SnippetsOutlined,
  LaptopOutlined, MobileOutlined, CloudOutlined, GlobalOutlined,
  ApiOutlined, SettingOutlined, ToolOutlined, LockOutlined,
  KeyOutlined, WifiOutlined, DatabaseOutlined, CodeOutlined,
  BugOutlined, DeploymentUnitOutlined, ClusterOutlined, ConsoleSqlOutlined,
  AppstoreOutlined, ShopOutlined, ShoppingCartOutlined, CarOutlined,
  HomeOutlined, GiftOutlined, BellOutlined, TagOutlined,
  PictureOutlined, SoundOutlined, VideoCameraOutlined, EnvironmentOutlined,
  CompassOutlined, CalendarOutlined, ClockCircleOutlined, SendOutlined,
  // 扩充
  ThunderboltFilled, FireFilled, TrophyFilled, HeartFilled, StarFilled,
  CrownFilled, BellFilled, GiftFilled, HomeFilled, ShopFilled,
  CameraOutlined, PrinterOutlined, WalletOutlined, SkinOutlined,
  MedicineBoxOutlined, CoffeeOutlined, CarryOutOutlined, ScanOutlined,
  RadarChartOutlined, RiseOutlined, FallOutlined, InteractionOutlined,
  ShoppingOutlined, MoneyCollectOutlined, TransactionOutlined, PropertySafetyOutlined,
  CloudServerOutlined, CloudDownloadOutlined, CloudUploadOutlined,
  HddOutlined, DesktopOutlined, TabletOutlined, UsbOutlined,
  AlertOutlined, CheckCircleOutlined, CloseCircleOutlined, InfoCircleOutlined,
  ExclamationCircleOutlined, QuestionCircleOutlined, StopOutlined,
  UserOutlined, UsergroupAddOutlined, ManOutlined, WomanOutlined,
  PhoneOutlined, MailOutlined, MessageOutlined, NotificationOutlined,
  LikeOutlined, DislikeOutlined, ShareAltOutlined, LinkOutlined,
  ScissorOutlined, CopyOutlined, DeleteOutlined, SaveOutlined,
  DownloadOutlined, UploadOutlined, SwapOutlined, SyncOutlined,
  FilterOutlined, SearchOutlined, ZoomInOutlined, ZoomOutOutlined,
  ExpandOutlined, CompressOutlined, FullscreenOutlined,
  EyeOutlined, EyeInvisibleOutlined, PaperClipOutlined,
  PushpinOutlined, BarChartOutlined, PieChartOutlined, LineChartOutlined,
  FieldTimeOutlined, HistoryOutlined, HourglassOutlined,
  FolderOutlined, FolderOpenOutlined, FileZipOutlined, FileImageOutlined,
  FilePdfOutlined, FileExcelOutlined, FileWordOutlined, FileMarkdownOutlined,
  BoxPlotOutlined, BuildOutlined, CiOutlined, NumberOutlined,
} from '@ant-design/icons-vue';

// ─── Lucide Icons ───
import {
  Zap, Activity, Flame, Target, Timer, Footprints, Bike, Mountain,
  Swords, Dumbbell, Trophy as LTrophy, Medal, Crosshair, Gauge,
  Briefcase, Building2, Users, CalendarDays, Wallet, BadgeCheck,
  ClipboardList, FileBarChart, Handshake, Banknote, Receipt, UserCheck,
  BookOpen, GraduationCap, Pencil, Lightbulb, FlaskConical, Microscope,
  ScrollText, FileCheck, NotebookPen, FileQuestion, Brain, Atom,
  Monitor, Smartphone, Wifi as LWifi, Cloud as LCloud, Server, Code2,
  Terminal, Shield, Link, HardDrive, Cpu, CircuitBoard,
  Database as LDatabase, GitBranch, Globe as LGlobe, Router,
  Home as LHome, ShoppingCart as LCart, Camera, Music, MapPin, Bell as LBell,
  Gift as LGift, Search, Clock as LClock, Package, Truck, Plane,
  Heart as LHeart, Star as LStar, Sun, Moon, CloudSun, Umbrella,
  TreePine, Leaf, Droplets, Wind, Thermometer, Snowflake,
  Eye, EyeOff, Lock as LLock, Unlock, Key as LKey, Fingerprint,
  Mail, Phone, MessageCircle, Send as LSend, Share2, Download, Upload,
  Scissors, Copy, Trash2, Save, Printer, QrCode,
  BarChart3, PieChart, TrendingUp, TrendingDown, ArrowUpDown,
  FolderOpen, File, FileText as LFileText, FileImage, FileArchive,
  Settings, Wrench, Hammer, Paintbrush, Palette, Layers,
  Navigation, Compass as LCompass, Map, Route, Locate,
  Wifi as LWifi2, Bluetooth, Radio, Podcast, Headphones, Volume2,
  Image, Video, Film, Clapperboard, Mic, Speaker,
  Coffee, UtensilsCrossed, Pizza, IceCreamCone, Cake, Cherry,
  Dog, Cat, Bird, Fish, Bug as LBug, Squirrel,
} from 'lucide-vue-next';

export interface IconEntry { comp: Component; label: string }

export const iconRegistry: Record<string, IconEntry> = {
  // ════════ Ant Design Outlined ════════
  thunderbolt: { comp: ThunderboltOutlined, label: '闪电' },
  fire: { comp: FireOutlined, label: '火焰' },
  trophy: { comp: TrophyOutlined, label: '奖杯' },
  aim: { comp: AimOutlined, label: '靶心' },
  dashboard: { comp: DashboardOutlined, label: '仪表盘' },
  heart: { comp: HeartOutlined, label: '心跳' },
  rocket: { comp: RocketOutlined, label: '火箭' },
  flag: { comp: FlagOutlined, label: '旗帜' },
  crown: { comp: CrownOutlined, label: '皇冠' },
  star: { comp: StarOutlined, label: '星星' },
  safety: { comp: SafetyCertificateOutlined, label: '安全' },
  smile: { comp: SmileOutlined, label: '笑脸' },
  radar: { comp: RadarChartOutlined, label: '雷达' },
  rise: { comp: RiseOutlined, label: '上升' },
  fall: { comp: FallOutlined, label: '下降' },
  interaction: { comp: InteractionOutlined, label: '交互' },
  bank: { comp: BankOutlined, label: '银行' },
  team: { comp: TeamOutlined, label: '团队' },
  solution: { comp: SolutionOutlined, label: '解决方案' },
  schedule: { comp: ScheduleOutlined, label: '日程' },
  idcard: { comp: IdcardOutlined, label: '证件' },
  account: { comp: AccountBookOutlined, label: '账本' },
  dollar: { comp: DollarOutlined, label: '金额' },
  fund: { comp: FundOutlined, label: '趋势' },
  audit: { comp: AuditOutlined, label: '审计' },
  contacts: { comp: ContactsOutlined, label: '通讯录' },
  reconciliation: { comp: ReconciliationOutlined, label: '对账' },
  insurance: { comp: InsuranceOutlined, label: '保险' },
  shopping: { comp: ShoppingOutlined, label: '购物' },
  moneycollect: { comp: MoneyCollectOutlined, label: '收款' },
  transaction: { comp: TransactionOutlined, label: '交易' },
  propertysafety: { comp: PropertySafetyOutlined, label: '资产安全' },
  read: { comp: ReadOutlined, label: '阅读' },
  book: { comp: BookOutlined, label: '书本' },
  filetext: { comp: FileTextOutlined, label: '文档' },
  highlight: { comp: HighlightOutlined, label: '高亮笔' },
  experiment: { comp: ExperimentOutlined, label: '实验' },
  bulb: { comp: BulbOutlined, label: '灯泡' },
  form: { comp: FormOutlined, label: '表单' },
  filedone: { comp: FileDoneOutlined, label: '已完成' },
  filesearch: { comp: FileSearchOutlined, label: '文件搜索' },
  edit: { comp: EditOutlined, label: '编辑' },
  profile: { comp: ProfileOutlined, label: '档案' },
  snippets: { comp: SnippetsOutlined, label: '片段' },
  copy: { comp: CopyOutlined, label: '复制' },
  scissor: { comp: ScissorOutlined, label: '剪刀' },
  laptop: { comp: LaptopOutlined, label: '电脑' },
  mobile: { comp: MobileOutlined, label: '手机' },
  cloud: { comp: CloudOutlined, label: '云' },
  global: { comp: GlobalOutlined, label: '全球' },
  api: { comp: ApiOutlined, label: '接口' },
  setting: { comp: SettingOutlined, label: '设置' },
  tool: { comp: ToolOutlined, label: '工具' },
  lock: { comp: LockOutlined, label: '锁' },
  key: { comp: KeyOutlined, label: '钥匙' },
  wifi: { comp: WifiOutlined, label: '信号' },
  database: { comp: DatabaseOutlined, label: '数据库' },
  code: { comp: CodeOutlined, label: '代码' },
  bug: { comp: BugOutlined, label: '调试' },
  deployment: { comp: DeploymentUnitOutlined, label: '部署' },
  cluster: { comp: ClusterOutlined, label: '集群' },
  console: { comp: ConsoleSqlOutlined, label: '控制台' },
  cloudserver: { comp: CloudServerOutlined, label: '云服务器' },
  clouddown: { comp: CloudDownloadOutlined, label: '云下载' },
  cloudup: { comp: CloudUploadOutlined, label: '云上传' },
  hdd: { comp: HddOutlined, label: '硬盘' },
  desktop: { comp: DesktopOutlined, label: '桌面' },
  tablet: { comp: TabletOutlined, label: '平板' },
  usb: { comp: UsbOutlined, label: 'USB' },
  appstore: { comp: AppstoreOutlined, label: '应用' },
  shop: { comp: ShopOutlined, label: '商店' },
  cart: { comp: ShoppingCartOutlined, label: '购物车' },
  car: { comp: CarOutlined, label: '汽车' },
  home: { comp: HomeOutlined, label: '主页' },
  gift: { comp: GiftOutlined, label: '礼物' },
  bell: { comp: BellOutlined, label: '铃铛' },
  tag: { comp: TagOutlined, label: '标签' },
  picture: { comp: PictureOutlined, label: '图片' },
  sound: { comp: SoundOutlined, label: '音乐' },
  video: { comp: VideoCameraOutlined, label: '视频' },
  environment: { comp: EnvironmentOutlined, label: '定位' },
  compass: { comp: CompassOutlined, label: '指南针' },
  calendar: { comp: CalendarOutlined, label: '日历' },
  clock: { comp: ClockCircleOutlined, label: '时钟' },
  send: { comp: SendOutlined, label: '发送' },
  camera: { comp: CameraOutlined, label: '相机' },
  printer: { comp: PrinterOutlined, label: '打印' },
  wallet: { comp: WalletOutlined, label: '钱包' },
  skin: { comp: SkinOutlined, label: '皮肤' },
  medicine: { comp: MedicineBoxOutlined, label: '医疗' },
  coffee: { comp: CoffeeOutlined, label: '咖啡' },
  carryout: { comp: CarryOutOutlined, label: '外卖' },
  scan: { comp: ScanOutlined, label: '扫描' },
  alert: { comp: AlertOutlined, label: '警告' },
  checkcircle: { comp: CheckCircleOutlined, label: '成功' },
  closecircle: { comp: CloseCircleOutlined, label: '失败' },
  infocircle: { comp: InfoCircleOutlined, label: '信息' },
  exclamation: { comp: ExclamationCircleOutlined, label: '感叹' },
  question: { comp: QuestionCircleOutlined, label: '疑问' },
  stop: { comp: StopOutlined, label: '停止' },
  user: { comp: UserOutlined, label: '用户' },
  usergroup: { comp: UsergroupAddOutlined, label: '加人' },
  man: { comp: ManOutlined, label: '男' },
  woman: { comp: WomanOutlined, label: '女' },
  phone: { comp: PhoneOutlined, label: '电话' },
  mail: { comp: MailOutlined, label: '邮件' },
  message: { comp: MessageOutlined, label: '消息' },
  notification: { comp: NotificationOutlined, label: '通知' },
  like: { comp: LikeOutlined, label: '点赞' },
  dislike: { comp: DislikeOutlined, label: '踩' },
  sharealt: { comp: ShareAltOutlined, label: '分享' },
  link: { comp: LinkOutlined, label: '链接' },
  delete: { comp: DeleteOutlined, label: '删除' },
  save: { comp: SaveOutlined, label: '保存' },
  download: { comp: DownloadOutlined, label: '下载' },
  upload: { comp: UploadOutlined, label: '上传' },
  swap: { comp: SwapOutlined, label: '交换' },
  sync: { comp: SyncOutlined, label: '同步' },
  filter: { comp: FilterOutlined, label: '筛选' },
  search: { comp: SearchOutlined, label: '搜索' },
  zoomin: { comp: ZoomInOutlined, label: '放大' },
  zoomout: { comp: ZoomOutOutlined, label: '缩小' },
  eye: { comp: EyeOutlined, label: '查看' },
  eyeinvisible: { comp: EyeInvisibleOutlined, label: '隐藏' },
  paperclip: { comp: PaperClipOutlined, label: '附件' },
  pushpin: { comp: PushpinOutlined, label: '图钉' },
  barchart: { comp: BarChartOutlined, label: '柱状图' },
  piechart: { comp: PieChartOutlined, label: '饼图' },
  linechart: { comp: LineChartOutlined, label: '折线图' },
  fieldtime: { comp: FieldTimeOutlined, label: '时间字段' },
  history: { comp: HistoryOutlined, label: '历史' },
  hourglass: { comp: HourglassOutlined, label: '沙漏' },
  folder: { comp: FolderOutlined, label: '文件夹' },
  folderopen: { comp: FolderOpenOutlined, label: '打开文件夹' },
  filezip: { comp: FileZipOutlined, label: '压缩包' },
  fileimage: { comp: FileImageOutlined, label: '图片文件' },
  filepdf: { comp: FilePdfOutlined, label: 'PDF' },
  fileexcel: { comp: FileExcelOutlined, label: 'Excel' },
  fileword: { comp: FileWordOutlined, label: 'Word' },
  filemd: { comp: FileMarkdownOutlined, label: 'Markdown' },
  build: { comp: BuildOutlined, label: '构建' },
  number: { comp: NumberOutlined, label: '数字' },
  // ════════ Ant Design Filled ════════
  'thunderbolt-f': { comp: ThunderboltFilled, label: '闪电·实底' },
  'fire-f': { comp: FireFilled, label: '火焰·实底' },
  'trophy-f': { comp: TrophyFilled, label: '奖杯·实底' },
  'heart-f': { comp: HeartFilled, label: '心跳·实底' },
  'star-f': { comp: StarFilled, label: '星星·实底' },
  'crown-f': { comp: CrownFilled, label: '皇冠·实底' },
  'bell-f': { comp: BellFilled, label: '铃铛·实底' },
  'gift-f': { comp: GiftFilled, label: '礼物·实底' },
  'home-f': { comp: HomeFilled, label: '主页·实底' },
  'shop-f': { comp: ShopFilled, label: '商店·实底' },
  // ════════ Lucide ════════
  'l-zap': { comp: Zap, label: '闪电' },
  'l-activity': { comp: Activity, label: '活动' },
  'l-flame': { comp: Flame, label: '火焰' },
  'l-target': { comp: Target, label: '靶心' },
  'l-timer': { comp: Timer, label: '计时器' },
  'l-footprints': { comp: Footprints, label: '脚印' },
  'l-bike': { comp: Bike, label: '自行车' },
  'l-mountain': { comp: Mountain, label: '山' },
  'l-swords': { comp: Swords, label: '剑' },
  'l-dumbbell': { comp: Dumbbell, label: '哑铃' },
  'l-trophy': { comp: LTrophy, label: '奖杯' },
  'l-medal': { comp: Medal, label: '奖牌' },
  'l-crosshair': { comp: Crosshair, label: '准星' },
  'l-gauge': { comp: Gauge, label: '仪表' },
  'l-briefcase': { comp: Briefcase, label: '公文包' },
  'l-building': { comp: Building2, label: '大楼' },
  'l-users': { comp: Users, label: '用户组' },
  'l-calendardays': { comp: CalendarDays, label: '日历' },
  'l-wallet': { comp: Wallet, label: '钱包' },
  'l-badgecheck': { comp: BadgeCheck, label: '认证' },
  'l-clipboard': { comp: ClipboardList, label: '清单' },
  'l-filebarchart': { comp: FileBarChart, label: '报表' },
  'l-handshake': { comp: Handshake, label: '握手' },
  'l-banknote': { comp: Banknote, label: '钞票' },
  'l-receipt': { comp: Receipt, label: '收据' },
  'l-usercheck': { comp: UserCheck, label: '审核' },
  'l-bookopen': { comp: BookOpen, label: '打开的书' },
  'l-graduation': { comp: GraduationCap, label: '毕业帽' },
  'l-pencil': { comp: Pencil, label: '铅笔' },
  'l-lightbulb': { comp: Lightbulb, label: '灯泡' },
  'l-flask': { comp: FlaskConical, label: '烧瓶' },
  'l-microscope': { comp: Microscope, label: '显微镜' },
  'l-scroll': { comp: ScrollText, label: '卷轴' },
  'l-filecheck': { comp: FileCheck, label: '文件勾选' },
  'l-notebookpen': { comp: NotebookPen, label: '笔记本' },
  'l-filequestion': { comp: FileQuestion, label: '文件问号' },
  'l-brain': { comp: Brain, label: '大脑' },
  'l-atom': { comp: Atom, label: '原子' },
  'l-monitor': { comp: Monitor, label: '显示器' },
  'l-smartphone': { comp: Smartphone, label: '手机' },
  'l-wifi': { comp: LWifi, label: '无线' },
  'l-cloud': { comp: LCloud, label: '云' },
  'l-server': { comp: Server, label: '服务器' },
  'l-code': { comp: Code2, label: '代码' },
  'l-terminal': { comp: Terminal, label: '终端' },
  'l-shield': { comp: Shield, label: '盾牌' },
  'l-link': { comp: Link, label: '链接' },
  'l-harddrive': { comp: HardDrive, label: '硬盘' },
  'l-cpu': { comp: Cpu, label: 'CPU' },
  'l-circuit': { comp: CircuitBoard, label: '电路板' },
  'l-database': { comp: LDatabase, label: '数据库' },
  'l-git': { comp: GitBranch, label: 'Git' },
  'l-globe': { comp: LGlobe, label: '地球' },
  'l-router': { comp: Router, label: '路由器' },
  'l-home': { comp: LHome, label: '主页' },
  'l-cart': { comp: LCart, label: '购物车' },
  'l-camera': { comp: Camera, label: '相机' },
  'l-music': { comp: Music, label: '音乐' },
  'l-mappin': { comp: MapPin, label: '定位' },
  'l-bell': { comp: LBell, label: '铃铛' },
  'l-gift': { comp: LGift, label: '礼物' },
  'l-search': { comp: Search, label: '搜索' },
  'l-clock': { comp: LClock, label: '时钟' },
  'l-package': { comp: Package, label: '包裹' },
  'l-truck': { comp: Truck, label: '卡车' },
  'l-plane': { comp: Plane, label: '飞机' },
  'l-heart': { comp: LHeart, label: '红心' },
  'l-star': { comp: LStar, label: '星星' },
  'l-sun': { comp: Sun, label: '太阳' },
  'l-moon': { comp: Moon, label: '月亮' },
  'l-cloudsun': { comp: CloudSun, label: '多云' },
  'l-umbrella': { comp: Umbrella, label: '雨伞' },
  'l-tree': { comp: TreePine, label: '松树' },
  'l-leaf': { comp: Leaf, label: '叶子' },
  'l-droplets': { comp: Droplets, label: '水滴' },
  'l-wind': { comp: Wind, label: '风' },
  'l-thermometer': { comp: Thermometer, label: '温度计' },
  'l-snowflake': { comp: Snowflake, label: '雪花' },
  'l-eye': { comp: Eye, label: '查看' },
  'l-eyeoff': { comp: EyeOff, label: '隐藏' },
  'l-lock': { comp: LLock, label: '锁' },
  'l-unlock': { comp: Unlock, label: '解锁' },
  'l-key': { comp: LKey, label: '钥匙' },
  'l-fingerprint': { comp: Fingerprint, label: '指纹' },
  'l-mail': { comp: Mail, label: '邮件' },
  'l-phone': { comp: Phone, label: '电话' },
  'l-messagecircle': { comp: MessageCircle, label: '聊天' },
  'l-send': { comp: LSend, label: '发送' },
  'l-share': { comp: Share2, label: '分享' },
  'l-download': { comp: Download, label: '下载' },
  'l-upload': { comp: Upload, label: '上传' },
  'l-scissors': { comp: Scissors, label: '剪刀' },
  'l-copy': { comp: Copy, label: '复制' },
  'l-trash': { comp: Trash2, label: '删除' },
  'l-save': { comp: Save, label: '保存' },
  'l-printer': { comp: Printer, label: '打印' },
  'l-qrcode': { comp: QrCode, label: '二维码' },
  'l-barchart': { comp: BarChart3, label: '柱状图' },
  'l-piechart': { comp: PieChart, label: '饼图' },
  'l-trendup': { comp: TrendingUp, label: '上升' },
  'l-trenddown': { comp: TrendingDown, label: '下降' },
  'l-folderopen': { comp: FolderOpen, label: '打开文件夹' },
  'l-file': { comp: File, label: '文件' },
  'l-filetext': { comp: LFileText, label: '文档' },
  'l-fileimage': { comp: FileImage, label: '图片文件' },
  'l-filearchive': { comp: FileArchive, label: '压缩包' },
  'l-settings': { comp: Settings, label: '设置' },
  'l-wrench': { comp: Wrench, label: '扳手' },
  'l-hammer': { comp: Hammer, label: '锤子' },
  'l-paintbrush': { comp: Paintbrush, label: '画笔' },
  'l-palette': { comp: Palette, label: '调色板' },
  'l-layers': { comp: Layers, label: '图层' },
  'l-navigation': { comp: Navigation, label: '导航' },
  'l-compass': { comp: LCompass, label: '指南针' },
  'l-map': { comp: Map, label: '地图' },
  'l-route': { comp: Route, label: '路线' },
  'l-locate': { comp: Locate, label: '定位' },
  'l-headphones': { comp: Headphones, label: '耳机' },
  'l-volume': { comp: Volume2, label: '音量' },
  'l-image': { comp: Image, label: '图片' },
  'l-video': { comp: Video, label: '视频' },
  'l-film': { comp: Film, label: '胶片' },
  'l-mic': { comp: Mic, label: '麦克风' },
  'l-coffee': { comp: Coffee, label: '咖啡' },
  'l-utensils': { comp: UtensilsCrossed, label: '餐具' },
  'l-pizza': { comp: Pizza, label: '披萨' },
  'l-cake': { comp: Cake, label: '蛋糕' },
  'l-dog': { comp: Dog, label: '狗' },
  'l-cat': { comp: Cat, label: '猫' },
  'l-bird': { comp: Bird, label: '鸟' },
  'l-fish': { comp: Fish, label: '鱼' },
  'l-bug': { comp: LBug, label: '虫子' },
};

// 分类配置
export const iconGroups: Record<string, { key: string; label: string; icons: string[] }[]> = {
  antd: [
    { key: 'sport', label: '运动', icons: ['thunderbolt', 'fire', 'trophy', 'aim', 'dashboard', 'heart', 'rocket', 'flag', 'crown', 'star', 'safety', 'smile', 'radar', 'rise', 'fall', 'interaction'] },
    { key: 'work', label: '工作', icons: ['bank', 'team', 'solution', 'schedule', 'idcard', 'account', 'dollar', 'fund', 'audit', 'contacts', 'reconciliation', 'insurance', 'shopping', 'moneycollect', 'transaction', 'propertysafety'] },
    { key: 'study', label: '学习', icons: ['read', 'book', 'filetext', 'highlight', 'experiment', 'bulb', 'form', 'filedone', 'filesearch', 'edit', 'profile', 'snippets', 'copy', 'scissor'] },
    { key: 'tech', label: '技术', icons: ['laptop', 'mobile', 'cloud', 'global', 'api', 'setting', 'tool', 'lock', 'key', 'wifi', 'database', 'code', 'bug', 'deployment', 'cluster', 'console', 'cloudserver', 'clouddown', 'cloudup', 'hdd', 'desktop', 'tablet', 'usb'] },
    { key: 'common', label: '通用', icons: ['appstore', 'shop', 'cart', 'car', 'home', 'gift', 'bell', 'tag', 'picture', 'sound', 'video', 'environment', 'compass', 'calendar', 'clock', 'send', 'camera', 'printer', 'wallet', 'skin', 'medicine', 'coffee', 'carryout', 'scan'] },
    { key: 'status', label: '状态', icons: ['alert', 'checkcircle', 'closecircle', 'infocircle', 'exclamation', 'question', 'stop', 'like', 'dislike'] },
    { key: 'people', label: '用户', icons: ['user', 'usergroup', 'man', 'woman', 'phone', 'mail', 'message', 'notification'] },
    { key: 'action', label: '操作', icons: ['search', 'filter', 'delete', 'save', 'download', 'upload', 'swap', 'sync', 'sharealt', 'link', 'eye', 'eyeinvisible', 'paperclip', 'pushpin', 'zoomin', 'zoomout'] },
    { key: 'chart', label: '图表', icons: ['barchart', 'piechart', 'linechart', 'fieldtime', 'history', 'hourglass'] },
    { key: 'file', label: '文件', icons: ['folder', 'folderopen', 'filezip', 'fileimage', 'filepdf', 'fileexcel', 'fileword', 'filemd', 'build', 'number'] },
    { key: 'filled', label: '实底', icons: ['thunderbolt-f', 'fire-f', 'trophy-f', 'heart-f', 'star-f', 'crown-f', 'bell-f', 'gift-f', 'home-f', 'shop-f'] },
  ],
  lucide: [
    { key: 'l-sport', label: '运动', icons: ['l-zap', 'l-activity', 'l-flame', 'l-target', 'l-timer', 'l-footprints', 'l-bike', 'l-mountain', 'l-swords', 'l-dumbbell', 'l-trophy', 'l-medal', 'l-crosshair', 'l-gauge'] },
    { key: 'l-work', label: '工作', icons: ['l-briefcase', 'l-building', 'l-users', 'l-calendardays', 'l-wallet', 'l-badgecheck', 'l-clipboard', 'l-filebarchart', 'l-handshake', 'l-banknote', 'l-receipt', 'l-usercheck'] },
    { key: 'l-study', label: '学习', icons: ['l-bookopen', 'l-graduation', 'l-pencil', 'l-lightbulb', 'l-flask', 'l-microscope', 'l-scroll', 'l-filecheck', 'l-notebookpen', 'l-filequestion', 'l-brain', 'l-atom'] },
    { key: 'l-tech', label: '技术', icons: ['l-monitor', 'l-smartphone', 'l-wifi', 'l-cloud', 'l-server', 'l-code', 'l-terminal', 'l-shield', 'l-link', 'l-harddrive', 'l-cpu', 'l-circuit', 'l-database', 'l-git', 'l-globe', 'l-router'] },
    { key: 'l-common', label: '通用', icons: ['l-home', 'l-cart', 'l-camera', 'l-music', 'l-mappin', 'l-bell', 'l-gift', 'l-search', 'l-clock', 'l-package', 'l-truck', 'l-plane', 'l-heart', 'l-star'] },
    { key: 'l-nature', label: '自然', icons: ['l-sun', 'l-moon', 'l-cloudsun', 'l-umbrella', 'l-tree', 'l-leaf', 'l-droplets', 'l-wind', 'l-thermometer', 'l-snowflake'] },
    { key: 'l-security', label: '安全', icons: ['l-eye', 'l-eyeoff', 'l-lock', 'l-unlock', 'l-key', 'l-fingerprint'] },
    { key: 'l-comm', label: '通讯', icons: ['l-mail', 'l-phone', 'l-messagecircle', 'l-send', 'l-share', 'l-headphones', 'l-volume'] },
    { key: 'l-action', label: '操作', icons: ['l-download', 'l-upload', 'l-scissors', 'l-copy', 'l-trash', 'l-save', 'l-printer', 'l-qrcode'] },
    { key: 'l-chart', label: '图表', icons: ['l-barchart', 'l-piechart', 'l-trendup', 'l-trenddown'] },
    { key: 'l-file', label: '文件', icons: ['l-folderopen', 'l-file', 'l-filetext', 'l-fileimage', 'l-filearchive'] },
    { key: 'l-design', label: '设计', icons: ['l-settings', 'l-wrench', 'l-hammer', 'l-paintbrush', 'l-palette', 'l-layers'] },
    { key: 'l-nav', label: '导航', icons: ['l-navigation', 'l-compass', 'l-map', 'l-route', 'l-locate'] },
    { key: 'l-media', label: '媒体', icons: ['l-image', 'l-video', 'l-film', 'l-mic'] },
    { key: 'l-food', label: '食物', icons: ['l-coffee', 'l-utensils', 'l-pizza', 'l-cake'] },
    { key: 'l-animal', label: '动物', icons: ['l-dog', 'l-cat', 'l-bird', 'l-fish', 'l-bug'] },
  ],
};
</script>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Popover, Input, Tabs, TabPane, Tooltip, Segmented } from 'ant-design-vue';
import { SearchOutlined as SearchIcon } from '@ant-design/icons-vue';

const modelValue = defineModel<string>({ default: '' });
const keyword = ref('');
const visible = ref(false);
const activeLib = ref<string>('antd');
const activeTab = ref('sport');

// 切换图标库时重置 Tab
watch(activeLib, (lib) => {
  const groups = iconGroups[lib];
  if (groups && groups.length > 0) activeTab.value = groups[0]!.key;
});

const currentGroups = computed(() => iconGroups[activeLib.value] || []);

const filteredIcons = computed(() => {
  const group = currentGroups.value.find((g) => g.key === activeTab.value);
  if (!group) return [];
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return group.icons;
  return group.icons.filter((k) => {
    const entry = iconRegistry[k];
    return entry && (k.includes(kw) || entry.label.includes(kw));
  });
});

// 搜索时跨分类全局匹配
const globalSearchResults = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return null;
  const all: string[] = [];
  for (const group of currentGroups.value) {
    for (const k of group.icons) {
      const entry = iconRegistry[k];
      if (entry && (k.includes(kw) || entry.label.includes(kw))) all.push(k);
    }
  }
  return all;
});

const displayIcons = computed(() => globalSearchResults.value ?? filteredIcons.value);

function selectIcon(key: string) {
  modelValue.value = key;
  visible.value = false;
  keyword.value = '';
}

function clearIcon() {
  modelValue.value = '';
}
</script>

<template>
  <Popover
    v-model:open="visible"
    trigger="click"
    placement="bottomLeft"
  >
    <template #content>
      <div style="width: 380px;">
        <div class="flex items-center gap-2 mb-2">
          <Input
            v-model:value="keyword"
            placeholder="搜索图标..."
            size="small"
            allow-clear
            class="flex-1"
          >
            <template #prefix><SearchIcon /></template>
          </Input>
          <Segmented v-model:value="activeLib" size="small" :options="[{ value: 'antd', label: 'Ant Design' }, { value: 'lucide', label: 'Lucide' }]" />
        </div>
        <Tabs v-if="!globalSearchResults" v-model:activeKey="activeTab" size="small" :tab-bar-style="{ marginBottom: '8px' }">
          <TabPane v-for="g in currentGroups" :key="g.key" :tab="g.label" />
        </Tabs>
        <div v-else class="text-xs text-gray-400 mb-2">搜索结果 ({{ displayIcons.length }})</div>
        <div class="grid grid-cols-8 gap-0.5 max-h-[220px] overflow-y-auto" style="scrollbar-width: thin;">
          <Tooltip v-for="k in displayIcons" :key="k" :title="iconRegistry[k]?.label">
            <div
              class="flex items-center justify-center rounded cursor-pointer h-9 w-9 text-base transition-colors"
              :class="modelValue === k ? 'bg-blue-50 text-blue-500 ring-1 ring-blue-300' : 'text-gray-600 hover:bg-gray-50'"
              @click="selectIcon(k)"
            >
              <component :is="iconRegistry[k]?.comp" />
            </div>
          </Tooltip>
        </div>
        <div v-if="displayIcons.length === 0" class="text-center text-gray-400 py-4 text-sm">
          没有匹配的图标
        </div>
      </div>
    </template>

    <div
      class="flex items-center border rounded px-3 py-1.5 cursor-pointer hover:border-blue-400 transition-colors min-h-[32px]"
      :class="visible ? 'border-blue-400 shadow-sm' : 'border-gray-300'"
    >
      <component :is="iconRegistry[modelValue]?.comp" v-if="modelValue && iconRegistry[modelValue]" class="text-base text-gray-600 mr-2" />
      <span v-if="modelValue && iconRegistry[modelValue]" class="text-sm text-gray-500 flex-1">{{ iconRegistry[modelValue]?.label }}</span>
      <span v-else class="text-sm text-gray-400 flex-1">点击选择图标</span>
      <span
        v-if="modelValue"
        class="text-gray-400 hover:text-gray-600 ml-2 text-xs cursor-pointer"
        @click.stop="clearIcon"
      >✕</span>
    </div>
  </Popover>
</template>

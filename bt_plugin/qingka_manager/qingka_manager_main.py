# coding: utf-8
import sys, os, json, time, re, hashlib
panelPath = '/www/server/panel'
if panelPath not in sys.path:
    sys.path.insert(0, panelPath)
import public

class qingka_manager_main:
    __plugin_path = '/www/server/panel/plugin/qingka_manager'
    __site_dir = '/www/wwwroot/qingka'
    __admin_dir = '/www/wwwroot/qingka/admin'
    __mall_dir = '/www/wwwroot/qingka/mall'
    __go_dir = '/www/wwwroot/qingka/go-api'
    __bin_name = 'server'
    __log_file = '/www/wwwroot/qingka/go-api/go-api.log'
    __pid_file = '/www/wwwroot/qingka/go-api/go-api.pid'
    __config_file = '/www/wwwroot/qingka/go-api/config/config.yaml'
    __version_file = '/www/wwwroot/qingka/version.json'
    __service_name = 'qingka-api'
    __update_server = 'https://raw.githubusercontent.com/Jinxuans/LD_Resources/refs/heads/main'


    def __init__(self):
        pass

    def _get_site_root(self):
        return self.__admin_dir

    def _get_mall_root(self):
        return self.__mall_dir

    # ==================== 输入校验 ====================

    def _kill_process(self, pid):
        """安全杀掉指定 PID 进程"""
        if not pid or not os.path.exists('/proc/%s' % pid):
            return
        public.ExecShell('kill %s' % pid)
        time.sleep(1)
        if os.path.exists('/proc/%s' % pid):
            public.ExecShell('kill -9 %s' % pid)
            time.sleep(0.5)

    def _kill_process_by_port(self, port):
        """根据端口杀掉相关进程"""
        if not port: return
        old_pids = public.ExecShell("lsof -i :%s -t 2>/dev/null" % port)[0].strip()
        if old_pids:
            for p in old_pids.split('\n'):
                p = p.strip()
                if p and p.isdigit():
                    public.ExecShell('kill -9 %s' % p)
                    public.WriteLog('qingka_manager', '已杀掉占用端口 %s 的旧进程 PID: %s' % (port, p))
            time.sleep(1)

    def _safe_pid(self, pid):
        if pid and str(pid).strip().isdigit():
            return str(pid).strip()
        return None

    def _safe_int(self, val, default=200, min_val=1, max_val=10000):
        try:
            v = int(val)
            return max(min_val, min(v, max_val))
        except (ValueError, TypeError):
            return default

    def _safe_db_param(self, val):
        """过滤数据库参数中的危险字符"""
        return re.sub(r'[;`$\\]', '', str(val).strip())

    # ==================== 状态 ==

    def get_status(self, args):
        pid = self._get_pid()
        running = pid and os.path.exists('/proc/%s' % pid)
        data = {
            'status': bool(running),
            'pid': pid if running else 0,
            'mem': '', 'uptime': '',
            'version': self._get_version(),
            'install_path': self.__go_dir,
            'bin_exists': os.path.isfile(os.path.join(self.__go_dir, self.__bin_name)),
            'domain': self._get_domain()
        }
        if running:
            try:
                mem = public.ExecShell("cat /proc/%s/status | grep VmRSS | awk '{print $2}'" % pid)[0].strip()
                if mem and mem.isdigit(): data['mem'] = str(round(int(mem) / 1024, 1)) + ' MB'
            except Exception as e:
                public.WriteLog('qingka_manager', '获取内存信息失败: %s' % str(e))
            try:
                data['uptime'] = public.ExecShell("ps -p %s -o etime= | xargs" % pid)[0].strip()
            except Exception as e:
                public.WriteLog('qingka_manager', '获取运行时间失败: %s' % str(e))
        return public.returnMsg(True, json.dumps(data))

    def get_init_status(self, args):
        data = {
            'installed': os.path.isfile(os.path.join(self.__go_dir, self.__bin_name)) and os.path.isfile(self.__config_file),
            'has_binary': os.path.isfile(os.path.join(self.__go_dir, self.__bin_name)),
            'has_config': os.path.isfile(self.__config_file),
            'has_domain': bool(self._get_domain())
        }
        return public.returnMsg(True, json.dumps(data))

    # ==================== 服务控制 ====================

    def _is_systemd_registered(self, service_name):
        return os.path.isfile('/etc/systemd/system/%s.service' % service_name)

    def start(self, args):
        pid = self._get_pid()
        if pid and os.path.exists('/proc/%s' % pid):
            return public.returnMsg(False, '服务已在运行中，PID: %s' % pid)
        bin_path = os.path.join(self.__go_dir, self.__bin_name)
        if not os.path.isfile(bin_path):
            return public.returnMsg(False, '可执行文件不存在')
        # 自动杀掉占用端口的旧进程
        port = self._get_port()
        if port:
            self._kill_process_by_port(port)
        os.chmod(bin_path, 0o755)
        # 优先使用 systemd
        if self._is_systemd_registered('qingka-api'):
            public.ExecShell('systemctl start qingka-api.service')
            time.sleep(2)
            pid = self._get_pid()
            if pid and os.path.exists('/proc/%s' % pid):
                return public.returnMsg(True, '启动成功（systemd），PID: %s' % pid)
            return public.returnMsg(False, '启动失败，请查看日志')
        cmd = 'cd %s && nohup ./%s > %s 2>&1 & echo $!' % (self.__go_dir, self.__bin_name, self.__log_file)
        result = public.ExecShell(cmd)[0].strip()
        pid = self._safe_pid(result)
        if pid:
            self._save_pid(pid)
            time.sleep(2)
            if os.path.exists('/proc/%s' % pid):
                return public.returnMsg(True, '启动成功，PID: %s' % pid)
        return public.returnMsg(False, '启动失败，请查看日志')

    def stop(self, args):
        pid = self._get_pid()
        if not pid:
            return public.returnMsg(False, '服务未运行')
        
        # 优先使用 systemd
        if self._is_systemd_registered('qingka-api'):
            public.ExecShell('systemctl stop qingka-api.service')
        
        self._kill_process(pid)
        self._del_pid()
        
        if os.path.exists('/proc/%s' % pid):
            return public.returnMsg(False, '停止失败，进程仍在运行，PID: %s' % pid)
        return public.returnMsg(True, '已停止')

    def restart(self, args):
        self.stop(args)
        time.sleep(1)
        return self.start(args)

    def _start_no_auth(self):
        """内部启动（跳过授权检查），仅供 init_install 等内部流程使用"""
        pid = self._get_pid()
        if pid and os.path.exists('/proc/%s' % pid):
            return
        bin_path = os.path.join(self.__go_dir, self.__bin_name)
        if not os.path.isfile(bin_path):
            return
        port = self._get_port()
        if port:
            self._kill_process_by_port(port)
        os.chmod(bin_path, 0o755)
        cmd = 'cd %s && nohup ./%s > %s 2>&1 & echo $!' % (self.__go_dir, self.__bin_name, self.__log_file)
        result = public.ExecShell(cmd)[0].strip()
        pid = self._safe_pid(result)
        if pid:
            self._save_pid(pid)

    def _restart_go_service(self):
        """内部重启 Go 服务"""
        pid = self._get_pid()
        if pid:
            self._kill_process(pid)
            self._del_pid()
            time.sleep(1)
        self._start_no_auth()

    # ==================== 联合操作 ====================

    def restart_all(self, args):
        """重启 Go 服务"""
        self.stop(args)
        time.sleep(1)
        go_res = self.start(args)
        go_ok = go_res.get('status', False) if isinstance(go_res, dict) else False
        msg = go_res.get('msg', '') if isinstance(go_res, dict) else str(go_res)
        return public.returnMsg(go_ok, msg)

    # ==================== 健康检查 ====================

    def health_check(self, args):
        """检查 Go 服务是否存活，挂掉则自动拉起"""
        results = []
        # Go 服务检查
        go_pid = self._get_pid()
        go_running = go_pid and os.path.exists('/proc/%s' % go_pid)
        if not go_running:
            bin_path = os.path.join(self.__go_dir, self.__bin_name)
            if os.path.isfile(bin_path):
                # 仅在非 systemd 模式下自动拉起，systemd 会自动重启
                if not self._is_systemd_registered('qingka-api'):
                    self.start(args)
                    results.append('Go 服务已自动拉起')
                    public.WriteLog('qingka_manager', '健康检查：Go 服务异常退出，已自动重启')
            else:
                results.append('Go 二进制不存在，跳过')
        else:
            results.append('Go 服务正常 (PID: %s)' % go_pid)

        return public.returnMsg(True, ' | '.join(results))

    # ==================== Systemd 服务管理 ====================

    def setup_systemd(self, args):
        """注册 Go 为 systemd 服务，实现开机自启和崩溃自动重启"""
        results = []
        # Go 服务
        go_service = '''[Unit]
Description=QingKa Go API Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s/%s
Restart=always
RestartSec=5
LimitNOFILE=65535
StandardOutput=append:%s
StandardError=append:%s

[Install]
WantedBy=multi-user.target
''' % (self.__go_dir, self.__go_dir, self.__bin_name, self.__log_file, self.__log_file)
        public.writeFile('/etc/systemd/system/qingka-api.service', go_service)
        results.append('Go 服务已注册')

        # reload + enable
        public.ExecShell('systemctl daemon-reload')
        public.ExecShell('systemctl enable qingka-api.service 2>/dev/null')
        results.append('已设置开机自启')

        return public.returnMsg(True, ' | '.join(results))

    def remove_systemd(self, args):
        """移除 systemd 服务"""
        public.ExecShell('systemctl stop qingka-api.service 2>/dev/null')
        public.ExecShell('systemctl disable qingka-api.service 2>/dev/null')
        for f in ['/etc/systemd/system/qingka-api.service']:
            if os.path.isfile(f):
                os.remove(f)
        public.ExecShell('systemctl daemon-reload')
        return public.returnMsg(True, 'Systemd 服务已移除')

    def get_systemd_status(self, args):
        """获取 systemd 服务状态"""
        go_enabled = os.path.isfile('/etc/systemd/system/qingka-api.service')
        data = {
            'go_registered': go_enabled,
        }
        if go_enabled:
            result = public.ExecShell('systemctl is-active qingka-api.service 2>/dev/null')[0].strip()
            data['go_active'] = result == 'active'
        return public.returnMsg(True, json.dumps(data))

    # ==================== 日志 ====================

    def get_log(self, args):
        if not os.path.isfile(self.__log_file):
            return public.returnMsg(True, '暂无日志')
        try:
            n = self._safe_int(getattr(args, 'line_count', 200), default=200, max_val=5000)
            log = public.ExecShell('tail -n %d %s' % (n, self.__log_file))[0]
            return public.returnMsg(True, log)
        except Exception as e:
            public.WriteLog('qingka_manager', '读取日志失败: %s' % str(e))
            return public.returnMsg(False, '读取日志失败')

    # ==================== 配置 ====================

    def get_config(self, args):
        if not os.path.isfile(self.__config_file):
            return public.returnMsg(False, '配置文件不存在')
        return public.returnMsg(True, public.readFile(self.__config_file))

    def save_config(self, args):
        if not hasattr(args, 'config'):
            return public.returnMsg(False, '未提供配置内容')
        try:
            public.writeFile(self.__config_file, args.config)
            return public.returnMsg(True, '配置已保存，重启后生效')
        except Exception as e:
            public.WriteLog('qingka_manager', '保存配置失败: %s' % str(e))
            return public.returnMsg(False, '保存配置失败: %s' % str(e))

    def test_db(self, args):
        db_info = self._read_db_config()
        if not db_info:
            return public.returnMsg(False, '无法读取数据库配置')
        result = self._safe_mysql_cmd(db_info['user'], db_info['pass'], db_info['name'])
        output = (str(result[0]) + str(result[1])).strip()
        if 'Access denied' in output or 'ERROR' in output:
            err = output.split('ERROR')[-1].strip() if 'ERROR' in output else output
            return public.returnMsg(False, '数据库连接失败: %s。请到「配置文件」标签页修改 database 部分的 user 和 password' % err[:200])
        return public.returnMsg(True, '数据库连接成功')
    # ==================== 域名管理 ====================

    def setup_domain(self, args):
        domain = getattr(args, 'domain', '').strip()
        if not domain:
            return public.returnMsg(False, '请输入域名')
        if not re.match(r'^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$', domain) or len(domain) > 253:
            return public.returnMsg(False, '域名格式不正确')

        site_root = self._get_site_root()
        mall_root = self._get_mall_root()
        site_base = self.__site_dir
        conf_dir = '/www/server/panel/vhost/nginx'
        conf_file = os.path.join(conf_dir, '%s.conf' % domain)

        # 创建统一部署目录
        for d in (site_base, site_root, mall_root):
            os.makedirs(d, exist_ok=True)
        public.ExecShell('chmod -R 755 ' + site_base)
        public.ExecShell('chown -R www:www ' + site_base)

        # 注册到宝塔面板数据库
        if not public.M('sites').where('name=?', (domain,)).count():
            pid = public.M('sites').add('name,path,status,ps,type_id,addtime,project_type',
                (domain, site_base, '1', '青卡管理系统', 0, time.strftime('%Y-%m-%d %H:%M:%S'), 'PHP'))
            public.M('domain').add('pid,name,port,addtime',
                (pid, domain, 80, time.strftime('%Y-%m-%d %H:%M:%S')))

        # 创建 well-known 配置
        wk_dir = '/www/server/panel/vhost/nginx/well-known'
        os.makedirs(wk_dir, exist_ok=True)
        public.writeFile(os.path.join(wk_dir, '%s.conf' % domain),
            'location ~ \\.well-known { allow all; }')

        # 生成宝塔兼容的 nginx 配置
        nginx_conf = '''server
{
    listen 80;
    server_name %s;
    index index.html;
    root %s;

    #error_page 404/404.html;

    #SSL-START SSL相关配置
    #SSL-END

    #ERROR-PAGE-START 错误页配置
    #ERROR-PAGE-END

    #REWRITE-START URL重写规则
    #REWRITE-END

    include /www/server/panel/vhost/nginx/well-known/%s.conf;

    # 前端路由
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 商城 H5
    location ^~ /mall/ {
        root %s;
        try_files $uri $uri/ /mall/index.html;
    }

    # PHP兼容API代理（下游系统通过 /api.php?act=xxx 调用）
    location = /api.php {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
    }


    # WebSocket
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
    }

    # 静态资源缓存
    location ~ .*\\.(js|css|gif|jpg|jpeg|png|bmp|swf|ico|svg|woff2?)$ {
        expires 30d;
        access_log /dev/null;
    }

    location ~ \\.well-known {
        allow all;
    }

    access_log /www/wwwlogs/%s.log;
    error_log /www/wwwlogs/%s.error.log;
}''' % (domain, site_root, domain, site_base, domain, domain)

        public.writeFile(conf_file, nginx_conf)

        # 保存域名到本地记录
        self._save_domain(domain)

        # 测试并重载 nginx
        test = public.ExecShell('nginx -t 2>&1')
        test_output = (str(test[0]) + str(test[1])).strip()
        if 'successful' not in test_output:
            os.remove(conf_file)
            self._save_domain('')
            return public.returnMsg(False, 'Nginx 配置测试失败: %s' % test_output[:300])
        public.serviceReload()
        return public.returnMsg(True, '域名 %s 绑定成功' % domain)

    def remove_domain(self, args):
        domain = self._get_domain()
        if not domain:
            return public.returnMsg(False, '当前未绑定域名')
        conf_file = '/www/server/panel/vhost/nginx/%s.conf' % domain
        if os.path.isfile(conf_file):
            os.remove(conf_file)
        # 删除 SSL 相关
        ssl_dir = '/www/server/panel/vhost/cert/%s' % domain
        if os.path.isdir(ssl_dir):
            public.ExecShell('rm -rf %s' % ssl_dir)
        # 删除 well-known 配置
        wk_file = '/www/server/panel/vhost/nginx/well-known/%s.conf' % domain
        if os.path.isfile(wk_file):
            os.remove(wk_file)
        # 从宝塔面板数据库删除
        pid = public.M('sites').where('name=?', (domain,)).getField('id')
        if pid:
            public.M('domain').where('pid=?', (pid,)).delete()
            public.M('sites').where('id=?', (pid,)).delete()
        self._save_domain('')
        public.serviceReload()
        return public.returnMsg(True, '已解绑域名 %s' % domain)

    # ==================== 远程更新 ====================

    def check_update(self, args):
        current = self._get_version()
        try:
            import urllib.request
            url = self.__update_server + '/update/version.json'
            req = urllib.request.Request(url, headers={'User-Agent': 'QingkaPlugin/1.0'})
            resp = urllib.request.urlopen(req, timeout=10)
            remote = json.loads(resp.read().decode())
        except Exception as e:
            return public.returnMsg(True, json.dumps({
                'has_update': False,
                'current_version': current,
                'error': str(e)
            }))

        has_update = remote.get('version', '') != current
        data = {
            'has_update': has_update,
            'current_version': current,
            'latest_version': remote.get('version', ''),
            'changelog': remote.get('changelog', ''),
            'size': remote.get('size', ''),
            'date': remote.get('date', '')
        }
        return public.returnMsg(True, json.dumps(data))

    def do_update(self, args):
        update_type = getattr(args, 'type', 'full')
        if update_type not in ('full', 'backend', 'frontend', 'mall'):
            return public.returnMsg(False, '无效的更新类型: %s' % update_type)
        try:
            import urllib.request
            base = self.__update_server + '/update'

            if update_type in ('backend', 'full'):
                self._download_and_extract(base + '/backend.tar.gz', self.__go_dir)
            if update_type in ('frontend', 'full'):
                self._download_and_extract(base + '/frontend.tar.gz', self._get_site_root())
            if update_type in ('mall', 'full'):
                self._download_and_extract(base + '/mall.tar.gz', self._get_mall_root())
            # 后端更新时执行新的数据库迁移
            if update_type in ('backend', 'full'):
                self._run_migrations()
            # 更新版本号
            try:
                url = self.__update_server + '/update/version.json'
                req = urllib.request.Request(url, headers={'User-Agent': 'QingkaPlugin/1.0'})
                resp = urllib.request.urlopen(req, timeout=10)
                remote = json.loads(resp.read().decode())
                self._save_version(remote.get('version', ''))
            except Exception as e:
                public.WriteLog('qingka_manager', '获取远程版本号失败: %s' % str(e))
            if update_type in ('backend', 'full'):
                self._ensure_config()
                self.restart(args)
            return public.returnMsg(True, '更新成功（%s）' % update_type)
        except Exception as e:
            return public.returnMsg(False, '更新失败: %s' % str(e))

    def rollback(self, args):
        bin_path = os.path.join(self.__go_dir, self.__bin_name)
        bak_path = bin_path + '.bak'
        if not os.path.isfile(bak_path):
            return public.returnMsg(False, '没有可回滚的备份')
        self.stop(args)
        time.sleep(1)
        if os.path.isfile(bin_path):
            os.remove(bin_path)
        os.rename(bak_path, bin_path)
        os.chmod(bin_path, 0o755)
        self.start(args)
        return public.returnMsg(True, '已回滚并重启')


    def get_rollback_info(self, args):
        bin_path = os.path.join(self.__go_dir, self.__bin_name)
        bak_path = bin_path + '.bak'
        has_backup = os.path.isfile(bak_path)
        data = {'has_backup': has_backup, 'backup_size': '', 'backup_time': ''}
        if has_backup:
            try:
                size = os.path.getsize(bak_path)
                data['backup_size'] = str(round(size / 1024 / 1024, 1)) + ' MB'
                data['backup_time'] = time.strftime('%Y-%m-%d %H:%M', time.localtime(os.path.getmtime(bak_path)))
            except Exception as e:
                public.WriteLog('qingka_manager', '获取备份信息失败: %s' % str(e))
        return public.returnMsg(True, json.dumps(data))

    def run_db_update(self, args):
        try:
            self._run_migrations()
            return public.returnMsg(True, '数据库迁移执行完成')
        except Exception as e:
            return public.returnMsg(False, '数据库迁移失败: %s' % str(e))

    def repair_db(self, args):
        """兼容旧调用，转发到 verify_db"""
        return self.verify_db(args)

    def verify_db(self, args):
        """模式1: 检验标准数据库并补全 —— 不删除任何已有数据，只补缺失的部分"""
        db_info = self._read_db_config()
        if not db_info:
            return public.returnMsg(False, '无法读取数据库配置，请检查 config.yaml')
        db_user, db_pass, db_name = db_info['user'], db_info['pass'], db_info['name']
        report = []  # 检验报告
        warnings = []
        # ---- 1. 检验表结构 ----
        missing_before = self._verify_tables(db_user, db_pass, db_name)
        deploy_dir = os.path.join(self.__go_dir, 'deploy')
        init_sql = os.path.join(deploy_dir, 'init_db.sql')
        if os.path.isfile(init_sql):
            import tempfile
            fd_tmp, filtered_sql = tempfile.mkstemp(suffix='.sql')
            os.close(fd_tmp)
            public.ExecShell('tr -d "\r" < %s | sed -e "/^CREATE DATABASE/d" -e "/^USE /d" -e "s/^DROP TABLE/-- DROP TABLE/" -e "s/CREATE TABLE /CREATE TABLE IF NOT EXISTS /" > %s' % (init_sql, filtered_sql))
            result = self._safe_mysql_cmd(db_user, db_pass, db_name, sql_file=filtered_sql)
            if os.path.isfile(filtered_sql): os.remove(filtered_sql)
            if result[1] and result[1].strip():
                warnings.append('建表: %s' % result[1].strip()[:200])
        else:
            warnings.append('init_db.sql 不存在，跳过建表')
        missing_after = self._verify_tables(db_user, db_pass, db_name)
        if missing_before:
            fixed = [t for t in missing_before if t not in missing_after]
            if fixed:
                report.append('✅ 已补全 %d 张缺失表: %s' % (len(fixed), ', '.join(fixed)))
            if missing_after:
                report.append('⚠️ 仍有 %d 张表缺失: %s' % (len(missing_after), ', '.join(missing_after)))
        else:
            report.append('✅ 表结构完整 (%d 张表全部存在)' % len(self._REQUIRED_TABLES))
        # ---- 2. 增量迁移 ----
        mig_dir = os.path.join(self.__go_dir, 'migrations')
        mig_count = 0
        if os.path.isdir(mig_dir):
            sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql') and f[:1].isdigit()])
            mig_count = len(sqls)
            for sql in sqls:
                result = self._safe_mysql_cmd(db_user, db_pass, db_name, sql_file=os.path.join(mig_dir, sql))
                if result[1] and result[1].strip():
                    warnings.append('%s: %s' % (sql, result[1].strip()[:100]))
        report.append('✅ 增量迁移已执行 (%d 个文件)' % mig_count)
        # ---- 3. 检验管理员账号 ----
        check_admin = self._exec_sql(db_user, db_pass, db_name, "SELECT COUNT(*) FROM qingka_wangke_user WHERE grade='3';")
        admin_exists = check_admin[0] and check_admin[0].strip() not in ('0', '')
        if admin_exists:
            report.append('✅ 管理员账号已存在')
        else:
            seed_sql = (
                "INSERT INTO qingka_wangke_user (uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth) "
                "VALUES (1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');"
            )
            result = self._exec_sql(db_user, db_pass, db_name, seed_sql)
            if result[1] and result[1].strip():
                warnings.append('管理员: %s' % result[1].strip()[:200])
                report.append('⚠️ 管理员账号创建失败')
            else:
                report.append('✅ 已创建默认管理员 admin/admin123')
        # ---- 4. 检验基础配置 ----
        check_cfg = self._exec_sql(db_user, db_pass, db_name, "SELECT COUNT(*) FROM qingka_wangke_config;")
        cfg_count = int(check_cfg[0].strip()) if check_cfg[0] and check_cfg[0].strip().isdigit() else 0
        config_sql = (
            "INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES "
            "('sitename',''),('sykg','1'),('version','1.0.0'),('user_yqzc','0'),"
            "('sjqykg','0'),('user_htkh','0'),('dl_pkkg','0'),('zdpay','0'),"
            "('flkg','1'),('fllx','0'),('djfl','0'),('notice',''),"
            "('bz',''),('logo',''),('hlogo',''),('tcgonggao',''),('pass2_kg','1');"
        )
        self._exec_sql(db_user, db_pass, db_name, config_sql)
        check_cfg2 = self._exec_sql(db_user, db_pass, db_name, "SELECT COUNT(*) FROM qingka_wangke_config;")
        cfg_count2 = int(check_cfg2[0].strip()) if check_cfg2[0] and check_cfg2[0].strip().isdigit() else 0
        if cfg_count2 > cfg_count:
            report.append('✅ 已补全 %d 条基础配置' % (cfg_count2 - cfg_count))
        else:
            report.append('✅ 基础配置完整 (%d 条)' % cfg_count2)
        # ---- 5. 检验 pass2 列 ----
        check_pass2 = self._exec_sql(db_user, db_pass, db_name,
            "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME='qingka_wangke_user' AND COLUMN_NAME='pass2';" % db_name)
        has_pass2 = check_pass2[0] and check_pass2[0].strip() not in ('0', '')
        if has_pass2:
            report.append('✅ pass2 列已存在')
        else:
            self._exec_sql(db_user, db_pass, db_name,
                "ALTER TABLE `qingka_wangke_user` ADD COLUMN `pass2` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '管理员二级密码' AFTER `pass`;")
            report.append('✅ 已自动添加 pass2 列')
        # ---- 汇总 ----
        summary = '\n'.join(report)
        if warnings:
            public.WriteLog('qingka_manager', '数据库检验警告: %s' % '; '.join(warnings))
            summary += '\n\n⚠️ 警告信息:\n' + '\n'.join(warnings)
        return public.returnMsg(True, summary)

    def reset_db(self, args):
        """模式2: 重置一套全新的标准数据库 —— 删除所有表后重新创建"""
        confirm = getattr(args, 'confirm', '')
        if confirm != 'YES':
            return public.returnMsg(False, '危险操作！请传入 confirm=YES 确认重置。此操作将删除所有数据！')
        db_info = self._read_db_config()
        if not db_info:
            return public.returnMsg(False, '无法读取数据库配置，请检查 config.yaml')
        db_user, db_pass, db_name = db_info['user'], db_info['pass'], db_info['name']
        logs = []
        # 1. 获取所有现有表并逐一 DROP
        result = self._safe_mysql_cmd(db_user, db_pass, db_name,
            sql_cmd="SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA='%s';" % db_name)
        if result[0] and result[0].strip():
            tables = [t.strip() for t in result[0].strip().split('\n') if t.strip() and t.strip() != 'TABLE_NAME']
            if tables:
                # 先关闭外键检查，再逐一 DROP
                drop_sql = 'SET FOREIGN_KEY_CHECKS=0;\n'
                for t in tables:
                    drop_sql += 'DROP TABLE IF EXISTS `%s`;\n' % t
                drop_sql += 'SET FOREIGN_KEY_CHECKS=1;\n'
                self._exec_sql(db_user, db_pass, db_name, drop_sql)
                logs.append('已删除 %d 张旧表' % len(tables))
        # 2. 重新建表（直接执行 init_db.sql，不过滤 DROP 语句）
        deploy_dir = os.path.join(self.__go_dir, 'deploy')
        init_sql = os.path.join(deploy_dir, 'init_db.sql')
        if os.path.isfile(init_sql):
            import tempfile
            fd_tmp, clean_sql = tempfile.mkstemp(suffix='.sql')
            os.close(fd_tmp)
            public.ExecShell('tr -d "\r" < %s | sed -e "/^CREATE DATABASE/d" -e "/^USE /d" > %s' % (init_sql, clean_sql))
            result = self._safe_mysql_cmd(db_user, db_pass, db_name, sql_file=clean_sql)
            if os.path.isfile(clean_sql): os.remove(clean_sql)
            if result[1] and result[1].strip():
                logs.append('建表警告: %s' % result[1].strip()[:200])
            logs.append('已重新创建标准表结构')
        else:
            return public.returnMsg(False, 'init_db.sql 不存在，无法重置数据库')
        # 3. 执行增量迁移
        mig_dir = os.path.join(self.__go_dir, 'migrations')
        if os.path.isdir(mig_dir):
            sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql') and f[:1].isdigit()])
            for sql in sqls:
                self._safe_mysql_cmd(db_user, db_pass, db_name, sql_file=os.path.join(mig_dir, sql))
            logs.append('已执行 %d 个增量迁移' % len(sqls))
        # 4. 插入默认管理员
        seed_sql = (
            "INSERT INTO qingka_wangke_user (uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth) "
            "VALUES (1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');"
        )
        result = self._exec_sql(db_user, db_pass, db_name, seed_sql)
        if result[1] and result[1].strip():
            logs.append('管理员创建警告: %s' % result[1].strip()[:200])
        else:
            logs.append('已创建默认管理员 admin/admin123')
        # 5. 插入默认配置
        config_sql = (
            "INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES "
            "('sitename',''),('sykg','1'),('version','1.0.0'),('user_yqzc','0'),"
            "('sjqykg','0'),('user_htkh','0'),('dl_pkkg','0'),('zdpay','0'),"
            "('flkg','1'),('fllx','0'),('djfl','0'),('notice',''),"
            "('bz',''),('logo',''),('hlogo',''),('tcgonggao',''),('pass2_kg','1');"
        )
        self._exec_sql(db_user, db_pass, db_name, config_sql)
        logs.append('已写入默认系统配置')
        # 6. 验证
        missing = self._verify_tables(db_user, db_pass, db_name)
        if missing:
            logs.append('⚠️ 重置后仍缺少 %d 张表: %s' % (len(missing), ', '.join(missing)))
        else:
            logs.append('✅ 所有 %d 张标准表已就绪' % len(self._REQUIRED_TABLES))
        public.WriteLog('qingka_manager', '数据库已重置: %s' % '; '.join(logs))
        return public.returnMsg(True, '数据库重置完成:\n' + '\n'.join(logs))

    # ==================== 一键安装 ====================

    def init_install(self, args):
        domain = getattr(args, 'domain', '').strip()
        db_user = self._safe_db_param(getattr(args, 'db_user', ''))
        db_pass = self._safe_db_param(getattr(args, 'db_pass', ''))
        db_name = self._safe_db_param(getattr(args, 'db_name', ''))
        redis_pass = self._safe_db_param(getattr(args, 'redis_pass', ''))

        if not domain or not db_user or not db_pass or not db_name:
            return public.returnMsg(False, '请填写完整信息')
        # 验证数据库连接
        test = self._safe_mysql_cmd(db_user, db_pass, '')
        test_output = (str(test[0]) + str(test[1])).strip()
        if 'Access denied' in test_output or 'ERROR' in test_output:
            return public.returnMsg(False, '数据库连接失败，请检查用户名和密码。错误: %s' % test_output[:200])

        try:
            site_root = self._get_site_root()
            mall_root = self._get_mall_root()
            for d in [self.__go_dir + '/config', site_root, mall_root]:
                os.makedirs(d, exist_ok=True)

            # 2. 从更新源下载文件
            import urllib.request
            base = self.__update_server + '/update'
            self._download_and_extract(base + '/backend.tar.gz', self.__go_dir)
            self._download_and_extract(base + '/frontend.tar.gz', site_root)
            self._download_and_extract(base + '/mall.tar.gz', mall_root)
            # 验证前端文件是否存在
            if not os.path.isfile(os.path.join(site_root, 'index.html')):
                return public.returnMsg(False, '前端文件下载失败，请检查网络连接')

            # 3. 生成配置文件
            jwt_secret = hashlib.sha256(os.urandom(32)).hexdigest()[:32]
            bridge_secret = hashlib.md5(os.urandom(16)).hexdigest()
            variables = {
                'bridge_secret': bridge_secret,
                'db_user': db_user, 'db_pass': db_pass, 'db_name': db_name,
                'redis_pass': redis_pass, 'jwt_secret': jwt_secret,
            }
            config = self._render_template('config.yaml.tpl', variables)
            if not config:
                return public.returnMsg(False, '配置模板文件不存在，请检查插件完整性')
            public.writeFile(self.__config_file, config)

            # 4. 创建数据库
            result = self._safe_mysql_cmd(db_user, db_pass, '', sql_cmd='CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;' % db_name)
            if result[1] and result[1].strip():
                public.WriteLog('qingka_manager', '创建数据库失败: %s' % result[1].strip()[:200])
            # 5. 建表（init_db.sql）—— 安全模式：注释 DROP TABLE、加 IF NOT EXISTS
            deploy_dir = os.path.join(self.__go_dir, 'deploy')
            init_sql = os.path.join(deploy_dir, 'init_db.sql')
            if os.path.isfile(init_sql):
                import tempfile
                fd_tmp, filtered_sql = tempfile.mkstemp(suffix='.sql')
                os.close(fd_tmp)
                public.ExecShell('tr -d "\r" < %s | sed -e "/^CREATE DATABASE/d" -e "/^USE /d" -e "s/^DROP TABLE/-- DROP TABLE/" -e "s/CREATE TABLE /CREATE TABLE IF NOT EXISTS /" > %s' % (init_sql, filtered_sql))
                result = self._safe_mysql_cmd(db_user, db_pass, db_name, sql_file=filtered_sql)
                if os.path.isfile(filtered_sql): os.remove(filtered_sql)
                if result[1] and result[1].strip():
                    public.WriteLog('qingka_manager', '建表警告: %s' % result[1].strip()[:300])

            # 6. 执行增量迁移（仅运行编号格式的迁移文件）
            self._run_migrations()
            # 7. 确保管理员账号和基础配置存在（仅在无管理员时插入，不硬编码uid）
            check_admin = self._exec_sql(db_user, db_pass, db_name, "SELECT COUNT(*) FROM qingka_wangke_user WHERE grade='3';")
            admin_exists = check_admin[0] and check_admin[0].strip() not in ('0', '')
            if not admin_exists:
                seed_sql = (
                    "INSERT INTO qingka_wangke_user (uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth) "
                    "VALUES (1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');"
                )
                result = self._exec_sql(db_user, db_pass, db_name, seed_sql)
                if result[1] and result[1].strip():
                    public.WriteLog('qingka_manager', '插入管理员账号失败: %s' % result[1].strip()[:200])

            # 8. 绑定域名
            self.setup_domain(type('Args', (), {'domain': domain})())

            # 9. 设置权限并启动
            bin_path = os.path.join(self.__go_dir, self.__bin_name)
            if os.path.isfile(bin_path):
                os.chmod(bin_path, 0o755)

            # 9.5 验证关键数据表是否存在
            missing = self._verify_tables(db_user, db_pass, db_name)
            if missing:
                public.WriteLog('qingka_manager', '安装后缺少数据表: %s，尝试自动修复...' % ', '.join(missing))
                # 自动执行 repair_db 补全
                self.repair_db(type('Args', (), {})())
                # 再检查一次
                still_missing = self._verify_tables(db_user, db_pass, db_name)
                if still_missing:
                    public.WriteLog('qingka_manager', '自动修复后仍缺少: %s' % ', '.join(still_missing))

            # 10. 保存版本
            try:
                url = self.__update_server + '/update/version.json'
                req = urllib.request.Request(url, headers={'User-Agent': 'QingkaPlugin/1.0'})
                resp = urllib.request.urlopen(req, timeout=10)
                remote = json.loads(resp.read().decode())
                self._save_version(remote.get('version', '1.0.0'))
            except Exception as e:
                public.WriteLog('qingka_manager', '获取初始版本号失败: %s' % str(e))
                self._save_version('1.0.0')
            # 11. 启动 Go 服务
            self._start_no_auth()

            return public.returnMsg(True, '安装成功！域名: %s' % domain)
        except Exception as e:
            return public.returnMsg(False, '安装失败: %s' % str(e))

    # ==================== 演示模式 ====================

    __demo_flag = '/www/wwwroot/qingka/go-api/.demo_mode'

    def get_demo_mode(self, args):
        """获取演示模式状态"""
        enabled = os.path.isfile(self.__demo_flag)
        return public.returnMsg(True, json.dumps({'enabled': enabled}))

    def set_demo_mode(self, args):
        """设置演示模式开关"""
        enabled = getattr(args, 'enabled', 'false')
        if isinstance(enabled, str):
            enabled = enabled.lower() in ('true', '1', 'yes')
        if enabled:
            os.makedirs(os.path.dirname(self.__demo_flag), exist_ok=True)
            public.writeFile(self.__demo_flag, 'demo')
            public.WriteLog('qingka_manager', '演示模式已开启')
            return public.returnMsg(True, '演示模式已开启，所有写操作将被拦截')
        else:
            if os.path.isfile(self.__demo_flag):
                os.remove(self.__demo_flag)
            public.WriteLog('qingka_manager', '演示模式已关闭')
            return public.returnMsg(True, '演示模式已关闭')

    # ==================== 卸载 ====================

    def full_uninstall(self, args):
        remove_data = getattr(args, 'remove_data', False)
        if isinstance(remove_data, str):
            remove_data = remove_data in ('true', '1', 'True')

        try:
            # 读取配置获取数据库信息（卸载前读取）
            db_info = self._read_db_config()

            # 停止服务
            self.stop(args)
            domain = self._get_domain()
            if domain:
                self.remove_domain(args)
            # 清理统一部署目录（remove_domain 已删除宝塔DB和nginx配置）
            dirs_to_clean = [self.__go_dir, self._get_site_root(), self._get_mall_root()]
            for d in dirs_to_clean:
                if os.path.isdir(d):
                    public.ExecShell('rm -rf %s' % d)

            # 移除 systemd 服务
            self.remove_systemd(args)
            # 删除数据库
            if remove_data and db_info:
                self._safe_mysql_cmd(db_info['user'], db_info['pass'], '', sql_cmd='DROP DATABASE IF EXISTS `%s`;' % db_info['name'])
            # 清理站点目录（保留 domain.txt 和 version.json 除非彻底删除）
            if remove_data and os.path.isdir(self.__site_dir):
                public.ExecShell('rm -rf %s' % self.__site_dir)
            elif os.path.isdir(self.__site_dir):
                for f in os.listdir(self.__site_dir):
                    fp = os.path.join(self.__site_dir, f)
                    if f in ('domain.txt', 'version.json'): continue
                    if os.path.isdir(fp):
                        public.ExecShell('rm -rf %s' % fp)
                    else:
                        os.remove(fp)

            return public.returnMsg(True, '卸载完成')
        except Exception as e:
            return public.returnMsg(False, '卸载失败: %s' % str(e))

    # ==================== 内部方法 ====================

    def _exec_sql(self, db_user, db_pass, db_name, sql):
        return self._safe_mysql_cmd(db_user, db_pass, db_name, sql_cmd=sql)
    def _get_port(self):
        try:
            content = public.readFile(self.__config_file)
            m = re.search(r'port:\s*(\d+)', content)
            if m: return m.group(1)
        except Exception:
            pass
        return '8080'

    def _get_pid(self):
        if os.path.isfile(self.__pid_file):
            try:
                pid = public.readFile(self.__pid_file).strip()
                if pid and pid.isdigit():
                    return pid
            except Exception as e:
                public.WriteLog('qingka_manager', '读取PID文件失败: %s' % str(e))
        try:
            pid = public.ExecShell("pgrep -f '%s/%s'" % (self.__go_dir, self.__bin_name))[0].strip()
            if pid and '\n' in pid:
                pid = pid.split('\n')[0]
            if pid and pid.isdigit():
                self._save_pid(pid)
                return pid
        except Exception as e:
            public.WriteLog('qingka_manager', '进程检测失败: %s' % str(e))
        return None

    def _save_pid(self, pid):
        public.writeFile(self.__pid_file, str(pid))

    def _del_pid(self):
        try:
            if os.path.isfile(self.__pid_file):
                os.remove(self.__pid_file)
        except Exception as e:
            public.WriteLog('qingka_manager', '删除PID文件失败: %s' % str(e))

    def _get_version(self):
        if os.path.isfile(self.__version_file):
            try:
                d = json.loads(public.readFile(self.__version_file))
                return d.get('version', '未知')
            except Exception as e:
                public.WriteLog('qingka_manager', '读取版本文件失败: %s' % str(e))
        return '未知'

    def _save_version(self, ver):
        os.makedirs(os.path.dirname(self.__version_file), exist_ok=True)
        public.writeFile(self.__version_file, json.dumps({'version': ver}))

    def _get_domain(self):
        domain_file = os.path.join(self.__site_dir, 'domain.txt')
        if os.path.isfile(domain_file):
            return public.readFile(domain_file).strip()
        return ''

    def _save_domain(self, domain):
        os.makedirs(self.__site_dir, exist_ok=True)
        public.writeFile(os.path.join(self.__site_dir, 'domain.txt'), domain)

    def _download_and_extract(self, url, target_dir):
        import urllib.request, tempfile
        os.makedirs(target_dir, exist_ok=True)
        fd, tmp = tempfile.mkstemp(suffix='.tar.gz')
        os.close(fd)
        try:
            req = urllib.request.Request(url, headers={'User-Agent': 'QingkaPlugin/1.0'})
            resp = urllib.request.urlopen(req, timeout=120)
            with open(tmp, 'wb') as f:
                while True:
                    chunk = resp.read(8192)
                    if not chunk: break
                    f.write(chunk)
            # 前端/商城更新时保护网站图标
            favicon_backup = None
            if '/frontend' in url or '/mall' in url:
                favicon_path = os.path.join(target_dir, 'favicon.ico')
                if os.path.isfile(favicon_path):
                    favicon_backup = public.readFile(favicon_path)
            # 后端特殊处理：备份旧二进制 + 保护配置文件
            config_backup = None
            if '/backend' in url:
                bin_path = os.path.join(self.__go_dir, self.__bin_name)
                if os.path.isfile(bin_path):
                    bak = bin_path + '.bak'
                    if os.path.isfile(bak): os.remove(bak)
                    os.rename(bin_path, bak)
                # 保护用户配置文件
                if os.path.isfile(self.__config_file):
                    config_backup = public.readFile(self.__config_file)
            public.ExecShell('tar -xzf %s -C %s' % (tmp, target_dir))
            public.ExecShell('chown -R www:www %s' % target_dir)
            if config_backup is not None:
                public.writeFile(self.__config_file, config_backup)
            if favicon_backup is not None:
                public.writeFile(os.path.join(target_dir, 'favicon.ico'), favicon_backup)
        finally:
            if os.path.isfile(tmp): os.remove(tmp)

    def _ensure_config(self):
        """首次安装时自动生成默认config.yaml，已有则不动"""
        if os.path.isfile(self.__config_file):
            return
        os.makedirs(os.path.dirname(self.__config_file), exist_ok=True)
        jwt_secret = hashlib.sha256(os.urandom(32)).hexdigest()[:32]
        variables = {
            'bridge_secret': hashlib.md5(os.urandom(16)).hexdigest(),
            'db_user': 'root', 'db_pass': '', 'db_name': 'qingka',
            'redis_pass': '', 'jwt_secret': jwt_secret,
        }
        config = self._render_template('config.yaml.tpl', variables)
        if not config:
            public.WriteLog('qingka_manager', '模板文件不存在，跳过配置生成')
            return
        public.writeFile(self.__config_file, config)
        public.WriteLog('qingka_manager', '已自动生成默认配置文件 config.yaml')


    def _render_template(self, tpl_name, variables):
        """从 templates 目录读取模板并替换变量"""
        tpl_path = os.path.join(self.__plugin_path, 'templates', tpl_name)
        if not os.path.isfile(tpl_path):
            return None
        content = public.readFile(tpl_path)
        for key, val in variables.items():
            content = content.replace('{{%s}}' % key, str(val))
        return content

    def _safe_mysql_cmd(self, db_user, db_pass, db_name, sql_cmd='', sql_file=''):
        """安全执行 MySQL 命令，密码写入临时文件而非命令行参数"""
        import tempfile
        fd, cnf = tempfile.mkstemp(suffix='.cnf')
        try:
            tmp_sql = None
            os.write(fd, ('[client]\nuser=%s\npassword=%s\n' % (db_user, db_pass)).encode())
            os.close(fd)
            os.chmod(cnf, 0o600)
            if sql_file:
                cmd = 'mysql --defaults-extra-file=%s --force %s < %s 2>&1' % (cnf, db_name, sql_file)
            elif sql_cmd:
                # 写入临时 SQL 文件
                fd2, tmp_sql = tempfile.mkstemp(suffix='.sql')
                os.write(fd2, sql_cmd.encode('utf-8'))
                os.close(fd2)
                cmd = 'mysql --defaults-extra-file=%s --force %s < %s 2>&1' % (cnf, db_name, tmp_sql)
            else:
                cmd = 'mysql --defaults-extra-file=%s %s -e "SELECT 1" 2>&1' % (cnf, db_name)
                tmp_sql = None
            result = public.ExecShell(cmd)
            if sql_cmd and tmp_sql and os.path.isfile(tmp_sql):
                os.remove(tmp_sql)
            return result
        finally:
            if os.path.isfile(cnf):
                os.remove(cnf)

    def _read_db_config(self):
        try:
            import yaml
            conf = yaml.safe_load(public.readFile(self.__config_file))
            db = conf.get('database', {})
            return {'user': db.get('user', ''), 'pass': db.get('password', ''), 'name': db.get('dbname', '')}
        except Exception as e:
            public.WriteLog('qingka_manager', '读取数据库配置失败(yaml): %s' % str(e))
            try:
                content = public.readFile(self.__config_file)
                user = re.search(r'user:\s*(.+)', content)
                pwd = re.search(r'password:\s*["\']?([^"\'\n]+)', content)
                name = re.search(r'dbname:\s*["\']?([^"\'\n]+)', content)
                if user and pwd and name:
                    return {'user': user.group(1).strip(), 'pass': pwd.group(1).strip(), 'name': name.group(1).strip()}
            except Exception as e2:
                public.WriteLog('qingka_manager', '读取数据库配置失败(正则): %s' % str(e2))
        return None

    # 关键数据表列表（init_db.sql 中所有表）
    _REQUIRED_TABLES = [
        # 核心表
        'qingka_wangke_user', 'qingka_wangke_config', 'qingka_wangke_order',
        'qingka_wangke_class', 'qingka_wangke_fenlei', 'qingka_wangke_dengji',
        'qingka_wangke_huoyuan', 'qingka_wangke_log', 'qingka_wangke_moneylog',
        'qingka_wangke_km', 'qingka_wangke_gonggao', 'qingka_wangke_huodong',
        'qingka_wangke_mijia', 'qingka_wangke_pay', 'qingka_wangke_user_favorite',
        'qingka_wangke_checkin', 'qingka_wangke_ticket', 'qingka_wangke_push_logs',
        'qingka_wangke_huodong_record', 'qingka_wangke_zhiya_config', 'qingka_wangke_zhiya_records',
        'qingka_wangke_sync_config', 'qingka_wangke_sync_log',
        # 聊天/邮件
        'qingka_chat_list', 'qingka_chat_msg', 'qingka_chat_msg_archive', 'qingka_mail',
        'qingka_smtp_config', 'qingka_email_pool', 'qingka_email_template',
        'qingka_email_log', 'qingka_email_send_log',
        # 平台/模块/菜单
        'qingka_platform_config', 'qingka_dynamic_module', 'menu_config', 'qingka_ext_menu',
        # 商城/租户
        'qingka_mall_pay_order', 'qingka_tenant', 'qingka_tenant_product', 'qingka_c_user',
        # 辅助模块
        'mlsx_gslb', 'mlsx_wj_wq', 'qingka_wangke_flash_sdxy',
        # 打卡/运动
        'qingka_wangke_appui', 'qingka_wangke_yfdk', 'qingka_wangke_yfdk_projects', 'qingka_wangke_sxdk',
        'qingka_wangke_hzw_ydsj', 'xm_project', 'xm_order',
        'w_app', 'w_order',
        'yy_ydsj_dd', 'yy_ydsj_student',
        # 凸知打卡
        'qingka_wangke_dakaaz', 'qingka_wangke_daka_query_record',
        # 图图强国/土拨鼠
        'tutuqg', 'qingka_wangke_dialogue',
        'points_product', 'points_product_code', 'points_exchange_record',
        # 智文论文
        'qingka_wangke_lunwen',
    ]

    def _verify_tables(self, db_user, db_pass, db_name):
        """检查关键数据表是否存在，返回缺失表名列表"""
        try:
            result = self._safe_mysql_cmd(db_user, db_pass, db_name,
                sql_cmd="SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA='%s';" % db_name)
            existing = set(result[0].strip().split('\n')) if result[0] else set()
            missing = [t for t in self._REQUIRED_TABLES if t not in existing]
            return missing
        except Exception as e:
            public.WriteLog('qingka_manager', '验证数据表失败: %s' % str(e))
            return []

    def _run_migrations(self):
        db_info = self._read_db_config()
        if not db_info:
            return
        mig_dir = os.path.join(self.__go_dir, 'migrations')
        if not os.path.isdir(mig_dir):
            return
        sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql') and f[:1].isdigit()])
        for sql in sqls:
            sql_path = os.path.join(mig_dir, sql)
            try:
                result = self._safe_mysql_cmd(db_info['user'], db_info['pass'], db_info['name'], sql_file=sql_path)
                if result[1] and result[1].strip():
                    public.WriteLog('qingka_manager', '迁移 %s 警告: %s' % (sql, result[1].strip()[:200]))
            except Exception as e:
                public.WriteLog('qingka_manager', '迁移 %s 失败: %s' % (sql, str(e)))

    def update_plugin(self, args):
        try:
            import urllib.request, tempfile
            url = self.__update_server + '/update/plugin.tar.gz'
            fd, tmp = tempfile.mkstemp(suffix='.tar.gz')
            os.close(fd)
            req = urllib.request.Request(url, headers={'User-Agent': 'QingkaPlugin/1.0'})
            resp = urllib.request.urlopen(req, timeout=30)
            with open(tmp, 'wb') as f:
                while True:
                    chunk = resp.read(8192)
                    if not chunk: break
                    f.write(chunk)
            public.ExecShell('tar -xzf %s -C %s' % (tmp, self.__plugin_path))
            os.remove(tmp)
            return public.returnMsg(True, '插件已更新，请刷新页面')
        except Exception as e:
            return public.returnMsg(False, '插件更新失败: %s' % str(e))

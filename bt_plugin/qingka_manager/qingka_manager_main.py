# coding: utf-8
import sys, os, json, time, re, hashlib
panelPath = '/www/server/panel'
if panelPath not in sys.path:
    sys.path.insert(0, panelPath)
import public

class qingka_manager_main:
    __plugin_path = '/www/server/panel/plugin/qingka_manager'
    __site_dir = '/www/wwwroot/qingka'
    __go_dir = '/www/wwwroot/qingka/go-api'
    __bin_name = 'server'
    __log_file = '/www/wwwroot/qingka/go-api/go-api.log'
    __pid_file = '/www/wwwroot/qingka/go-api/go-api.pid'
    __config_file = '/www/wwwroot/qingka/go-api/config/config.yaml'
    __version_file = '/www/wwwroot/qingka/version.json'
    __service_name = 'qingka-api'
    __update_server = 'https://29.colnt.com'
    __license_file = '/www/wwwroot/qingka/license.key'


    def __init__(self):
        pass

    def _get_site_root(self):
        domain = self._get_domain()
        if domain:
            return '/www/wwwroot/' + domain
        return '/www/wwwroot/qingka/admin'

    def _get_mall_root(self):
        return self._get_site_root() + '/mall'

    # ==================== 输入校验 ====================

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

    # ==================== 授权验证 ====================

    def get_license_status(self, args):
        key = self._get_license_key()
        if not key:
            return public.returnMsg(True, json.dumps({'licensed': False, 'msg': '未输入授权码'}))
        ok, msg = self._verify_license(key)
        return public.returnMsg(True, json.dumps({'licensed': ok, 'key': key[:8] + '****', 'msg': msg}))

    def save_license(self, args):
        key = getattr(args, 'license_key', '').strip()
        if not key:
            return public.returnMsg(False, '请输入授权码')
        ok, msg = self._verify_license(key)
        if not ok:
            return public.returnMsg(False, msg)
        os.makedirs(os.path.dirname(self.__license_file), exist_ok=True)
        public.writeFile(self.__license_file, key)
        return public.returnMsg(True, '授权验证成功')

    def _get_license_key(self):
        if os.path.isfile(self.__license_file):
            return public.readFile(self.__license_file).strip()
        return ''

    def _cache_sign(self, data_str):
        """对缓存数据生成 HMAC-SHA256 签名，防篡改"""
        import hmac
        secret = self.__update_server + self.__license_file
        return hmac.new(secret.encode(), data_str.encode(), hashlib.sha256).hexdigest()

    def _verify_license(self, key):
        try:
            import urllib.request, hmac as _hmac
            domain = self._get_domain()
            mid = ''
            if os.path.isfile('/etc/machine-id'):
                mid = open('/etc/machine-id').read().strip()
            ts = int(time.time())
            sign_str = 'domain=%s&license_key=%s&machine_id=%s&timestamp=%d&version=' % (domain or '', key, mid, ts)
            sign = _hmac.new(self._get_client_secret().encode(), sign_str.encode(), hashlib.sha256).hexdigest()
            url = self.__update_server + '/api/v1/license/verify'
            body = json.dumps({
                'license_key': key, 'domain': domain or '', 'machine_id': mid,
                'version': '', 'timestamp': ts, 'sign': sign
            }).encode()
            req = urllib.request.Request(url, data=body, headers={'Content-Type': 'application/json', 'User-Agent': 'QingkaPlugin/1.0'})
            resp = urllib.request.urlopen(req, timeout=10)
            result = json.loads(resp.read().decode())
            if result.get('code') == 0 and result.get('data', {}).get('valid'):
                # 缓存授权结果（带 HMAC 签名防篡改）
                cache_path = os.path.join(self.__site_dir, '.license_cache')
                os.makedirs(self.__site_dir, exist_ok=True)
                cache_data = json.dumps({'key': key, 'ts': ts}, separators=(',', ':'))
                cache_sign = self._cache_sign(cache_data)
                public.writeFile(cache_path, json.dumps({'d': cache_data, 's': cache_sign}))
                return True, '授权有效'
            return False, result.get('message', result.get('msg', '授权码无效'))
        except Exception as e:
            cache_path = os.path.join(self.__site_dir, '.license_cache')
            if os.path.isfile(cache_path):
                try:
                    raw = json.loads(public.readFile(cache_path))
                    cache_data = raw.get('d', '')
                    cache_sign = raw.get('s', '')
                    # 验证 HMAC 签名
                    if cache_data and cache_sign == self._cache_sign(cache_data):
                        c = json.loads(cache_data)
                        if c.get('key') == key and time.time() - c.get('ts', 0) < 86400 * 7:
                            return True, '授权有效（离线缓存）'
                except Exception:
                    pass
            return False, '授权验证失败: %s' % str(e)

    def _get_client_secret(self):
        """读取客户端签名密钥（与授权站 config.toml 中 client_secret 一致）"""
        secret_file = os.path.join(self.__site_dir, '.client_secret')
        if os.path.isfile(secret_file):
            return open(secret_file).read().strip()
        return 'default-client-secret'
    # ==================== 状态 ====================

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

    def start(self, args):
        # 授权验证
        key = self._get_license_key()
        if key:
            ok, msg = self._verify_license(key)
            if not ok:
                return public.returnMsg(False, '授权验证失败: %s' % msg)
        else:
            return public.returnMsg(False, '请先在「首页概览」输入授权码')
        pid = self._get_pid()
        if pid and os.path.exists('/proc/%s' % pid):
            return public.returnMsg(False, '服务已在运行中，PID: %s' % pid)
        bin_path = os.path.join(self.__go_dir, self.__bin_name)
        if not os.path.isfile(bin_path):
            return public.returnMsg(False, '可执行文件不存在')
        # 自动杀掉占用端口的旧进程
        port = self._get_port()
        if port:
            old_pids = public.ExecShell("lsof -i :%s -t 2>/dev/null" % port)[0].strip()
            if old_pids:
                for p in old_pids.split('\n'):
                    p = p.strip()
                    if p and p.isdigit():
                        public.ExecShell('kill -9 %s' % p)
                        public.WriteLog('qingka_manager', '已杀掉占用端口 %s 的旧进程 PID: %s' % (port, p))
                time.sleep(1)
        os.chmod(bin_path, 0o755)
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
        public.ExecShell('kill %s' % pid)
        time.sleep(1)
        if os.path.exists('/proc/%s' % pid):
            public.ExecShell('kill -9 %s' % pid)
            time.sleep(0.5)
        self._del_pid()
        if os.path.exists('/proc/%s' % pid):
            return public.returnMsg(False, '停止失败，进程仍在运行，PID: %s' % pid)
        return public.returnMsg(True, '已停止')

    def restart(self, args):
        self.stop(args)
        time.sleep(1)
        return self.start(args)

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
        result = public.ExecShell('mysql -u%s -p"%s" %s -e "SELECT 1" 2>&1' % (db_info['user'], db_info['pass'], db_info['name']))
        output = (str(result[0]) + str(result[1])).strip()
        # 过滤 MySQL 密码警告
        lines = [l for l in output.split('\n') if 'Using a password on the command line' not in l]
        clean = '\n'.join(lines).strip()
        if 'Access denied' in output or 'ERROR' in output:
            err = clean.split('ERROR')[-1].strip() if 'ERROR' in clean else clean
            return public.returnMsg(False, '数据库连接失败: %s。请到「配置文件」标签页修改 database 部分的 user 和 password' % err[:200])
        return public.returnMsg(True, '数据库连接成功')
    # ==================== 域名管理 ====================

    def setup_domain(self, args):
        domain = getattr(args, 'domain', '').strip()
        if not domain:
            return public.returnMsg(False, '请输入域名')
        if not re.match(r'^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$', domain) or len(domain) > 253:
            return public.returnMsg(False, '域名格式不正确')

        site_root = '/www/wwwroot/' + domain
        conf_dir = '/www/server/panel/vhost/nginx'
        conf_file = os.path.join(conf_dir, '%s.conf' % domain)

        # 创建站点目录
        os.makedirs(site_root, exist_ok=True)
        os.makedirs(site_root + '/mall', exist_ok=True)
        public.ExecShell('chmod -R 755 ' + site_root)
        public.ExecShell('chown -R www:www ' + site_root)

        # 注册到宝塔面板数据库
        if not public.M('sites').where('name=?', (domain,)).count():
            pid = public.M('sites').add('name,path,status,ps,type_id,addtime,project_type',
                (domain, site_root, '1', '青卡管理系统', 0, time.strftime('%Y-%m-%d %H:%M:%S'), 'PHP'))
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
    location /mall/ {
        try_files $uri $uri/ /mall/index.html;
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
}''' % (domain, site_root, domain, domain, domain)

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

    def apply_ssl(self, args):
        domain = self._get_domain()
        if not domain:
            return public.returnMsg(False, '请先绑定域名')
        if not re.match(r'^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$', domain) or len(domain) > 253:
            return public.returnMsg(False, '域名格式异常，请重新绑定')
        # 方案 1: 宝塔面板内置 SSL 申请 (panelSSL)
        try:
            import panelSSL
            ssl_obj = panelSSL.panelSSL()
            sargs = type('Args', (), {'siteName': domain, 'domains': json.dumps([domain]), 'force': 'true', 'auth_type': 'http'})()
            result = ssl_obj.apply_cert_api(sargs)
            if isinstance(result, dict) and result.get('status'):
                return public.returnMsg(True, 'SSL 证书申请成功')
        except Exception as e:
            public.WriteLog('qingka_manager', '宝塔SSL申请失败: %s' % str(e))
        # 方案 2: 宝塔 acme.sh
        acme_paths = ['/root/.acme.sh/acme.sh', '/usr/local/bin/acme.sh', '/www/server/panel/pyenv/bin/acme.sh']
        acme = next((p for p in acme_paths if os.path.isfile(p)), None)
        if acme:
            result = public.ExecShell('%s --issue -d %s --webroot %s --force 2>&1' % (acme, domain, self._get_site_root()))
            output = str(result[0]) + str(result[1])
            if 'Cert success' in output or 'already been issued' in output:
                cert_dir = '/www/server/panel/vhost/cert/%s' % domain
                os.makedirs(cert_dir, exist_ok=True)
                public.ExecShell('%s --install-cert -d %s --key-file %s/privkey.pem --fullchain-file %s/fullchain.pem --reloadcmd "/etc/init.d/nginx reload" 2>&1' % (acme, domain, cert_dir, cert_dir))
                self._enable_ssl_nginx(domain, cert_dir)
                return public.returnMsg(True, 'SSL 证书申请成功')
        # 方案 3: certbot
        certbot = public.ExecShell('which certbot 2>/dev/null')[0].strip()
        if certbot:
            result = public.ExecShell('certbot --nginx -d %s --non-interactive --agree-tos --register-unsafely-without-email 2>&1' % domain)
            output = str(result[0]) + str(result[1])
            if 'Congratulations' in output or 'Successfully' in output:
                return public.returnMsg(True, 'SSL 证书申请成功')
        return public.returnMsg(False, 'SSL 申请失败。请在宝塔面板「网站」中手动为 %s 申请 SSL 证书，或先在宝塔「网站」中添加该域名站点' % domain)

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
        db_info = self._read_db_config()
        if not db_info:
            return public.returnMsg(False, '无法读取数据库配置，请检查 config.yaml')
        db_user, db_pass, db_name = db_info['user'], db_info['pass'], db_info['name']
        logs = []
        # 1. 建表（注释掉 DROP TABLE，只创建不存在的表）
        deploy_dir = os.path.join(self.__go_dir, 'deploy')
        init_sql = os.path.join(deploy_dir, 'init_db.sql')
        if os.path.isfile(init_sql):
            result = public.ExecShell('tr -d "\r" < %s | sed -e "/^CREATE DATABASE/d" -e "/^USE /d" -e "s/^DROP TABLE/-- DROP TABLE/" -e "s/CREATE TABLE /CREATE TABLE IF NOT EXISTS /" | mysql -u%s -p"%s" %s 2>&1' % (init_sql, db_user, db_pass, db_name))
            if result[1] and result[1].strip():
                logs.append('建表: %s' % result[1].strip()[:200])
        else:
            logs.append('init_db.sql 不存在，跳过建表')
        # 2. 执行增量迁移
        mig_dir = os.path.join(self.__go_dir, 'migrations')
        if os.path.isdir(mig_dir):
            sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql')])
            for sql in sqls:
                result = public.ExecShell('mysql --force -u%s -p"%s" %s < %s 2>&1' % (db_user, db_pass, db_name, os.path.join(mig_dir, sql)))
                if result[1] and result[1].strip():
                    logs.append('%s: %s' % (sql, result[1].strip()[:100]))
        # 3. 补管理员账号
        seed_sql = (
            "INSERT IGNORE INTO qingka_wangke_user (uid, uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth) "
            "VALUES (1, 1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');"
        )
        result = self._exec_sql(db_user, db_pass, db_name, seed_sql)
        if result[1] and result[1].strip():
            logs.append('管理员: %s' % result[1].strip()[:200])
        # 4. 补基础配置
        config_sql = (
            "INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES "
            "('sitename',''),('sykg','1'),('version','1.0.0'),('user_yqzc','0'),"
            "('sjqykg','0'),('user_htkh','0'),('dl_pkkg','0'),('zdpay','0'),"
            "('flkg','1'),('fllx','0'),('djfl','0'),('notice',''),"
            "('bz',''),('logo',''),('hlogo',''),('tcgonggao','');"
        )
        self._exec_sql(db_user, db_pass, db_name, config_sql)
        if logs:
            public.WriteLog('qingka_manager', '数据库修复警告: %s' % '; '.join(logs))
            return public.returnMsg(True, '数据库补全完成（有警告）:\n' + '\n'.join(logs))
        return public.returnMsg(True, '数据库补全完成，表结构/迁移/管理员/配置均已检查')

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
        test = public.ExecShell('mysql -u%s -p"%s" -e "SELECT 1" 2>&1' % (db_user, db_pass))
        if 'Access denied' in str(test[0]) + str(test[1]) or 'ERROR' in str(test[0]) + str(test[1]):
            return public.returnMsg(False, '数据库连接失败，请检查用户名和密码。错误: %s' % (str(test[1] or test[0]).strip()[:200]))

        try:
            site_root = '/www/wwwroot/' + domain
            for d in [self.__go_dir + '/config', site_root, site_root + '/mall']:
                os.makedirs(d, exist_ok=True)

            # 2. 从更新源下载文件
            import urllib.request
            base = self.__update_server + '/update'
            self._download_and_extract(base + '/backend.tar.gz', self.__go_dir)
            self._download_and_extract(base + '/frontend.tar.gz', site_root)
            self._download_and_extract(base + '/mall.tar.gz', site_root + '/mall')
            # 验证前端文件是否存在
            if not os.path.isfile(os.path.join(site_root, 'index.html')):
                return public.returnMsg(False, '前端文件下载失败，请检查网络连接')

            # 3. 生成配置文件
            jwt_secret = hashlib.sha256(os.urandom(32)).hexdigest()[:32]
            config = '''server:
  port: 8080
  mode: release
  php_backend: ""
  php_public_url: ""
  bridge_secret: "%s"

database:
  host: 127.0.0.1
  port: 3306
  user: %s
  password: "%s"
  dbname: "%s"
  max_open_conns: 50
  max_idle_conns: 25

redis:
  host: 127.0.0.1
  port: 6379
  password: "%s"
  db: 0

jwt:
  secret: "%s"
  access_ttl: 604800
  refresh_ttl: 2592000

cache:
  order_list_ttl: 30
  class_list_ttl: 300

smtp:
  host: "smtp.qq.com"
  port: 465
  user: ""
  password: ""
  from_name: "系统通知"
  encryption: "ssl"
''' % (hashlib.md5(os.urandom(16)).hexdigest(), db_user, db_pass, db_name, redis_pass, jwt_secret)
            public.writeFile(self.__config_file, config)

            # 4. 创建数据库
            result = public.ExecShell('mysql -u%s -p"%s" -e "CREATE DATABASE IF NOT EXISTS \\`%s\\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;" 2>&1' % (db_user, db_pass, db_name))
            if result[1] and result[1].strip():
                public.WriteLog('qingka_manager', '创建数据库失败: %s' % result[1].strip()[:200])
            # 5. 建表（init_db.sql）
            deploy_dir = os.path.join(self.__go_dir, 'deploy')
            init_sql = os.path.join(deploy_dir, 'init_db.sql')
            if os.path.isfile(init_sql):
                # 过滤掉硬编码的 CREATE DATABASE / USE 语句，用命令行指定的库名
                result = public.ExecShell('tr -d "\r" < %s | sed -e "/^CREATE DATABASE/d" -e "/^USE /d" | mysql -u%s -p"%s" %s 2>&1' % (init_sql, db_user, db_pass, db_name))
                if result[1] and result[1].strip():
                    public.WriteLog('qingka_manager', '建表失败: %s' % result[1].strip()[:300])

            # 6. 执行增量迁移
            mig_dir = os.path.join(self.__go_dir, 'migrations')
            if os.path.isdir(mig_dir):
                sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql')])
                for sql in sqls:
                    result = public.ExecShell('mysql -u%s -p"%s" %s < %s 2>&1' % (db_user, db_pass, db_name, os.path.join(mig_dir, sql)))
                    if result[1] and result[1].strip():
                        public.WriteLog('qingka_manager', '迁移 %s 失败: %s' % (sql, result[1].strip()[:200]))
            # 7. 确保管理员账号和基础配置存在
            seed_sql = (
                "INSERT IGNORE INTO qingka_wangke_user (uid, uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth) "
                "VALUES (1, 1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');"
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
            # 11. 启动服务
            self.start(type('Args', (), {})())

            return public.returnMsg(True, '安装成功！域名: %s' % domain)
        except Exception as e:
            return public.returnMsg(False, '安装失败: %s' % str(e))

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
            # 清理站点目录（remove_domain 已删除宝塔DB和nginx配置）
            site_root = '/www/wwwroot/' + domain if domain else None
            dirs_to_clean = [self.__go_dir]
            if site_root and os.path.isdir(site_root):
                dirs_to_clean.append(site_root)
            for d in dirs_to_clean:
                if os.path.isdir(d):
                    public.ExecShell('rm -rf %s' % d)

            # 删除数据库
            if remove_data and db_info:
                public.ExecShell('mysql -u%s -p"%s" -e "DROP DATABASE IF EXISTS \\`%s\\`;" 2>/dev/null' % (db_info['user'], db_info['pass'], db_info['name']))
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

    def _enable_ssl_nginx(self, domain, cert_dir):
        try:
            import panelSite
            sargs = type('Args', (), {'siteName': domain, 'first_domain': domain})()
            panelSite.panelSite().SetSSLConf(sargs)
        except Exception:
            conf_file = '/www/server/panel/vhost/nginx/%s.conf' % domain
            if not os.path.isfile(conf_file): return
            content = public.readFile(conf_file)
            if 'ssl_certificate' in content: return
            ssl_str = """    ssl_certificate    /www/server/panel/vhost/cert/%s/fullchain.pem;
    ssl_certificate_key    /www/server/panel/vhost/cert/%s/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers EECDH+CHACHA20:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:!MD5;
    ssl_prefer_server_ciphers on;""" % (domain, domain)
            content = content.replace('#error_page 404/404.html;', '#error_page 404/404.html;\n' + ssl_str)
            content = content.replace('listen 80;', 'listen 80;\n    listen 443 ssl http2;')
            public.writeFile(conf_file, content)
        public.serviceReload()
    def _exec_sql(self, db_user, db_pass, db_name, sql):
        import tempfile
        fd, tmp = tempfile.mkstemp(suffix='.sql')
        try:
            os.write(fd, sql.encode('utf-8'))
            os.close(fd)
            return public.ExecShell('mysql -u%s -p"%s" %s < %s 2>&1' % (db_user, db_pass, db_name, tmp))
        finally:
            if os.path.isfile(tmp): os.remove(tmp)
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
            # 恢复用户配置文件
            if config_backup is not None:
                public.writeFile(self.__config_file, config_backup)
        finally:
            if os.path.isfile(tmp): os.remove(tmp)

    def _ensure_config(self):
        """首次安装时自动生成默认config.yaml，已有则不动"""
        if os.path.isfile(self.__config_file):
            return
        os.makedirs(os.path.dirname(self.__config_file), exist_ok=True)
        jwt_secret = hashlib.sha256(os.urandom(32)).hexdigest()[:32]
        config = '''server:
  port: 8080
  mode: release
  php_backend: ""
  php_public_url: ""
  bridge_secret: "%s"

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""
  dbname: "qingka"
  max_open_conns: 50
  max_idle_conns: 25

redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "%s"
  access_ttl: 604800
  refresh_ttl: 2592000

cache:
  order_list_ttl: 30
  class_list_ttl: 300

smtp:
  host: "smtp.qq.com"
  port: 465
  user: ""
  password: ""
  from_name: "系统通知"
  encryption: "ssl"
''' % (hashlib.md5(os.urandom(16)).hexdigest(), jwt_secret)
        public.writeFile(self.__config_file, config)
        public.WriteLog('qingka_manager', '已自动生成默认配置文件 config.yaml')

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

    def _run_migrations(self):
        db_info = self._read_db_config()
        if not db_info:
            return
        mig_dir = os.path.join(self.__go_dir, 'migrations')
        if not os.path.isdir(mig_dir):
            return
        sqls = sorted([f for f in os.listdir(mig_dir) if f.endswith('.sql')])
        for sql in sqls:
            sql_path = os.path.join(mig_dir, sql)
            try:
                result = public.ExecShell('mysql --force -u%s -p"%s" %s < %s 2>&1' % (db_info['user'], db_info['pass'], db_info['name'], sql_path))
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

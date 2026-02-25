#!/usr/bin/env python3
# coding: utf-8
"""授权码管理工具 - 在 29.colnt.com 服务器上运行
用法:
  python3 license_manager.py add <域名> [备注]    # 生成授权码
  python3 license_manager.py list                  # 列出所有授权
  python3 license_manager.py revoke <授权码>       # 吊销授权
  python3 license_manager.py unbind <授权码>       # 解绑设备
"""
import sys, os, json, time, secrets

LICENSE_FILE = '/www/wwwroot/29.colnt.com/licenses_db.json'

def load_db():
    if os.path.isfile(LICENSE_FILE):
        return json.loads(open(LICENSE_FILE).read())
    return {}

def save_db(db):
    open(LICENSE_FILE, 'w').write(json.dumps(db, indent=2, ensure_ascii=False))

def gen_key():
    return 'QK-' + secrets.token_hex(16).upper()

def cmd_add(domain, note=''):
    db = load_db()
    key = gen_key()
    db[key] = {'domain': domain, 'note': note, 'created': time.strftime('%Y-%m-%d %H:%M'), 'active': True}
    save_db(db)
    print('授权码: %s' % key)
    print('域名: %s' % domain)

def cmd_list():
    db = load_db()
    if not db:
        print('暂无授权')
        return
    for k, v in db.items():
        status = '✅' if v.get('active') else '❌'
        mid = v.get('machine_id', '')
        bound = ' [已绑定:%s]' % mid[:8] if mid else ' [未绑定]'
        print('%s %s → %s (%s) %s%s' % (status, k, v['domain'], v.get('note', ''), v['created'], bound))

def cmd_revoke(key):
    db = load_db()
    if key not in db:
        print('授权码不存在')
        return
    db[key]['active'] = False
    save_db(db)
    print('已吊销: %s' % key)

def cmd_unbind(key):
    db = load_db()
    if key not in db:
        print('授权码不存在')
        return
    if not db[key].get('machine_id'):
        print('该授权码未绑定设备')
        return
    del db[key]['machine_id']
    save_db(db)
    print('已解绑: %s' % key)

if __name__ == '__main__':
    args = sys.argv[1:]
    if not args:
        print(__doc__)
        sys.exit(0)
    cmd = args[0]
    if cmd == 'add' and len(args) >= 2:
        cmd_add(args[1], args[2] if len(args) > 2 else '')
    elif cmd == 'list':
        cmd_list()
    elif cmd == 'revoke' and len(args) >= 2:
        cmd_revoke(args[1])
    elif cmd == 'unbind' and len(args) >= 2:
        cmd_unbind(args[1])
    else:
        print(__doc__)
